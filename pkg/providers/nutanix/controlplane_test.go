package nutanix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
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
