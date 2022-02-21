package tinkerbell_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
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
