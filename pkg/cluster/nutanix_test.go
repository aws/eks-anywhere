package cluster

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
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

func TestNutanixConfigClientBuilder(t *testing.T) {
	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpec), clusterConf)
	require.NoError(t, err)

	expectedDCConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), expectedDCConf)
	require.NoError(t, err)

	expectedMachineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), expectedMachineConf)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	m := mocks.NewMockClient(ctrl)

	m.EXPECT().Get(gomock.Any(), clusterConf.Spec.DatacenterRef.Name, gomock.Any(), &anywherev1.NutanixDatacenterConfig{}).
		DoAndReturn(func(ctx context.Context, name, namespace string, obj client.Object) error {
			expectedDCConf.DeepCopyInto(obj.(*anywherev1.NutanixDatacenterConfig))
			return nil
		})

	m.EXPECT().Get(gomock.Any(), clusterConf.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, gomock.Any(), &anywherev1.NutanixMachineConfig{}).
		DoAndReturn(func(ctx context.Context, name, namespace string, obj client.Object) error {
			expectedMachineConf.DeepCopyInto(obj.(*anywherev1.NutanixMachineConfig))
			return nil
		})

	ccb := NewConfigClientBuilder().Register(
		getNutanixDatacenter,
		getNutanixMachineConfigs,
	)
	conf, err := ccb.Build(context.TODO(), m, clusterConf)
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, expectedDCConf, conf.NutanixDatacenter)
	assert.Equal(t, expectedMachineConf, conf.NutanixMachineConfig(clusterConf.Spec.ControlPlaneConfiguration.MachineGroupRef.Name))
}
