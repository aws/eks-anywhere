package framework

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

type Docker struct {
	t *testing.T
}

func NewDocker(t *testing.T) *Docker {
	return &Docker{
		t: t,
	}
}

func (d *Docker) Name() string {
	return "docker"
}

func (d *Docker) Setup() {}

func (d *Docker) CleanupVMs(_ string) error {
	return nil
}

func (d *Docker) WithProviderUpgradeGit() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		// There is no config for docker api objects, no-op
	}
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (d *Docker) ClusterConfigUpdates() []api.ClusterConfigFiller {
	return nil
}
