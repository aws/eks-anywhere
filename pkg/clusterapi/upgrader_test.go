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
)

type upgraderTest struct {
	*WithT
	ctx                context.Context
	capiClient         *mocks.MockCAPIClient
	upgrader           *clusterapi.Upgrader
	currentSpec        *cluster.Spec
	newSpec            *cluster.Spec
	cluster            *types.Cluster
	provider           *providerMocks.MockProvider
	providerChangeDiff *types.ComponentChangeDiff
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	capiClient := mocks.NewMockCAPIClient(ctrl)

	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles.Spec.Number = 1
		s.VersionsBundle.ClusterAPI.Version = "v0.1.0"
		s.VersionsBundle.ControlPlane.Version = "v0.1.0"
		s.VersionsBundle.Bootstrap.Version = "v0.1.0"
		s.VersionsBundle.ExternalEtcdBootstrap.Version = "v0.1.0"
		s.VersionsBundle.ExternalEtcdController.Version = "v0.1.0"
	})

	return &upgraderTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		capiClient:  capiClient,
		upgrader:    clusterapi.NewUpgrader(capiClient),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
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

func TestUpgraderUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(nil)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestUpgraderUpgradeProviderChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: tt.providerChangeDiff,
	}
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestUpgraderUpgradeCoreChanges(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.ClusterAPI.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
		Core: &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    "v0.2.0",
			OldVersion:    "v0.1.0",
		},
	}
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(nil)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestUpgraderUpgradeEverythingChangesStackedEtcd(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.VersionsBundle.ClusterAPI.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ControlPlane.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.Bootstrap.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ExternalEtcdBootstrap.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ExternalEtcdController.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
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
		},
		InfrastructureProvider: tt.providerChangeDiff,
	}
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestUpgraderUpgradeEverythingChangesExternalEtcd(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{}
	tt.newSpec.VersionsBundle.ClusterAPI.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ControlPlane.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.Bootstrap.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ExternalEtcdBootstrap.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.ExternalEtcdController.Version = "v0.2.0"
	changeDiff := &clusterapi.CAPIChangeDiff{
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
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff)

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).To(Succeed())
}

func TestUpgraderUpgradeCAPIClientError(t *testing.T) {
	tt := newUpgraderTest(t)
	changeDiff := &clusterapi.CAPIChangeDiff{
		InfrastructureProvider: tt.providerChangeDiff,
	}
	tt.provider.EXPECT().ChangeDiff(tt.currentSpec, tt.newSpec).Return(tt.providerChangeDiff)
	tt.capiClient.EXPECT().Upgrade(tt.ctx, tt.cluster, tt.provider, tt.newSpec, changeDiff).Return(errors.New("error from client"))

	tt.Expect(tt.upgrader.Upgrade(tt.ctx, tt.cluster, tt.provider, tt.currentSpec, tt.newSpec)).NotTo(Succeed())
}
