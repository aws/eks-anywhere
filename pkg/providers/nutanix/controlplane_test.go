package nutanix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestControlPlaneSpec(t *testing.T) {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")
	logger := test.NewNullLogger()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cp, err := ControlPlaneSpec(context.TODO(), logger, client, spec)
	assert.NoError(t, err)
	assert.NotNil(t, cp)
}

func TestControlPlaneSpecWithUpgradeRolloutStrategy(t *testing.T) {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")
	logger := test.NewNullLogger()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &v1alpha1.ControlPlaneUpgradeRolloutStrategy{
		RollingUpdate: &v1alpha1.ControlPlaneRollingUpdateParams{
			MaxSurge: 1,
		},
	}
	cp, err := ControlPlaneSpec(context.TODO(), logger, client, spec)
	assert.NoError(t, err)
	assert.NotNil(t, cp)
	assert.Equal(t, int32(1), cp.KubeadmControlPlane.Spec.RolloutStrategy.RollingUpdate.MaxSurge.IntVal)
}

func TestCPObjects(t *testing.T) {
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")
	logger := test.NewNullLogger()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	cp, err := ControlPlaneSpec(context.TODO(), logger, client, spec)
	assert.NoError(t, err)

	objs := cp.Objects()
	assert.NotEqual(t, 0, len(objs))
}
