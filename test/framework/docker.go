package framework

import (
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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

func (d *Docker) CustomizeProviderConfig(file string) []byte {
	providerConfig, err := v1alpha1.GetDockerDatacenterConfig(file)
	if err != nil {
		d.t.Fatalf("Unable to get provider config from file: %v", err)
	}

	providerOutput, err := yaml.Marshal(providerConfig)
	if err != nil {
		d.t.Fatalf("error marshalling cluster config: %v", err)
	}

	return providerOutput
}

func (d *Docker) WithProviderUpgradeGit() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = d.CustomizeProviderConfig(e.clusterConfigGitPath())
	}
}

func (d *Docker) ClusterConfigFillers() []api.ClusterFiller {
	return nil
}
