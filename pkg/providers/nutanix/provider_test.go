package nutanix

import (
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

func TestNutanixProvier(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	kubectl := executables.NewKubectl(executable)

	clusterConf := &anywherev1.Cluster{}
	err := yaml.Unmarshal([]byte(nutanixClusterConfigSpec), clusterConf)
	require.NoError(t, err)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err = yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConf := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConf)
	require.NoError(t, err)

	workerConfs := map[string]*anywherev1.NutanixMachineConfig{
		"eksa-unit-test": machineConf,
	}

	os.Setenv(nutanixUsernameKey, "admin")
	defer os.Unsetenv(nutanixUsernameKey)
	os.Setenv(nutanixPasswordKey, "password")
	defer os.Unsetenv(nutanixPasswordKey)

	provider, err := NewProvider(dcConf, workerConfs, clusterConf, kubectl, time.Now)
	require.NoError(t, err)
	assert.NotNil(t, provider)
}
