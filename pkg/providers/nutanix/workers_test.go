package nutanix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestWorkersSpec(t *testing.T) {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")

	logger := test.NewNullLogger()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster-multiple-worker-md.yaml")
	workers, err := WorkersSpec(context.TODO(), logger, client, spec)
	require.NoError(t, err)
	assert.Len(t, workers.Groups, 2)
}

func TestWorkersSpecWithUpgradeRolloutStrategy(t *testing.T) {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")

	logger := test.NewNullLogger()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	spec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
		{
			Count: ptr.Int(4),
			MachineGroupRef: &v1alpha1.Ref{
				Name: "eksa-unit-test",
			},
			Name: "eksa-unit-test",
			UpgradeRolloutStrategy: &v1alpha1.WorkerNodesUpgradeRolloutStrategy{
				RollingUpdate: &v1alpha1.WorkerNodesRollingUpdateParams{
					MaxSurge:       1,
					MaxUnavailable: 0,
				},
			},
		},
	}
	workers, err := WorkersSpec(context.TODO(), logger, client, spec)
	require.NoError(t, err)
	assert.Len(t, workers.Groups, 1)
	assert.Equal(t, int32(1), workers.Groups[0].MachineDeployment.Spec.Strategy.RollingUpdate.MaxSurge.IntVal)
	assert.Equal(t, int32(0), workers.Groups[0].MachineDeployment.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal)
}
