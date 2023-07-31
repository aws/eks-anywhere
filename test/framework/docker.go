package framework

import (
	"os"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

// Docker is a Provider for running end-to-end tests.
type Docker struct {
	t *testing.T
	executables.Docker
}

const dockerPodCidrVar = "T_DOCKER_POD_CIDR"

// NewDocker creates a new Docker object implementing the Provider interface
// for testing.
func NewDocker(t *testing.T) *Docker {
	docker := executables.BuildDockerExecutable()
	return &Docker{
		t:      t,
		Docker: *docker,
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

// UpdateKubeConfig customizes generated kubeconfig by replacing the server value with correct host
// and the docker LB port. This is required for the docker provider.
func (d *Docker) UpdateKubeConfig(content *[]byte, clusterName string) error {
	dockerClient := executables.BuildDockerExecutable()
	p := docker.NewProvider(
		nil,
		dockerClient,
		nil,
		time.Now,
	)
	return p.UpdateKubeConfig(content, clusterName)
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

// WithNewWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding DockerMachineConfig to the cluster config.
func (d *Docker) WithNewWorkerNodeGroup(machineConfig string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.ClusterToConfigFiller(workerNodeGroup.ClusterFiller())
}

// ClusterStateValidations returns a list of provider specific validations.
func (d *Docker) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version.
func (d *Docker) WithKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
	)
}
