package clustermanager_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type upgraderTest struct {
	*WithT
	ctx         context.Context
	client      *mocks.MockClusterClient
	currentSpec *cluster.Spec
	newSpec     *cluster.Spec
	upgrader    *clustermanager.Upgrader
	cluster     *types.Cluster
}

func newUpgraderTest(t *testing.T) *upgraderTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClusterClient(ctrl)
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Eksa.Version = "v0.1.0"
		s.VersionsBundle.EksD.Name = "eks-d-1"
	})

	return &upgraderTest{
		WithT:  NewWithT(t),
		ctx:    context.Background(),
		client: client,
		upgrader: clustermanager.NewUpgrader(clustermanager.NewRetrierClient(clustermanager.NewClient(client),
			retrier.NewWithMaxRetries(1, 0))),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestUpgraderEksaUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.upgrader.EksaUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpgraderEksaUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.Expect(tt.upgrader.EksaUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpgraderEksaUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.VersionsBundle.Eksa.Version = "v0.2.0"
	tt.newSpec.VersionsBundle.Eksa.Components = v1alpha1.Manifest{
		URI: "testdata/test_components.yaml",
	}

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "EKS-A",
				NewVersion:    "v0.2.0",
				OldVersion:    "v0.1.0",
			},
		},
	}

	tt.client.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, []byte("test data")).Return(nil)
	tt.client.EXPECT().WaitForDeployment(tt.ctx, tt.cluster, "30m", "Available", "eksa-controller-manager", "eksa-system")
	tt.Expect(tt.upgrader.EksaUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderEksaUpgradeInstallError(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.VersionsBundle.Eksa.Version = "v0.2.0"

	// components file not set so this should return an error in failing to load manifest
	_, err := tt.upgrader.EksaUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}

func TestUpgraderEksdUpgradeNoSelfManaged(t *testing.T) {
	tt := newUpgraderTest(t)
	tt.newSpec.Cluster.SetManagedBy("management-cluster")

	tt.Expect(tt.upgrader.EksdUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpgraderEksdUpgradeNoChanges(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.Expect(tt.upgrader.EksdUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpgraderEksdUpgradeSuccess(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.VersionsBundle.EksD.Name = "eks-d-2"
	tt.newSpec.VersionsBundle.EksD.Components = "testdata/test_components.yaml"
	tt.newSpec.VersionsBundle.EksD.EksDReleaseUrl = "testdata/test_components.yaml"

	wantDiff := &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "EKS-D",
				NewVersion:    "eks-d-2",
				OldVersion:    "eks-d-1",
			},
		},
	}

	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	tt.Expect(tt.upgrader.EksdUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)).To(Equal(wantDiff))
}

func TestUpgraderEksdUpgradeInstallError(t *testing.T) {
	tt := newUpgraderTest(t)

	tt.newSpec.VersionsBundle.EksD.Name = "eks-d-2"

	// components file not set so this should return an error in failing to load manifest
	_, err := tt.upgrader.EksdUpgrade(tt.ctx, tt.cluster, tt.currentSpec, tt.newSpec)
	tt.Expect(err).NotTo(BeNil())
}
