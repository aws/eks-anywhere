package curatedpackages_test

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestCreateBundleManagerWhenValidKubeVersion(t *testing.T) {
	bm := curatedpackages.CreateBundleManager(test.NewNullLogger())
	if bm == nil {
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

	clusterSpec := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube121,
		},
	}

	reader.EXPECT().ReadBundlesForVersion(eksaVersion).Return(bundles, nil)

	_, err := curatedpackages.GetVersionBundle(reader, eksaVersion, clusterSpec)
	if err != nil {
		t.Errorf("GetVersionBundle Should Pass When bundle exists")
	}
}

func TestGetVersionBundleFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	eksaVersion := "v1.0.0"

	reader.EXPECT().ReadBundlesForVersion(eksaVersion).Return(nil, errors.New("failed to read bundles"))

	clusterSpec := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube121,
		},
	}
	_, err := curatedpackages.GetVersionBundle(reader, eksaVersion, clusterSpec)
	if err == nil {
		t.Errorf("GetVersionBundle should fail when no bundles exist")
	}
}

func TestGetVersionBundleFailsWhenBundleNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	eksaVersion := "v1.0.0"
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					PackageController: releasev1.PackageBundle{
						Controller: releasev1.Image{
							URI: "test_host/test_env/test_repository:test-version",
						},
					},
					KubeVersion: "1.22",
				},
			},
		},
	}

	clusterSpec := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube121,
		},
	}

	reader.EXPECT().ReadBundlesForVersion(eksaVersion).Return(bundles, nil)

	_, err := curatedpackages.GetVersionBundle(reader, eksaVersion, clusterSpec)
	if err == nil {
		t.Errorf("GetVersionBundle should fail when version bundle for kubeversion doesn't exist")
	}
}

func TestGetRegistrySuccess(t *testing.T) {
	g := NewWithT(t)
	uri := "public.ecr.aws/l0g8r8j6/eks-anywhere-packages"
	registry := curatedpackages.GetRegistry(uri)
	expected := "public.ecr.aws/l0g8r8j6"
	g.Expect(registry).To(Equal(expected))
}

func TestGetRegistryFail(t *testing.T) {
	g := NewWithT(t)
	uri := "public.ecr.aws"
	registry := curatedpackages.GetRegistry(uri)
	expected := "public.ecr.aws"
	g.Expect(registry).To(Equal(expected))
}
