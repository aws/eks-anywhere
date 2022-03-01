package hardware_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestNewYamlParserSuccess(t *testing.T) {
	yaml, err := hardware.NewYamlParser()
	if err != nil {
		t.Fatalf("hardware.NewYamlParser() error = %v, expected = nil", err)
	}

	defer func() {
		yaml.Close()
		_ = os.RemoveAll("hardware-manifests")
	}()

	if err := yaml.WriteHardwareYaml("b14d7f5b-8903-4a4c-b38d-55889ba820ba", "worker1", "192.168.0.10", "supermicro", "admin", "Admin"); err != nil {
		t.Fatalf("hardware.NewYamlParser().WriteHardwareYaml() error = %v, expected = nil", err)
	}

	test.AssertFilesEquals(t, "hardware-manifests/hardware.yaml", "testdata/expected_hardware_yaml.yaml")
}
