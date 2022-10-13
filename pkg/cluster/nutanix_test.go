package cluster

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

//go:embed testdata/nutanix/eksa-cluster.yaml
var nutanixClusterConfigSpec string

//go:embed testdata/nutanix/datacenterConfig.yaml
var nutanixDatacenterConfigSpec string

//go:embed testdata/nutanix/machineConfig.yaml
var nutanixMachineConfigSpec string

func TestValidateNutanixEntry(t *testing.T) {
	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpec), clusterConf)
	require.NoError(t, err)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConf)
	require.NoError(t, err)

	config := &Config{
		Cluster:           clusterConf,
		NutanixDatacenter: dcConf,
		NutanixMachineConfigs: map[string]*anywherev1.NutanixMachineConfig{
			"eksa-unit-test": machineConf,
		},
	}

	assert.Equal(t, config.NutanixMachineConfig("eksa-unit-test"), machineConf)

	cm, err := NewDefaultConfigManager()
	assert.NoError(t, err)

	c, err := cm.Parse([]byte(nutanixClusterConfigSpec))
	assert.NoError(t, err)
	fmt.Println(c)

	err = cm.Validate(config)
	assert.NoError(t, err)
}
