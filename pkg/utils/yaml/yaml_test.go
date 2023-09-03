package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/utils/file"
	yamlutil "github.com/aws/eks-anywhere/pkg/utils/yaml"
)

func TestParseMultiYamlFile(t *testing.T) {
	fileName := "testdata/multi_resource_manifests.yaml"

	expectedYamls := []string{
		`
    apiVersion: v1
    kind: Cluster
    metadata:
      name: cluster-1
    spec:
      machineConfigRef:
        name: test-machine-config
      templateConfigRef:
        name: test-template-config`,

		`
    apiVersion: v1
    kind: MachineConfigRef
    metadata:
      name: test-machine-config
    spec:
      machine:
        count: 1`,

		`
    apiVersion: v1
    kind: TemplateConfigRef
    metadata:
      name: test-template-config
    spec:
      template:
        name: test-template`,
	}

	var expectedData interface{}
	var actualData interface{}

	t.Run("split yaml resources", func(t *testing.T) {
		r, err := file.ReadFile(fileName)
		require.NoError(t, err)

		got, err := yamlutil.SplitDocuments(r)
		require.NoError(t, err)
		require.Equal(t, len(expectedYamls), len(got))

		for i := 0; i < len(expectedYamls); i++ {
			err = yaml.Unmarshal([]byte(expectedYamls[i]), &expectedData)
			require.NoError(t, err)

			err = yaml.Unmarshal(got[i], &actualData)
			require.NoError(t, err)

			assert.Equal(t, expectedData, actualData)
		}
	})
}
