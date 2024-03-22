package management

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces/mocks"
)

var capiChangeDiff = types.NewChangeDiff(&types.ComponentChangeDiff{
	ComponentName: "vsphere",
	OldVersion:    "v0.0.1",
	NewVersion:    "v0.0.2",
})

var fluxChangeDiff = types.NewChangeDiff(&types.ComponentChangeDiff{
	ComponentName: "Flux",
	OldVersion:    "v0.0.1",
	NewVersion:    "v0.0.2",
})

var eksaChangeDiff = types.NewChangeDiff(&types.ComponentChangeDiff{
	ComponentName: "eks-a",
	OldVersion:    "v0.0.1",
	NewVersion:    "v0.0.2",
})

var managementComponentsVersionAnnotation = map[string]string{
	"anywhere.eks.amazonaws.com/management-components-version": "v0.19.0-dev+latest",
}

type TestMocks struct {
	mockCtrl       *gomock.Controller
	clientFactory  *mocks.MockClientFactory
	clusterManager *mocks.MockClusterManager
	gitOpsManager  *mocks.MockGitOpsManager
	provider       *providermocks.MockProvider
	writer         *writermocks.MockFileWriter
	eksdInstaller  *mocks.MockEksdInstaller
	eksdUpgrader   *mocks.MockEksdUpgrader
	capiManager    *mocks.MockCAPIManager
	validator      *mocks.MockValidator
}

func NewTestMocks(t *testing.T) *TestMocks {
	mockCtrl := gomock.NewController(t)
	return &TestMocks{
		mockCtrl:       mockCtrl,
		clientFactory:  mocks.NewMockClientFactory(mockCtrl),
		clusterManager: mocks.NewMockClusterManager(mockCtrl),
		gitOpsManager:  mocks.NewMockGitOpsManager(mockCtrl),
		provider:       providermocks.NewMockProvider(mockCtrl),
		writer:         writermocks.NewMockFileWriter(mockCtrl),
		eksdInstaller:  mocks.NewMockEksdInstaller(mockCtrl),
		eksdUpgrader:   mocks.NewMockEksdUpgrader(mockCtrl),
		capiManager:    mocks.NewMockCAPIManager(mockCtrl),
		validator:      mocks.NewMockValidator(mockCtrl),
	}
}

type upgradeManagementComponentsTest struct {
	ctx               context.Context
	runner            *UpgradeManagementComponentsWorkflow
	mocks             *TestMocks
	managementCluster *types.Cluster
	currentSpec       *cluster.Spec
	newSpec           *cluster.Spec
}

func newUpgradeManagementComponentsTest(t *testing.T) *upgradeManagementComponentsTest {
	mocks := NewTestMocks(t)
	runner := NewUpgradeManagementComponentsRunner(
		mocks.clientFactory,
		mocks.provider,
		mocks.capiManager,
		mocks.clusterManager,
		mocks.gitOpsManager,
		mocks.writer,
		mocks.eksdUpgrader,
		mocks.eksdInstaller,
	)

	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles = test.Bundle()
		s.EKSARelease = test.EKSARelease()
	})
	newSpec := currentSpec.DeepCopy()

	managementCluster := &types.Cluster{
		Name:           currentSpec.Cluster.Name,
		KubeconfigFile: kubeconfig.FromClusterName(currentSpec.Cluster.Name),
	}

	return &upgradeManagementComponentsTest{
		mocks:             mocks,
		runner:            runner,
		managementCluster: managementCluster,
		currentSpec:       currentSpec,
		newSpec:           newSpec,
	}
}

func TestRunnerHappyPath(t *testing.T) {
	tt := newUpgradeManagementComponentsTest(t)
	currentManagementComponents := cluster.ManagementComponentsFromBundles(tt.currentSpec.Bundles)
	newManagementComponents := cluster.ManagementComponentsFromBundles(tt.newSpec.Bundles)

	client := test.NewFakeKubeClient(tt.currentSpec.Cluster, tt.currentSpec.EKSARelease, tt.currentSpec.Bundles)
	tt.mocks.clusterManager.EXPECT().GetCurrentClusterSpec(tt.ctx, gomock.Any(), tt.managementCluster.Name).Return(tt.currentSpec, nil)
	gomock.InOrder(
		tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(nil),
		tt.mocks.provider.EXPECT().Name(),
		tt.mocks.provider.EXPECT().SetupAndValidateUpgradeManagementComponents(tt.ctx, tt.newSpec),
		tt.mocks.provider.EXPECT().PreCoreComponentsUpgrade(gomock.Any(), gomock.Any(), newManagementComponents, gomock.Any()),
		tt.mocks.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.managementCluster.KubeconfigFile).Return(client, nil),
		tt.mocks.capiManager.EXPECT().Upgrade(tt.ctx, tt.managementCluster, tt.mocks.provider, currentManagementComponents, newManagementComponents, tt.newSpec).Return(capiChangeDiff, nil),
		tt.mocks.gitOpsManager.EXPECT().Install(tt.ctx, tt.managementCluster, newManagementComponents, tt.currentSpec, tt.newSpec).Return(nil),
		tt.mocks.gitOpsManager.EXPECT().Upgrade(tt.ctx, tt.managementCluster, currentManagementComponents, newManagementComponents, tt.currentSpec, tt.newSpec).Return(fluxChangeDiff, nil),
		tt.mocks.clusterManager.EXPECT().Upgrade(tt.ctx, tt.managementCluster, currentManagementComponents, newManagementComponents, tt.newSpec).Return(eksaChangeDiff, nil),
		tt.mocks.eksdUpgrader.EXPECT().Upgrade(tt.ctx, tt.managementCluster, tt.currentSpec, tt.newSpec).Return(nil),
		tt.mocks.clusterManager.EXPECT().ApplyBundles(
			tt.ctx, tt.newSpec, tt.managementCluster,
		).Return(nil),
		tt.mocks.clusterManager.EXPECT().ApplyReleases(
			tt.ctx, tt.newSpec, tt.managementCluster,
		).Return(nil),
		tt.mocks.eksdInstaller.EXPECT().InstallEksdManifest(
			tt.ctx, tt.newSpec, tt.managementCluster,
		).Return(nil),
	)

	err := tt.runner.Run(tt.ctx, tt.newSpec, tt.managementCluster, tt.mocks.validator)
	if err != nil {
		t.Fatalf("UpgradeManagementComponents.Run() err = %v, want err = nil", err)
	}

	g := NewWithT(t)
	g.Expect(tt.newSpec.Cluster.Annotations).To(Equal(managementComponentsVersionAnnotation))
}

func TestRunnerStopsWhenValidationFailed(t *testing.T) {
	tt := newUpgradeManagementComponentsTest(t)
	tt.mocks.provider.EXPECT().Name()
	tt.mocks.provider.EXPECT().SetupAndValidateUpgradeManagementComponents(tt.ctx, tt.newSpec)
	tt.mocks.clusterManager.EXPECT().GetCurrentClusterSpec(tt.ctx, gomock.Any(), tt.managementCluster.Name).Return(tt.currentSpec, nil)
	tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(
		[]validations.Validation{
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Err: errors.New("validation failed"),
				}
			},
		})

	tt.mocks.writer.EXPECT().Write(fmt.Sprintf("%s-checkpoint.yaml", tt.newSpec.Cluster.Name), gomock.Any())
	err := tt.runner.Run(tt.ctx, tt.newSpec, tt.managementCluster, tt.mocks.validator)
	if err == nil {
		t.Fatalf("UpgradeManagementComponents.Run() err == nil, want err != nil")
	}
}
