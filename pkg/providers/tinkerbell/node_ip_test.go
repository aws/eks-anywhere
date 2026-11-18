package tinkerbell

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestIPv6NodeIPBootstrap(t *testing.T) {
	for _, osFamily := range []v1alpha1.OSFamily{v1alpha1.Ubuntu, v1alpha1.RedHat} {
		for _, isoBoot := range []bool{false, true} {
			name := string(osFamily) + "/netboot"
			if isoBoot {
				name = string(osFamily) + "/iso"
			}
			t.Run(name, func(t *testing.T) {
				g := NewWithT(t)
				spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
				spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2001:db8::10"
				spec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"2001:db8:100::/56"}
				spec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"2001:db8:200::/112"}
				spec.TinkerbellDatacenter.Spec.TinkerbellIP = "2001:db8::11"
				spec.TinkerbellDatacenter.Spec.DHCPv6Config = &v1alpha1.DHCPv6Config{
					DerivedAddressPool: "2001:db8::/64",
				}
				if isoBoot {
					spec.TinkerbellDatacenter.Spec.IsoBoot = true
					spec.TinkerbellDatacenter.Spec.DHCPv6Config = nil
				}
				for _, machine := range spec.TinkerbellMachineConfigs {
					machine.Spec.OSFamily = osFamily
				}

				cp, err := ControlPlaneSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.Files).To(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
				command := ipv6NodeIPScriptPath + " 2001:db8::11 2001:db8::10 || exit 1"
				g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.PreKubeadmCommands[0]).To(Equal(command))

				workers, err := WorkersSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
				g.Expect(err).NotTo(HaveOccurred())
				for _, worker := range workers.Groups {
					g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.Files).To(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
					g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.PreKubeadmCommands[0]).To(Equal(command))
				}
			})
		}
	}
}

func TestIPv6NodeIPHelperScript(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is needed to run the IPv6 node-IP helper tests")
	}
	path := filepath.Join(t.TempDir(), "ipv6-node-ip.py")
	g := NewWithT(t)
	g.Expect(os.WriteFile(path, []byte(ipv6NodeIPScript), 0o600)).To(Succeed())
	cmd := exec.Command(python, "-B", "config/ipv6-node-ip_test.py")
	cmd.Env = append(os.Environ(), "EKSA_NODE_IP_HELPER="+path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("IPv6 node-IP helper tests: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}

func TestIPv6NodeIPBootstrapActivation(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		dhcpv6     bool
		osFamily   v1alpha1.OSFamily
		wantHelper bool
		wantErr    string
	}{
		{"IPv6 without DHCPv6", "2001:db8::11", false, v1alpha1.Ubuntu, true, ""},
		{"IPv6 with DHCPv6", "2001:db8::11", true, v1alpha1.RedHat, true, ""},
		{"IPv4", "192.0.2.11", false, v1alpha1.Ubuntu, false, ""},
		{"mapped IPv4", "::ffff:192.0.2.11", false, v1alpha1.Ubuntu, false, ""},
		{"Bottlerocket", "2001:db8::11", true, v1alpha1.Bottlerocket, false, ""},
		{"invalid address without DHCPv6", "invalid", false, v1alpha1.Ubuntu, false, ""},
		{"invalid address with DHCPv6", "invalid", true, v1alpha1.Ubuntu, false, `invalid Tinkerbell IPv6 address "invalid"`},
		{"IPv4 with DHCPv6", "192.0.2.11", true, v1alpha1.Ubuntu, false, `invalid Tinkerbell IPv6 address "192.0.2.11"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
			spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2001:db8::10"
			dc := v1alpha1.TinkerbellDatacenterConfigSpec{TinkerbellIP: tt.address}
			if tt.dhcpv6 {
				dc.DHCPv6Config = &v1alpha1.DHCPv6Config{DerivedAddressPool: "2001:db8::/64"}
			}
			values := map[string]interface{}{}
			err := addIPv6NodeIPBootstrap(values, spec, dc, tt.osFamily)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(tt.wantErr))
			} else {
				g.Expect(err).To(Succeed())
			}
			_, present := values["ipv6NodeIPScript"]
			g.Expect(present).To(Equal(tt.wantHelper))
		})
	}
}

func TestIPv6NodeIPHelperFailureStopsBootstrap(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is needed to check bootstrap command failure")
	}
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2001:db8::10"
	values := map[string]interface{}{}
	g.Expect(addIPv6NodeIPBootstrap(values, spec, v1alpha1.TinkerbellDatacenterConfigSpec{TinkerbellIP: "2001:db8::11"}, v1alpha1.Ubuntu)).To(Succeed())
	dir := t.TempDir()
	helper := filepath.Join(dir, "helper")
	g.Expect(os.WriteFile(helper, []byte("#!/bin/sh\nexit 1\n"), 0o700)).To(Succeed())
	command := strings.Replace(values["ipv6NodeIPCommand"].(string), ipv6NodeIPScriptPath, helper, 1)
	script := filepath.Join(dir, "runcmd")
	g.Expect(os.WriteFile(script, []byte("#!/bin/sh\n"+command+"\necho kubeadm-reached > kubeadm-called\n"), 0o700)).To(Succeed())
	cmd := exec.Command(shell, script)
	cmd.Dir = dir
	_, err = cmd.CombinedOutput()
	g.Expect(err).To(HaveOccurred())
	_, err = os.Stat(filepath.Join(dir, "kubeadm-called"))
	g.Expect(os.IsNotExist(err)).To(BeTrue())
}

func TestIPv4BootstrapDoesNotAddIPv6NodeIPHelper(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	cp, err := ControlPlaneSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.Files).NotTo(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
	workers, err := WorkersSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
	g.Expect(err).NotTo(HaveOccurred())
	for _, worker := range workers.Groups {
		g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.Files).NotTo(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
	}
}
