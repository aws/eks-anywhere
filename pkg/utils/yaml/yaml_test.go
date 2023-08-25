package yaml_test

import (
	"testing"

	yamlutil "github.com/aws/eks-anywhere/pkg/utils/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type TestGenericYaml struct {
	ApiVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	MetaData   TestMetaData `yaml:"metadata"`
	Spec       interface{}  `yaml:"spec"`
}

type TestMetaData struct {
	Name string `yaml:"name"`
}

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
		got, err := yamlutil.ParseMultiYamlFile(fileName)
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
