package curatedpackages_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestValidateNoKubeVersionWhenClusterSucceeds(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("", "morby")
	if err != nil {
		t.Errorf("empty kubeVersion allowed when cluster specified")
	}
}

func TestValidateKubeVersionWhenClusterFails(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("1.21", "morby")
	if err == nil {
		t.Errorf("not both kube-version and cluster")
	}
}

func TestValidateKubeVersionWhenNoClusterFails(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("", "")
	if err == nil {
		t.Errorf("must specify cluster or kubeversion")
	}
}

func TestValidateKubeVersionWhenRegistrySucceeds(t *testing.T) {
	kubeVersion := "1.21"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, "")
	if err != nil {
		t.Errorf("Registry with %s should succeed", kubeVersion)
	}
}

func TestValidateKubeVersionWhenInvalidVersionFails(t *testing.T) {
	kubeVersion := "1.2.3"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, "")
	if err == nil {
		t.Errorf("Registry with %s should fail", kubeVersion)
	}
}
