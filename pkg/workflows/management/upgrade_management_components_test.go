package management

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	vm "github.com/aws/eks-anywhere/pkg/validations/mocks"
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

func TestPreflightValidationsSkipVersionSkew(t *testing.T) {
	tests := []struct {
		testName        string
		skipValidations []string
		setupMocks      func(*upgradeManagementComponentsTest)
		wantValidations int
	}{
		{
			testName:        "no skip validations - includes version skew check",
			skipValidations: []string{},
			setupMocks: func(tt *upgradeManagementComponentsTest) {
				tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(nil)
			},
			wantValidations: 3,
		},
		{
			testName:        "skip version skew validation",
			skipValidations: []string{validations.EksaVersionSkew},
			setupMocks: func(tt *upgradeManagementComponentsTest) {
				tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(nil)
			},
			wantValidations: 2,
		},
		{
			testName:        "skip non-existent validation - should include version skew",
			skipValidations: []string{"non-existent"},
			setupMocks: func(tt *upgradeManagementComponentsTest) {
				tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(nil)
			},
			wantValidations: 3,
		},
		{
			testName:        "multiple skip validations - only version skew should be skipped",
			skipValidations: []string{validations.EksaVersionSkew, "other-validation"},
			setupMocks: func(tt *upgradeManagementComponentsTest) {
				tt.mocks.validator.EXPECT().PreflightValidations(tt.ctx).Return(nil)
			},
			wantValidations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			compTest := newUpgradeManagementComponentsTest(t)
			g := NewWithT(t)
			tt.setupMocks(compTest)
			currentManagementComponents := cluster.ManagementComponentsFromBundles(compTest.currentSpec.Bundles)
			newManagementComponents := cluster.ManagementComponentsFromBundles(compTest.newSpec.Bundles)
			client := test.NewFakeKubeClient(compTest.currentSpec.Cluster, compTest.currentSpec.EKSARelease, compTest.currentSpec.Bundles)
			mockCtrl := gomock.NewController(t)
			k := vm.NewMockKubectlClient(mockCtrl)
			compTest.mocks.clusterManager.EXPECT().GetCurrentClusterSpec(compTest.ctx, gomock.Any(), compTest.managementCluster.Name).Return(compTest.currentSpec, nil)
			compTest.mocks.provider.EXPECT().Name()
			compTest.mocks.provider.EXPECT().SetupAndValidateUpgradeManagementComponents(compTest.ctx, compTest.newSpec)
			compTest.mocks.provider.EXPECT().PreCoreComponentsUpgrade(gomock.Any(), gomock.Any(), newManagementComponents, gomock.Any())
			compTest.mocks.clientFactory.EXPECT().BuildClientFromKubeconfig(compTest.managementCluster.KubeconfigFile).Return(client, nil)
			compTest.mocks.capiManager.EXPECT().Upgrade(compTest.ctx, compTest.managementCluster, compTest.mocks.provider, currentManagementComponents, newManagementComponents, compTest.newSpec).Return(capiChangeDiff, nil)
			compTest.mocks.gitOpsManager.EXPECT().Install(compTest.ctx, compTest.managementCluster, newManagementComponents, compTest.currentSpec, compTest.newSpec).Return(nil)
			compTest.mocks.gitOpsManager.EXPECT().Upgrade(compTest.ctx, compTest.managementCluster, currentManagementComponents, newManagementComponents, compTest.currentSpec, compTest.newSpec).Return(fluxChangeDiff, nil)
			compTest.mocks.clusterManager.EXPECT().Upgrade(compTest.ctx, compTest.managementCluster, currentManagementComponents, newManagementComponents, compTest.newSpec)
			compTest.mocks.eksdUpgrader.EXPECT().Upgrade(compTest.ctx, compTest.managementCluster, compTest.currentSpec, compTest.newSpec).Return(nil)
			compTest.mocks.clusterManager.EXPECT().ApplyBundles(
				compTest.ctx, compTest.newSpec, compTest.managementCluster,
			).Return(nil)
			compTest.mocks.clusterManager.EXPECT().ApplyReleases(
				compTest.ctx, compTest.newSpec, compTest.managementCluster,
			).Return(nil)
			compTest.mocks.eksdInstaller.EXPECT().InstallEksdManifest(
				compTest.ctx, compTest.newSpec, compTest.managementCluster,
			).Return(nil)
			k.EXPECT().ValidateControlPlaneNodes(compTest.ctx, compTest.managementCluster, compTest.managementCluster.Name).Return(nil)
			k.EXPECT().ValidateClustersCRD(compTest.ctx, compTest.managementCluster).Return(nil)

			version := v1alpha1.EksaVersion("v0.20.6")
			mgmt := &v1alpha1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "mgmt-cluster",
				},
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: "1.30",
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: "mgmt-cluster",
					},
					EksaVersion: &version,
				},
			}

			if !slices.Contains(tt.skipValidations, validations.EksaVersionSkew) {
				k.EXPECT().GetEksaCluster(compTest.ctx, compTest.managementCluster, compTest.managementCluster.Name).Return(mgmt, nil)
			}
			validator := NewUMCValidator(compTest.managementCluster, compTest.newSpec.EKSARelease, k, tt.skipValidations)
			err := compTest.runner.Run(compTest.ctx, compTest.newSpec, compTest.managementCluster, compTest.mocks.validator)
			g.Expect(err).To(BeNil())

			preflights := validator.PreflightValidations(compTest.ctx)
			g.Expect(preflights).To(HaveLen(tt.wantValidations))
			hasVersionSkew := false
			for _, v := range preflights {
				result := v()
				g.Expect(result).ToNot(BeNil())
				g.Expect(result.Name).ToNot(BeEmpty())

				if result.Name == "validate compatibility of management components version to cluster eksaVersion" {
					hasVersionSkew = true
				}
			}

			if slices.Contains(tt.skipValidations, validations.EksaVersionSkew) {
				g.Expect(hasVersionSkew).To(BeFalse(), "version skew validation should be skipped")
			} else {
				g.Expect(hasVersionSkew).To(BeTrue(), "version skew validation should be present")
			}
		})
	}
}
