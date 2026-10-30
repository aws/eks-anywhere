package cmd

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type templateConfigAddressTestCase struct {
	name, destination, override, source, network, wantErr string
}

func TestGenerateTinkerbellTemplateConfigBootstrapAddress(t *testing.T) {
	for _, tt := range []templateConfigAddressTestCase{
		{name: "IPv4 route", destination: "192.0.2.10", source: "192.0.2.20", network: "udp4"},
		{name: "IPv6 route", destination: "2001:db8::10", source: "2001:db8::20", network: "udp6"},
		{name: "IPv4 override", destination: "192.0.2.10", override: "192.0.2.30", source: "192.0.2.30"},
		{name: "IPv6 override", destination: "2001:db8::10", override: "2001:db8::30", source: "2001:db8::30"},
		{name: "wrong family override", destination: "2001:db8::10", override: "192.0.2.30", wantErr: "must use the same IP family"},
		{name: "invalid override", destination: "2001:db8::10", override: "invalid", wantErr: "invalid bootstrap IP"},
		{name: "wrong family source", destination: "2001:db8::10", source: "192.0.2.20", network: "udp6", wantErr: "must use the same IP family"},
		{name: "route error", destination: "2001:db8::10", network: "udp6", wantErr: "no route"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runTemplateConfigAddressCase(t, tt)
		})
	}
}

func runTemplateConfigAddressCase(t *testing.T, tt templateConfigAddressTestCase) {
	t.Helper()
	dialed := false
	conn := &templateConfigRouteConn{source: net.ParseIP(tt.source)}
	client := templateConfigAddressClient(t, tt, conn, &dialed)
	cfg, err := generateTinkerbellTemplateConfig(templateConfigSpec(tt.destination), tt.override, client)
	assertTemplateConfigGeneration(t, cfg, err, tt.wantErr, tt.source, tt.destination)
	if tt.override == "" && !dialed {
		t.Fatal("automatic selection did not perform a route lookup")
	}
	if tt.source != "" && dialed && !conn.closed {
		t.Fatal("route connection was not closed")
	}
}

func templateConfigAddressClient(t *testing.T, tt templateConfigAddressTestCase, conn *templateConfigRouteConn, dialed *bool) templateConfigNetClient {
	t.Helper()
	return func(network, address string, timeout time.Duration) (net.Conn, error) {
		*dialed = true
		if tt.override != "" {
			t.Fatal("explicit override performed a route lookup")
		}
		if network != tt.network || address != net.JoinHostPort(tt.destination, "80") || timeout != 5*time.Second {
			t.Fatalf("wrong route lookup: %s %s %s", network, address, timeout)
		}
		if tt.source == "" {
			return nil, errors.New("no route")
		}
		return conn, nil
	}
}

func assertTemplateConfigGeneration(t *testing.T, cfg *v1alpha1.TinkerbellTemplateConfig, err error, wantErr, source, destination string) {
	t.Helper()
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) || cfg != nil {
			t.Fatalf("expected no template and error containing %q, got template=%v error=%v", wantErr, cfg, err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	metadata := templateConfigMetadata(cfg)
	bootstrapURL := "http://" + net.JoinHostPort(source, "7172")
	workloadURL := "http://" + net.JoinHostPort(destination, "7172")
	if !strings.Contains(metadata, bootstrapURL) || !strings.Contains(metadata, workloadURL) {
		t.Fatalf("metadata does not contain the selected addresses: %s", metadata)
	}
	if strings.Index(metadata, bootstrapURL) >= strings.Index(metadata, workloadURL) {
		t.Fatal("bootstrap metadata URL must precede the workload URL")
	}
}

func templateConfigMetadata(cfg *v1alpha1.TinkerbellTemplateConfig) string {
	for _, action := range cfg.Spec.Template.Tasks[0].Actions {
		if action.Environment["DEST_PATH"] == "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg" {
			return action.Environment["CONTENTS"]
		}
	}
	return ""
}

func templateConfigSpec(tinkerbellIP string) *cluster.Spec {
	return &cluster.Spec{Config: &cluster.Config{
		Cluster: &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-example"},
			Spec: v1alpha1.ClusterSpec{ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{Name: "control-plane"},
			}},
		},
		TinkerbellDatacenter: &v1alpha1.TinkerbellDatacenterConfig{
			Spec: v1alpha1.TinkerbellDatacenterConfigSpec{TinkerbellIP: tinkerbellIP, OSImageURL: "http://example.invalid/os.img.gz"},
		},
		TinkerbellMachineConfigs: map[string]*v1alpha1.TinkerbellMachineConfig{
			"control-plane": {Spec: v1alpha1.TinkerbellMachineConfigSpec{OSFamily: v1alpha1.Ubuntu}},
		},
	}}
}

type templateConfigNetClient func(string, string, time.Duration) (net.Conn, error)

func (f templateConfigNetClient) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return f(network, address, timeout)
}

type templateConfigRouteConn struct {
	net.Conn
	source net.IP
	closed bool
}

func (c *templateConfigRouteConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: c.source}
}

func (c *templateConfigRouteConn) Close() error {
	c.closed = true
	return nil
}
