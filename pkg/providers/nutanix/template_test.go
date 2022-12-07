package nutanix

import (
	_ "embed"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/version"
)

//go:embed testdata/eksa-cluster.yaml
var nutanixClusterConfigSpec string

//go:embed testdata/datacenterConfig.yaml
var nutanixDatacenterConfigSpec string

//go:embed testdata/machineConfig.yaml
var nutanixMachineConfigSpec string

//go:embed testdata/eksa-cluster-autoscaler.yaml
var nutanixClusterConfigSpecWithAutoscaler string

func fakemarshal(v interface{}) ([]byte, error) {
	return []byte{}, errors.New("marshalling failed")
}

func restoremarshal(replace func(v interface{}) ([]byte, error)) {
	jsonMarshal = replace
}

func TestNewNutanixTemplateBuilder(t *testing.T) {
	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpec), clusterConf)
	require.NoError(t, err)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConf)
	require.NoError(t, err)

	workerConfs := map[string]anywherev1.NutanixMachineConfigSpec{
		"eksa-unit-test": machineConf.Spec,
	}

	t.Setenv(constants.NutanixUsernameKey, "admin")
	t.Setenv(constants.NutanixPasswordKey, "password")
	creds := GetCredsFromEnv()
	builder := NewNutanixTemplateBuilder(&dcConf.Spec, &machineConf.Spec, &machineConf.Spec, workerConfs, creds, time.Now)
	assert.NotNil(t, builder)

	v := version.Info{GitVersion: "v0.0.1"}
	buildSpec, err := cluster.NewSpecFromClusterConfig("testdata/eksa-cluster.yaml", v, cluster.WithReleasesManifest("testdata/simple_release.yaml"))
	assert.NoError(t, err)

	cpSpec, err := builder.GenerateCAPISpecControlPlane(buildSpec)
	assert.NoError(t, err)
	assert.NotNil(t, cpSpec)

	workloadTemplateNames := map[string]string{
		"eksa-unit-test": "eksa-unit-test",
	}
	kubeadmconfigTemplateNames := map[string]string{
		"eksa-unit-test": "eksa-unit-test",
	}
	workerSpec, err := builder.GenerateCAPISpecWorkers(buildSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	assert.NoError(t, err)
	assert.NotNil(t, workerSpec)

	secretSpec, err := builder.GenerateCAPISpecSecret(buildSpec)
	assert.NoError(t, err)
	assert.NotNil(t, secretSpec)
	expectedSecret, err := os.ReadFile("testdata/templated_secret.yaml")
	require.NoError(t, err)
	assert.Equal(t, expectedSecret, secretSpec)
}

func TestNewNutanixTemplateBuilderGenerateCAPISpecSecret(t *testing.T) {
	storedMarshal := jsonMarshal
	jsonMarshal = fakemarshal
	defer restoremarshal(storedMarshal)

	t.Setenv(constants.NutanixUsernameKey, "admin")
	t.Setenv(constants.NutanixPasswordKey, "password")
	creds := GetCredsFromEnv()
	builder := NewNutanixTemplateBuilder(nil, nil, nil, nil, creds, time.Now)
	assert.NotNil(t, builder)

	v := version.Info{GitVersion: "v0.0.1"}
	buildSpec, err := cluster.NewSpecFromClusterConfig("testdata/eksa-cluster.yaml", v, cluster.WithReleasesManifest("testdata/simple_release.yaml"))
	assert.NoError(t, err)

	secretSpec, err := builder.GenerateCAPISpecSecret(buildSpec)
	assert.Nil(t, secretSpec)
	assert.Error(t, err)
}

func TestNutanixTemplateBuilderGenerateCAPISpecForCreateWithAutoscalingConfiguration(t *testing.T) {
	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpecWithAutoscaler), clusterConf)
	require.NoError(t, err)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConf)
	require.NoError(t, err)

	workerConfs := map[string]anywherev1.NutanixMachineConfigSpec{
		"eksa-unit-test": machineConf.Spec,
	}

	t.Setenv(constants.NutanixUsernameKey, "admin")
	t.Setenv(constants.NutanixPasswordKey, "password")
	creds := GetCredsFromEnv()

	builder := NewNutanixTemplateBuilder(&dcConf.Spec, &machineConf.Spec, &machineConf.Spec, workerConfs, creds, time.Now)
	assert.NotNil(t, builder)

	v := version.Info{GitVersion: "v0.0.1"}
	buildSpec, err := cluster.NewSpecFromClusterConfig("testdata/eksa-cluster-autoscaler.yaml", v, cluster.WithReleasesManifest("testdata/simple_release.yaml"))
	assert.NoError(t, err)

	workloadTemplateNames := map[string]string{
		"eksa-unit-test": "eksa-unit-test",
	}
	kubeadmconfigTemplateNames := map[string]string{
		"eksa-unit-test": "eksa-unit-test",
	}
	workerSpec, err := builder.GenerateCAPISpecWorkers(buildSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	assert.NoError(t, err)
	expectedWorkerSpec, err := os.ReadFile("testdata/expected_results_autoscaling_md.yaml")
	require.NoError(t, err)
	assert.Equal(t, workerSpec, expectedWorkerSpec)
}
