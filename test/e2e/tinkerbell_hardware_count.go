package e2e

import (
	_ "embed"
	"fmt"

	"sigs.k8s.io/yaml"
)

//go:embed TINKERBELL_HARDWARE_COUNT.yaml
var tinkerbellHardwareCountFile []byte

// GetTinkerbellTestsHardwareRequirements returns a map of Tinkerbell test name to required hardware.
func GetTinkerbellTestsHardwareRequirements() (map[string]int, error) {
	var tinkerbellTestsHardwareRequirements map[string]int
	err := yaml.Unmarshal(tinkerbellHardwareCountFile, &tinkerbellTestsHardwareRequirements)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal hardware yaml: %v", err)
	}
	return tinkerbellTestsHardwareRequirements, nil
}
