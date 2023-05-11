package curatedpackages_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/version"
)

type bundleTest struct {
	*WithT
	ctx           context.Context
	kubeConfig    string
	kubeVersion   string
	kubeMajor     string
	kubeMinor     string
	cluster       string
	kubectl       *mocks.MockKubectlRunner
	bundleManager *mocks.MockManager
	Command       *curatedpackages.BundleReader
	activeBundle  string
	bundleCtrl    *packagesv1.PackageBundleController
	packageBundle *packagesv1.PackageBundle
	registry      *mocks.MockBundleRegistry
	cliVersion    version.Info
}

func newBundleTest(t *testing.T) *bundleTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	bm := mocks.NewMockManager(ctrl)
	kubeConfig := "test.kubeconfig"
	kubeVersion := "1.21"
	kubeMajor := "1"
	kubeMinor := "21"
	cluster := "billy"
	registry := mocks.NewMockBundleRegistry(ctrl)
	activeBundle := "v1.21-1000"
	cliVersion := version.Info{GitVersion: "v1.0.0"}
	bundleCtrl := packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: activeBundle,
		},
	}
	packageBundle := packagesv1.PackageBundle{
		Spec: packagesv1.PackageBundleSpec{
			Packages: []packagesv1.BundlePackage{
				{
					Name: "harbor",
				},
			},
		},
	}

	return &bundleTest{
		WithT:         NewWithT(t),
		ctx:           context.Background(),
		kubeConfig:    kubeConfig,
		kubeVersion:   kubeVersion,
		kubeMajor:     kubeMajor,
		kubeMinor:     kubeMinor,
		cluster:       cluster,
		kubectl:       k,
		bundleManager: bm,
		bundleCtrl:    &bundleCtrl,
		packageBundle: &packageBundle,
		activeBundle:  activeBundle,
		registry:      registry,
		cliVersion:    cliVersion,
	}
}

func TestGetLatestBundleFromClusterSucceeds(t *testing.T) {
	tt := newBundleTest(t)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(convertJsonToBytes(tt.bundleCtrl), nil)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(convertJsonToBytes(tt.packageBundle), nil)

	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)
	result, err := tt.Command.GetLatestBundle(tt.ctx, "")
	tt.Expect(err).To(BeNil())
	tt.Expect(result.Spec.Packages[0].Name).To(BeEquivalentTo(tt.packageBundle.Spec.Packages[0].Name))
}

func TestGetLatestBundleFromClusterFailsNoBundleName(t *testing.T) {
	tt := newBundleTest(t)
	noActiveBundle := tt.bundleCtrl
	noActiveBundle.Spec.ActiveBundle = ""
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(convertJsonToBytes(noActiveBundle), nil)

	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)
	result, err := tt.Command.GetLatestBundle(tt.ctx, "")
	tt.Expect(err).To(MatchError(ContainSubstring("no bundle name specified")))
	tt.Expect(result).To(BeNil())
}

func TestGetLatestBundleFromRegistrySucceeds(t *testing.T) {
	tt := newBundleTest(t)
	baseRef := "test_host/test_env/test_controller"
	tt.registry.EXPECT().GetRegistryBaseRef(tt.ctx).Return(baseRef, nil)
	tt.bundleManager.EXPECT().LatestBundle(tt.ctx, baseRef, tt.kubeMajor, tt.kubeMinor, "").Return(tt.packageBundle, nil)
	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, "", tt.kubectl, tt.bundleManager, tt.registry)
	result, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(BeNil())
	tt.Expect(result.Spec.Packages[0].Name).To(BeEquivalentTo(tt.packageBundle.Spec.Packages[0].Name))
}

func TestLatestBundleFromClusterUnknownBundle(t *testing.T) {
	tt := newBundleTest(t)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(convertJsonToBytes(tt.bundleCtrl), nil)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(bytes.Buffer{}, errors.New("error reading bundle"))
	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)
	_, err := tt.Command.GetLatestBundle(tt.ctx, "")
	tt.Expect(err).To(MatchError(ContainSubstring("error reading bundle")))
}

func TestGetLatestBundleFromRegistryWhenError(t *testing.T) {
	tt := newBundleTest(t)
	tt.registry.EXPECT().GetRegistryBaseRef(tt.ctx).Return("", errors.New("registry doesn't exist"))
	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, "", tt.kubectl, tt.bundleManager, tt.registry)
	_, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("registry doesn't exist")))
}

func TestLatestBundleFromClusterUnknownCtrl(t *testing.T) {
	tt := newBundleTest(t)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, gomock.Any()).Return(bytes.Buffer{}, errors.New("error fetching controller"))
	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)
	_, err := tt.Command.GetLatestBundle(tt.ctx, "")
	tt.Expect(err).To(MatchError(ContainSubstring("error fetching controller")))
}

func TestUpgradeBundleSucceeds(t *testing.T) {
	tt := newBundleTest(t)
	params := []string{"apply", "-f", "-", "--kubeconfig", tt.kubeConfig}
	newBundle := "new-bundle"
	expectedCtrl := packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: newBundle,
		},
	}
	ctrl, err := yaml.Marshal(expectedCtrl)
	tt.Expect(err).To(BeNil())
	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, ctrl, params).Return(bytes.Buffer{}, nil)

	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)

	err = tt.Command.UpgradeBundle(tt.ctx, tt.bundleCtrl, newBundle)
	tt.Expect(err).To(BeNil())
	tt.Expect(tt.bundleCtrl.Spec.ActiveBundle).To(Equal(newBundle))
}

func TestUpgradeBundleFails(t *testing.T) {
	tt := newBundleTest(t)
	params := []string{"apply", "-f", "-", "--kubeconfig", tt.kubeConfig}
	newBundle := "new-bundle"
	expectedCtrl := packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: newBundle,
		},
	}
	ctrl, err := yaml.Marshal(expectedCtrl)
	tt.Expect(err).To(BeNil())
	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, ctrl, params).Return(bytes.Buffer{}, errors.New("unable to apply yaml"))

	tt.Command = curatedpackages.NewBundleReader(tt.kubeConfig, tt.cluster, tt.kubectl, tt.bundleManager, tt.registry)

	err = tt.Command.UpgradeBundle(tt.ctx, tt.bundleCtrl, newBundle)
	tt.Expect(err).NotTo(BeNil())
}

func convertJsonToBytes(obj interface{}) bytes.Buffer {
	b, _ := json.Marshal(obj)
	return *bytes.NewBuffer(b)
}
