package v1alpha1_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// MockReaderWithError is a mock implementation of the Reader interface that returns an error.
type MockReaderWithError struct{}

// Read returns an error.
func (m *MockReaderWithError) Read() ([]byte, error) {
	return nil, errors.New("custom error")
}

func TestGetTinkerbellMachineConfigs(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		expectedError error
	}{
		{
			name:          "non-splitable manifest",
			fileName:      "testdata/invalid_manifest.yaml",
			expectedError: errors.New("invalid Yaml document separator: \\nkey: value\\ninvalid_separator\\n"),
		},
		{
			name:          "non existent file",
			fileName:      "testdata/non_existent_file.yaml",
			expectedError: errors.New("unable to read file due to: open testdata/non_existent_file.yaml: no such file or directory"),
		},
		{
			name:          "non existent file",
			fileName:      "testdata/cluster_vsphere.yaml",
			expectedError: errors.New("unable to find kind TinkerbellMachineConfig in file"),
		},
		{
			name:          "duplicate fields in machine config",
			fileName:      "testdata/tinkerbell_cluster_with_duplicate_mchine_config_fields.yaml",
			expectedError: errors.New("unable to unmarshall content from file due to: error converting YAML to JSON: yaml: unmarshal errors:\n  line 5: key \"name\" already set in map"),
		},
		{
			name:          "valid tinkerbell manifest",
			fileName:      "testdata/tinkerbell_cluster_without_worker_nodes.yaml",
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configs, err := v1alpha1.GetTinkerbellMachineConfigs(test.fileName)

			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, len(configs) > 0)
			}
		})
	}
}
