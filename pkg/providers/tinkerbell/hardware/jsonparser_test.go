package hardware_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestNewJsonParserSuccess(t *testing.T) {
	expected_hardware_json := "testdata/expected_hardware_json.json"
	json, err := hardware.NewJsonParser()
	if err != nil {
		t.Fatalf("hardware.NewJsonParser() error = %v, expected = nil", err)
	}

	defer json.CleanUp()

	hardware, err := json.GetHardwareJson("b14d7f5b-8903-4a4c-b38d-55889ba820ba", "worker1", "192.168.0.10", "192.168.0.1", "255.255.255.0", "00:00:00:00:00:00", "8.8.8.8|8.8.4.4")
	if err != nil {
		t.Fatalf("hardware.NewJsonParser().GetHardwareJson() error = %v, expected = nil", err)
	}

	test.AssertContentToFile(t, string(hardware), expected_hardware_json)
}
