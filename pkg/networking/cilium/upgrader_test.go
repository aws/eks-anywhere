package cilium

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

type upgraderTest struct {
	*WithT
	ctx                   context.Context
	u                     *Upgrader
	h                     *mocks.MockHelm
	client                *mocks.MockupgraderClient
	manifestPre, manifest []byte
	currentSpec, newSpec  *cluster.Spec
	cluster               *types.Cluster
	wantChangeDiff        *types.ChangeDiff
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	client := mocks.NewMockupgraderClient(ctrl)
	u := NewUpgrader(nil, h)
	u.client = client
	return &upgraderTest{
		WithT:    NewWithT(t),
		ctx:      context.Background(),
		h:        h,
		client:   client,
		u:        u,
		manifest: []byte("manifestContent"),
		currentSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium.Version = "v1.9.10-eksa.1"
		}),
		newSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium.Version = "v1.9.11-eksa.1"
		}),
		cluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
		wantChangeDiff: types.NewChangeDiff(&types.ComponentChangeDiff{
			ComponentName: "cilium",
			OldVersion:    "v1.9.10-eksa.1",
			NewVersion:    "v1.9.11-eksa.1",
		}),
	}
}

func (tt *upgraderTest) expectTemplatePreFlight() *gomock.Call {
	return tt.expectTemplate(tt.manifestPre)
}

func (tt *upgraderTest) expectTemplateManifest() *gomock.Call {
	return tt.expectTemplate(tt.manifest)
}

func (tt *upgraderTest) expectTemplate(manifest []byte) *gomock.Call {
	// Using Any because this already tested in the templater tests
	return tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(manifest, nil)
}

func TestUpgraderUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)
	// Templater and client and already tested individually so we only want to test the flow (order of calls)
	gomock.InOrder(
		tt.expectTemplatePreFlight(),
		tt.client.EXPECT().Apply(tt.ctx, tt.cluster, tt.manifestPre),
		tt.client.EXPECT().WaitForPreflightDaemonSet(tt.ctx, tt.cluster),
		tt.client.EXPECT().WaitForPreflightDeployment(tt.ctx, tt.cluster),
		tt.client.EXPECT().DeleteKubeSpecFromBytes(tt.ctx, tt.cluster, tt.manifestPre),
		tt.expectTemplateManifest(),
		tt.client.EXPECT().Apply(tt.ctx, tt.cluster, tt.manifest),
		tt.client.EXPECT().WaitForCiliumDaemonSet(tt.ctx, tt.cluster),
		tt.client.EXPECT().WaitForCiliumDeployment(tt.ctx, tt.cluster),
	)

	tt.Expect(tt.u.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(tt.wantChangeDiff), "upgrader.Upgrade() should succeed and return correct ChangeDiff")
}

func TestUpgraderUpgradeNotNeeded(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.currentSpec.VersionsBundle.Cilium.Version = "v1.0.0"
	tt.newSpec.VersionsBundle.Cilium.Version = "v1.0.0"

	tt.Expect(tt.u.Upgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil(), "upgrader.Upgrade() should succeed and return nil ChangeDiff")
}
