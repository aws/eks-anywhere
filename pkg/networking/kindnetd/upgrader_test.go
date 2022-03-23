package kindnetd_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
	"github.com/aws/eks-anywhere/pkg/types"
)

type upgraderTest struct {
	*kindnetdTest
	ctx                  context.Context
	u                    *kindnetd.Upgrader
	manifest             []byte
	currentSpec, newSpec *cluster.Spec
	cluster              *types.Cluster
	wantChangeDiff       *types.ChangeDiff
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	kt := newKindnetdTest(t)
	u := kindnetd.NewUpgrader(kt.client)
	return &upgraderTest{
		kindnetdTest: kt,
		ctx:          context.Background(),
		u:            u,
		manifest:     []byte(test.ReadFile(t, "testdata/expected_kindnetd_manifest.yaml")),
		currentSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.1.0/24"}
			s.VersionsBundle.Kindnetd = *KindnetdBundle.DeepCopy()
			s.VersionsBundle.Kindnetd.Version = "v1.9.10-eksa.1"
		}),
		newSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.1.0/24"}
			s.VersionsBundle.Kindnetd = *KindnetdBundle.DeepCopy()
			s.VersionsBundle.Kindnetd.Version = "v1.9.11-eksa.1"
		}),
		cluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
		wantChangeDiff: types.NewChangeDiff(&types.ComponentChangeDiff{
			ComponentName: "kindnetd",
			OldVersion:    "v1.9.10-eksa.1",
			NewVersion:    "v1.9.11-eksa.1",
		}),
	}
}

func TestUpgraderUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, tt.manifest)

	tt.Expect(tt.u.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec, []string{})).To(Equal(tt.wantChangeDiff), "upgrader.Upgrade() should succeed and return correct ChangeDiff")
}

func TestUpgraderUpgradeNotNeeded(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.currentSpec.VersionsBundle.Kindnetd.Version = "v1.0.0"
	tt.newSpec.VersionsBundle.Kindnetd.Version = "v1.0.0"

	tt.Expect(tt.u.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec, []string{})).To(BeNil(), "upgrader.Upgrade() should succeed and return nil ChangeDiff")
}
