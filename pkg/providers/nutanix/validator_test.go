package nutanix

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestNutanixValidator(t *testing.T) {
	ctrl := gomock.NewController(t)

	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err := yaml.Unmarshal([]byte(nutanixDatacenterConfigSpec), dcConf)
	require.NoError(t, err)

	machineConfig := &anywherev1.NutanixMachineConfig{}
	err = yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
	require.NoError(t, err)

	mockClient := NewMockclient(ctrl)
	mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no clusters found"))
	mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no subnets found"))
	mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("no images found"))
	validator, err := NewValidator(mockClient)
	assert.NoError(t, err)
	require.NotNil(t, validator)

	err = validator.ValidateMachineConfig(context.Background(), machineConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no clusters found")
	assert.Contains(t, err.Error(), "no subnets found")
	assert.Contains(t, err.Error(), "no images found")
}
