package vsphere

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
)

const testDataClusterConfigMainFailueDomainFilename = "cluster_vsphere_failuredomain.yaml"

func TestFailureDomainsSpecSuccess(t *testing.T) {
	spec := test.NewFullClusterSpec(t, path.Join(testDataDir, testDataClusterConfigMainFailueDomainFilename))
	logger := test.NewNullLogger()
	failureDomains, err := FailureDomainsSpec(logger, spec)
	assert.Nil(t, err)
	assert.True(t, len(failureDomains.Objects()) > 0)
	assert.Equal(t, failureDomains.Groups[0].VsphereDeploymentZone.Name, "test-test-fd-1")
}
