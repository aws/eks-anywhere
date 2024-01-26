package clusterapi_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/mocks"
	providerMocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type upgraderTest struct {
	*WithT
	ctx                   context.Context
	capiClient            *mocks.MockCAPIClient
	kubectlClient         *mocks.MockKubectlClient
	upgrader              *clusterapi.Upgrader
	currentVersionsBundle *releasev1alpha1.VersionsBundle
	newVersionsBundle     *releasev1alpha1.VersionsBundle
	newSpec               *cluster.Spec
	cluster               *types.Cluster
	provider              *providerMocks.MockProvider
	providerChangeDiff    *types.ComponentChangeDiff
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	capiClient := mocks.NewMockCAPIClient(ctrl)
	kubectlClient := mocks.NewMockKubectlClient(ctrl)

	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles.Spec.Number = 1
	})

	currentVersionsBundle := cluster.ManagementComponentsVersionsBundle(currentSpec.Bundles)
	currentVersionsBundle.CertManager.Version = "v0.1.0"
	currentVersionsBundle.ClusterAPI.Version = "v0.1.0"
	currentVersionsBundle.ControlPlane.Version = "v0.1.0"
	currentVersionsBundle.Bootstrap.Version = "v0.1.0"
	currentVersionsBundle.ExternalEtcdBootstrap.Version = "v0.1.0"
	currentVersionsBundle.ExternalEtcdController.Version = "v0.1.0"

	newSpec := currentSpec.DeepCopy()
	newVersionsBundle := cluster.ManagementComponentsVersionsBundle(newSpec.Bundles)

	return &upgraderTest{
		WithT:                 NewWithT(t),
		ctx:                   context.Background(),
		capiClient:            capiClient,
		kubectlClient:         kubectlClient,
		upgrader:              clusterapi.NewUpgrader(capiClient, kubectlClient),
		currentVersionsBundle: currentVersionsBundle,
		newVersionsBundle:     newVersionsBundle,
		newSpec:               newSpec,
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
		provider: providerMocks.NewMockProvider(ctrl),
		providerChangeDiff: &types.ComponentChangeDiff{
			ComponentName: "vsphere",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
	}
}

func TestUpgraderUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(BeNil())
}

func TestUpgraderUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(nil)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(BeNil())
}

func TestUpgraderUpgradeProviderChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: tt.providerChangeDiff,
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{*tt.providerChangeDiff},
	}

	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderUpgradeCoreChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ClusterAPI.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
		Core: &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{*changeDiff.Core},
	}

	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(nil)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderUpgradeEverythingChangesStackedEtcd(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Bundles.Spec.VersionsBundles[0].CertManager.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ClusterAPI.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ControlPlane.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].Bootstrap.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ExternalEtcdBootstrap.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ExternalEtcdController.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
		CertManager: &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		Core: &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		ControlPlane: &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		BootstrapProviders: []types.ComponentChangeDiff{
			{
				ComponentName: "kubeadm",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
			{
				ComponentName: "etcdadm-bootstrap",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
			{
				ComponentName: "etcdadm-controller",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
		InfrastructureProvider: tt.providerChangeDiff,
	}

	components := []types.ComponentChangeDiff{*changeDiff.CertManager, *changeDiff.Core, *changeDiff.ControlPlane, *tt.providerChangeDiff}
	bootstrapProviders := append(components, changeDiff.BootstrapProviders...)
	wantDiff := &types.ChangeDiff{
		ComponentReports: bootstrapProviders,
	}

	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderUpgradeEverythingChangesExternalEtcd(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{}
	tt.newSpec.Bundles.Spec.VersionsBundles[0].CertManager.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ClusterAPI.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ControlPlane.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].Bootstrap.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ExternalEtcdBootstrap.Version = "v0.2.0"
	tt.newSpec.Bundles.Spec.VersionsBundles[0].ExternalEtcdController.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
		CertManager: &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		Core: &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		ControlPlane: &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
		BootstrapProviders: []types.ComponentChangeDiff{
			{
				ComponentName: "kubeadm",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
			{
				ComponentName: "etcdadm-bootstrap",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
			{
				ComponentName: "etcdadm-controller",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
		InfrastructureProvider: tt.providerChangeDiff,
	}
	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			*changeDiff.CertManager, *changeDiff.Core, *changeDiff.ControlPlane, *tt.providerChangeDiff,
			changeDiff.BootstrapProviders[0],
			changeDiff.BootstrapProviders[1],
			changeDiff.BootstrapProviders[2],
		},
	}

	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderUpgradeCAPIClientError(t *testing.T) {
	tt := newUpgraderTest(t)
	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: tt.providerChangeDiff,
	}

	tt.provider.EXPECT().ChangeDiff(tt.currentVersionsBundle, &tt.newSpec.Bundles.Spec.VersionsBundles[0]).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff).Return(errors.New("error from client"))

	_, err := tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentVersionsBundle, tt.newVersionsBundle, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}
