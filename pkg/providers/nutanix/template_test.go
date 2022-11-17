package nutanix

import (
	_ "embed"
	"errors"
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

//go:embed testdata/datacenterConfig_with_trust_bundle.yaml
var nutanixDatacenterConfigSpecWithTrustBundle string

//go:embed testdata/eksa-cluster.json
var nutanixClusterConfigSpecJSON string

//go:embed testdata/machineConfig.yaml
var nutanixMachineConfigSpec string

//go:embed testdata/datacenterConfig.json
var nutanixDatacenterConfigSpecJSON string

//go:embed testdata/machineConfig.json
var nutanixMachineConfigSpecJSON string

//go:embed testdata/machineDeployment.json
var nutanixMachineDeploymentSpecJSON string

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
}
