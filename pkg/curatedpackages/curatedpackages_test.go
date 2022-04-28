package curatedpackages_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestCreateBundleManagerWhenValidKubeVersion(t *testing.T) {
	bm := curatedpackages.CreateBundleManager("1.21")
	if bm == nil {
		t.Errorf("Bundle Manager should be successful when valid kubeversion")
	}
}

func TestCreateBundleManagerWhenInValidKubeVersion(t *testing.T) {
	bm := curatedpackages.CreateBundleManager("1")
	if bm != nil {
		t.Errorf("Bundle Manager should be successful when valid kubeversion")
	}
}
