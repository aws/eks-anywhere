package tinkerbell_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	networkutilsmocks "github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	tinkerbellmocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
)

func TestValidateMinimumRequiredTinkerbellHardwareAvailable_SufficientHardware(t *testing.T) {
	for name, tt := range map[string]struct {
		AvailableHardware int
		ClusterSpec       v1alpha1.ClusterSpec
	}{
		"SufficientHardware":  {3, newValidClusterSpec(1, 1, 1)},
		"SuperfluousHardware": {5, newValidClusterSpec(1, 1, 1)},
	} {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hardwareConfig := newHardwareConfigWithHardware(tt.AvailableHardware)

			var validator tinkerbell.Validator
			tinkerbell.SetValidatorHardwareConfig(&validator, hardwareConfig)

			assert.NoError(t, validator.ValidateMinimumRequiredTinkerbellHardwareAvailable(tt.ClusterSpec))
		})
	}
}

func TestValidateMinimumRequiredTinkerbellHardwareAvailable_InsufficientHardware(t *testing.T) {
	clusterSpec := newValidClusterSpec(1, 1, 1)

	hardwareConfig := newHardwareConfigWithHardware(2)

	var validator tinkerbell.Validator
	tinkerbell.SetValidatorHardwareConfig(&validator, hardwareConfig)

	assert.Error(t, validator.ValidateMinimumRequiredTinkerbellHardwareAvailable(clusterSpec))
}

func TestValidateMinimumRequiredTinkerbellHardware_EtcdUnspecified(t *testing.T) {
	clusterSpec := newValidClusterSpec(1, 0, 1)
	clusterSpec.ExternalEtcdConfiguration = nil

	hardwareConfig := newHardwareConfigWithHardware(3)

	var validator tinkerbell.Validator
	tinkerbell.SetValidatorHardwareConfig(&validator, hardwareConfig)

	assert.NoError(t, validator.ValidateMinimumRequiredTinkerbellHardwareAvailable(clusterSpec))
}

func TestValidateTinkerbellConfig_ValidAuthorities(t *testing.T) {
	ctrl := gomock.NewController(t)
	pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
	tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
	net := networkutilsmocks.NewMockNetClient(ctrl)

	tink.EXPECT().GetHardware(gomock.Any())

	var hardware hardware.HardwareConfig
	datacenter := newValidTinkerbellDatacenterConfig()

	validator := tinkerbell.NewValidator(tink, net, hardware, pbnj)
	err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

	assert.NoError(t, err)
}

func TestValidateTinkerbellConfig_InvalidGrpcAuthority(t *testing.T) {
	cases := map[string]string{
		"Missing port":     "1.1.1.1",
		"Missing hostname": ":44",
		"Port is alpha":    "1.1.1.1:foo",
		"Port too large":   "1.1.1.1:99999",
	}

	for name, address := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
			tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
			net := networkutilsmocks.NewMockNetClient(ctrl)

			tink.EXPECT().GetHardware(gomock.Any())

			var hardware hardware.HardwareConfig
			datacenter := newValidTinkerbellDatacenterConfig()
			datacenter.Spec.TinkerbellGRPCAuth = address

			validator := tinkerbell.NewValidator(tink, net, hardware, pbnj)
			err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

			assert.Error(t, err)
		})
	}
}

func TestValidateTinkerbellConfig_InvalidPbnjAuthority(t *testing.T) {
	cases := map[string]string{
		"Missing port":     "1.1.1.1",
		"Missing hostname": ":44",
		"Port is alpha":    "1.1.1.1:foo",
		"Port too large":   "1.1.1.1:99999",
	}

	for name, address := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			pbnj := tinkerbellmocks.NewMockProviderPbnjClient(ctrl)
			tink := tinkerbellmocks.NewMockProviderTinkClient(ctrl)
			net := networkutilsmocks.NewMockNetClient(ctrl)

			tink.EXPECT().GetHardware(gomock.Any())

			var hardware hardware.HardwareConfig
			datacenter := newValidTinkerbellDatacenterConfig()
			datacenter.Spec.TinkerbellPBnJGRPCAuth = address

			validator := tinkerbell.NewValidator(tink, net, hardware, pbnj)
			err := validator.ValidateTinkerbellConfig(context.Background(), datacenter)

			assert.Error(t, err)
		})
	}
}

func newValidClusterSpec(cp, etcd, worker int) v1alpha1.ClusterSpec {
	return v1alpha1.ClusterSpec{
		ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
			Count: cp,
		},
		ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
			Count: etcd,
		},
		WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
			{Count: worker},
		},
	}
}

func newHardwareConfigWithHardware(hardwareCount int) hardware.HardwareConfig {
	return hardware.HardwareConfig{
		Hardwares: make([]tinkv1alpha1.Hardware, hardwareCount),
	}
}

func newValidTinkerbellDatacenterConfig() *v1alpha1.TinkerbellDatacenterConfig {
	return &v1alpha1.TinkerbellDatacenterConfig{
		Status: v1alpha1.TinkerbellDatacenterConfigStatus{},
		Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP:           "1.1.1.1",
			TinkerbellCertURL:      "http://domain.com/path",
			TinkerbellGRPCAuth:     "1.1.1.1:444",
			TinkerbellPBnJGRPCAuth: "1.1.1.1:444",
		},
	}
}
