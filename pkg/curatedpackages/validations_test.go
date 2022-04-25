package curatedpackages_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestValidateKubeVersionWhenClusterSucceeds(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("", curatedpackages.Cluster)
	if err != nil {
		t.Errorf("Cluster Source has no kubeVersion validation")
	}
}

func TestValidateKubeVersionWhenRegistrySucceeds(t *testing.T) {
	kubeVersion := "1.21"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, curatedpackages.Registry)
	if err != nil {
		t.Errorf("Registry with %s should succeed", kubeVersion)
	}
}

func TestValidateKubeVersionWhenInvalidVersionFails(t *testing.T) {
	kubeVersion := "1.2.3"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, curatedpackages.Registry)
	if err == nil {
		t.Errorf("Registry with %s should fail", kubeVersion)
	}
}
