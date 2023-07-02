package e2e

import (
	_ "embed"
	"fmt"

	"sigs.k8s.io/yaml"
)

//go:embed tinkerbell_hardware_count.yaml
var tinkerbellHardwareCountFile string

// TinkerbellTest maps each Tinkbell test with the hardware count needed for the test.
type TinkerbellTest struct {
	Name  string `yaml:"name"`
	Count int    `yaml:"count"`
}

// GetTinkerbellHardwareMap returns a map of tink tests with associated hardware count.
func GetTinkerbellHardwareMap() (map[string]int, error) {
	tinkerbellHardwareMap, err := readTinkerbellTestFile()
	if err != nil {
		return nil, fmt.Errorf("error getting hardware count for tinkerbell tests: %v", err)
	}
	return tinkerbellHardwareMap, nil
}

func readTinkerbellTestFile() (map[string]int, error) {
	tinkerbellTests := make(map[string][]TinkerbellTest)
	err := yaml.Unmarshal([]byte(tinkerbellHardwareCountFile), &tinkerbellTests)
	if err != nil {
		return nil, fmt.Errorf("unable to Unmarshal %s file: %v", TinkerbellHardwareCountFile, err)
	}
	tinkerbellTestsHardwareCount, ok := tinkerbellTests[tinkerbellTestsHardwareCountIdentifier]
	if !ok {
		return nil, fmt.Errorf("error in type assertion: %v", tinkerbellTests[tinkerbellTestsHardwareCountIdentifier])
	}
	tinkerbellTestHardwareMap := make(map[string]int)
	for _, test := range tinkerbellTestsHardwareCount {
		tinkerbellTestHardwareMap[test.Name] = test.Count
	}
	return tinkerbellTestHardwareMap, nil
}
