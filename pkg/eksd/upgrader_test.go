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
)

type upgraderTest struct {
	*WithT
	ctx          context.Context
	client       *mocks.MockEksdInstallerClient
	reader       *m.MockReader
	currentSpec  *cluster.Spec
	newSpec      *cluster.Spec
	eksdUpgrader *eksd.Upgrader
	cluster      *types.Cluster
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockEksdInstallerClient(ctrl)
	reader := m.NewMockReader(ctrl)
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].EksD.Name = "eks-d-1"
	})

	return &upgraderTest{
		WithT:        NewWithT(t),
		ctx:          context.Background(),
		client:       client,
		reader:       reader,
		eksdUpgrader: eksd.NewUpgrader(client, reader),
		currentSpec:  currentSpec,
		newSpec:      currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestEksdUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestEksdUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestEksdUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.VersionsBundles["1.19"].EksD.Name = "eks-d-2"
	tt.newSpec.Bundles = bundle()

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "EKS-D",
				NewVersion:    "eks-d-2",
				OldVersion:    "eks-d-1",
			},
		},
	}

	tt.reader.EXPECT().ReadFile(testdataFile).Return([]byte("test data"), nil)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	tt.Expect(tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderEksdUpgradeInstallError(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.eksdUpgrader.SetRetrier(retrier.NewWithMaxRetries(1, 0))
	tt.newSpec.VersionsBundles["1.19"].EksD.Name = "eks-d-2"
	tt.newSpec.Bundles = bundle()
	tt.newSpec.Bundles.Spec.VersionsBundles[0].EksD.Components = ""
	tt.newSpec.Bundles.Spec.VersionsBundles[1].EksD.Components = ""

	tt.reader.EXPECT().ReadFile(tt.newSpec.Bundles.Spec.VersionsBundles[0].EksD.Components).Return([]byte(""), fmt.Errorf("error"))
	// components file not set so this should return an error in failing to load manifest
	_, err := tt.eksdUpgrader.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}
