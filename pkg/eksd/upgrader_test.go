package eksd_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	m "github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/eksd"
	"github.com/aws/eks-anywhere/pkg/eksd/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type upgraderTest struct {
	*WithT
	ctx                                       context.Context
	client                                    *mocks.MockEksdInstallerClient
	reader                                    *m.MockReader
	currentManagementComponentsVersionsBundle *releasev1alpha1.VersionsBundle
	newSpec                                   *cluster.Spec
	eksdUpgrader                              *eksd.Upgrader
	cluster                                   *types.Cluster
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockEksdInstallerClient(ctrl)
	reader := m.NewMockReader(ctrl)
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.FirstVersionsBundle().EksD.Name = "kubernetes-1-19-eks-1"
		s.FirstVersionsBundle().EksD.ReleaseChannel = "1-19"
		s.FirstVersionsBundle().EksD.KubeVersion = "v1.19.1"
	})

	currentManagementComponentsVersionsBundle := currentSpec.FirstVersionsBundle().DeepCopy()

	return &upgraderTest{
		WithT:        NewWithT(t),
		ctx:          context.Background(),
		client:       client,
		reader:       reader,
		eksdUpgrader: eksd.NewUpgrader(client, reader),
		currentManagementComponentsVersionsBundle: currentManagementComponentsVersionsBundle,
		newSpec: currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestEksdUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponentsVersionsBundle, tt.newSpec)).To(BeNil())
}

func TestEksdUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponentsVersionsBundle, tt.newSpec)).To(BeNil())
}

func TestEksdUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.Bundles = bundle()
	tt.newSpec.FirstVersionsBundle().EksD.Name = "kubernetes-1-19-eks-2"
	tt.newSpec.FirstVersionsBundle().EksD.ReleaseChannel = "1-19"
	tt.newSpec.FirstVersionsBundle().EksD.KubeVersion = "v1.19.1"

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "kubernetes",
				OldVersion:    "v1.19.1-eks-1-19-1",
				NewVersion:    "v1.19.1-eks-1-19-2",
			},
		},
	}

	tt.reader.EXPECT().ReadFile(testdataFile).Return([]byte("test data"), nil)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponentsVersionsBundle, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderEksdUpgradeInstallError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.eksdUpgrader.SetRetrier(retrier.NewWithMaxRetries(1, 0))
	tt.newSpec.Bundles = bundle()
	tt.newSpec.FirstVersionsBundle().EksD.Name = "kubernetes-1-19-eks-2"
	tt.newSpec.FirstVersionsBundle().EksD.Components = ""
	tt.newSpec.Bundles.Spec.VersionsBundles[1].EksD.Components = ""

	tt.reader.EXPECT().ReadFile(tt.newSpec.FirstVersionsBundle().EksD.Components).Return([]byte(""), fmt.Errorf("error"))
	// components file not set so this should return an error in failing to load manifest
	_, err := tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentManagementComponentsVersionsBundle, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}
