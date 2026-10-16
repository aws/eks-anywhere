package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tinkv1alpha1 "github.com/tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

type generateHardwareTestCase struct {
	name       string
	ip, gw, nm string
	nameserver string
	slaac      bool
	wantErr    string
}

const previousHardwareOutput = "existing hardware output"

func TestGenerateHardwareIPv6SLAAC(t *testing.T) {
	for _, tt := range []generateHardwareTestCase{
		{name: "default static IPv4", ip: "192.0.2.100", gw: "192.0.2.1", nm: "255.255.255.0", nameserver: "192.0.2.53"},
		{name: "default rejects addressless", nameserver: "2001:db8::53", wantErr: "IPAddress is empty"},
		{name: "explicit SLAAC", nameserver: "2001:db8::53", slaac: true},
		{name: "SLAAC rejects static IPv6", ip: "2001:db8::100", gw: "2001:db8::1", nm: "64", nameserver: "2001:db8::53", slaac: true, wantErr: "ip_address must be empty for IPv6 SLAAC"},
		{name: "SLAAC rejects static IPv4", ip: "192.0.2.100", gw: "192.0.2.1", nm: "255.255.255.0", nameserver: "2001:db8::53", slaac: true, wantErr: "ip_address must be empty for IPv6 SLAAC"},
		{name: "SLAAC rejects partial gateway", gw: "2001:db8::1", nameserver: "2001:db8::53", slaac: true, wantErr: "gateway must be empty for IPv6 SLAAC"},
		{name: "SLAAC retains DNS validation", nameserver: "invalid", slaac: true, wantErr: "Nameservers:"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runHardwareGeneration(t, tt)
			if tt.wantErr != "" {
				assertRejectedHardwareGeneration(t, result, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assertGeneratedHardware(t, result, tt)
		})
	}
}

func restoreHardwareGenerationOptions(t *testing.T) {
	t.Helper()
	savedOptions := *hOpts
	t.Cleanup(func() {
		for _, name := range []string{"hardware-csv", "output", "ipv6-slaac"} {
			flag := generateHardwareCmd.Flags().Lookup(name)
			if err := flag.Value.Set(flag.DefValue); err != nil {
				t.Error(err)
			}
			flag.Changed = false
		}
		*hOpts = savedOptions
	})
}

func runHardwareGeneration(t *testing.T, tt generateHardwareTestCase) ([]byte, error) {
	t.Helper()
	restoreHardwareGenerationOptions(t)
	dir := t.TempDir()
	csvPath, outputPath := filepath.Join(dir, "hardware.csv"), filepath.Join(dir, "hardware.yaml")
	content := "hostname,ip_address,netmask,gateway,nameservers,mac,disk,labels\n" +
		"node-example," + tt.ip + "," + tt.nm + "," + tt.gw + "," + tt.nameserver + ",02:00:00:00:00:01,/dev/sda,type=cp\n"
	if err := os.WriteFile(csvPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte(previousHardwareOutput), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--hardware-csv", csvPath, "--output", outputPath}
	if tt.slaac {
		args = append(args, "--ipv6-slaac")
	}
	if err := generateHardwareCmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	err := generateHardwareCmd.RunE(generateHardwareCmd, nil)
	result, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return result, err
}

func assertRejectedHardwareGeneration(t *testing.T, result []byte, err error, wantErr string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("expected error containing %q, got %v", wantErr, err)
	}
	if string(result) != previousHardwareOutput {
		t.Fatal("invalid CSV changed the existing output")
	}
}

func assertGeneratedHardware(t *testing.T, result []byte, tt generateHardwareTestCase) {
	t.Helper()
	catalogue := hardware.NewCatalogue()
	if err := hardware.ParseYAMLCatalogue(catalogue, strings.NewReader(string(result))); err != nil {
		t.Fatal(err)
	}
	if len(catalogue.AllHardware()) != 1 {
		t.Fatalf("expected one Hardware, got %d", len(catalogue.AllHardware()))
	}
	hw := catalogue.AllHardware()[0]
	if hw.Name != "node-example" || hw.Spec.Interfaces[0].DHCP.MAC != "02:00:00:00:00:01" {
		t.Fatal("generated hardware lost its identity")
	}
	assertGeneratedHardwareAddresses(t, hw, tt)
}

func assertGeneratedHardwareAddresses(t *testing.T, hw *tinkv1alpha1.Hardware, tt generateHardwareTestCase) {
	t.Helper()
	if tt.slaac {
		if hw.Spec.Interfaces[0].DHCP.IP != nil || len(hw.Spec.Metadata.Instance.Ips) != 0 {
			t.Fatal("SLAAC Hardware contains static addresses")
		}
	} else if hw.Spec.Interfaces[0].DHCP.IP == nil || hw.Spec.Interfaces[0].DHCP.IP.Address != tt.ip || hw.Spec.Interfaces[0].DHCP.IP.Family != 4 {
		t.Fatal("default generation changed the static IPv4 address")
	}
}

func TestGenerateHardwareIPv6SLAACFlagDefault(t *testing.T) {
	flag := generateHardwareCmd.Flags().Lookup("ipv6-slaac")
	if flag == nil || flag.DefValue != "false" || flag.Value.Type() != "bool" {
		t.Fatal("ipv6-slaac must be an optional bool flag defaulting to false")
	}
}
