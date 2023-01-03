package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

// Docker is a Provider for running end-to-end tests.
type Docker struct {
	t *testing.T
}

const dockerPodCidrVar = "T_DOCKER_POD_CIDR"

// NewDocker creates a new Docker object implementing the Provider interface
// for testing.
func NewDocker(t *testing.T) *Docker {
	return &Docker{
		t: t,
	}
}

// Name implements the Provider interface.
func (d *Docker) Name() string {
	return "docker"
}

// Setup implements the Provider interface.
func (d *Docker) Setup() {}

// CleanupVMs implements the Provider interface.
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
	f := []api.ClusterFiller{}
	podCidr := os.Getenv(dockerPodCidrVar)
	if podCidr != "" {
		f = append(f, api.WithPodCidr(podCidr))
	}
	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...)}
}

// WithWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding DockerMachineConfig to the cluster config.
func (d *Docker) WithWorkerNodeGroup(workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.ClusterToConfigFiller(workerNodeGroup.ClusterFiller())
}
