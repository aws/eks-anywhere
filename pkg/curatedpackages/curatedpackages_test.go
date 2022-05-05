package curatedpackages_test

import (
	"errors"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
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

func TestGetVersionBundleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	eksaVersion := "v1.0.0"
	kubeVersion := "1.21"
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					PackageController: releasev1.PackageBundle{
						Controller: releasev1.Image{
							URI: "test_host/test_env/test_repository:test-version",
						},
					},
					KubeVersion: kubeVersion,
				},
			},
		},
	}

	reader.EXPECT().ReadBundlesForVersion(eksaVersion).Return(bundles, nil)

	_, err := curatedpackages.GetVersionBundle(reader, eksaVersion, kubeVersion)
	if err != nil {
		t.Errorf("GetVersionBundle Should Pass When bundle exists")
	}
}

func TestGetVersionBundleFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	eksaVersion := "v1.0.0"
	kubeVersion := "1.21"

	reader.EXPECT().ReadBundlesForVersion(eksaVersion).Return(nil, errors.New("failed to read bundles"))

	_, err := curatedpackages.GetVersionBundle(reader, eksaVersion, kubeVersion)
	if err == nil {
		t.Errorf("GetVersionBundle should fail when no bundles exist")
	}
}
