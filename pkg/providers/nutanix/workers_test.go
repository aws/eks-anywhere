package nutanix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/constants"
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
