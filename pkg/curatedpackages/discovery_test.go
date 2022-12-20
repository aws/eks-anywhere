package curatedpackages_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestServerVersionSucceeds(t *testing.T) {
	kubeVersion := curatedpackages.NewKubeVersion("1", "21")
	discovery := curatedpackages.NewDiscovery(kubeVersion)

	_, err := discovery.ServerVersion()
	if err != nil {
		t.Errorf("Server Version should succeed when valid kubernetes version is provided")
	}
}
