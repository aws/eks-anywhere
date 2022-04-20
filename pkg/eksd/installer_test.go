package eksd_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/eksd"
	"github.com/aws/eks-anywhere/pkg/eksd/mocks"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
)

type installerTest struct {
	*WithT
	ctx         context.Context
	client      *mocks.MockEksdInstallerClient
	clusterSpec *cluster.Spec
	eksd        *eksd.Installer
	cluster     *types.Cluster
}

func newInstallerTest(t *testing.T) *installerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockEksdInstallerClient(ctrl)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.EksD.Name = "eks-d-1"
	})

	return &installerTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		client:      client,
		eksd:        eksd.NewEksdInstaller(client),
		clusterSpec: clusterSpec,
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestInstallEksdComponentsSuccess(t *testing.T) {
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	err := os.Unsetenv(features.CloudStackProviderEnvVar)
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)
	if err != nil {
		return
	}

	tt := newInstallerTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Components = "testdata/testdata.yaml"
	tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl = "testdata/testdata.yaml"

	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	if err := tt.eksd.InstallEksdCRDs(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("Eksd.InstallEksdComponents() error = %v, wantErr nil", err)
	}
}

func TestInstallEksdComponentsErrorReadingManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Components = "fake.yaml"

	if err := tt.eksd.InstallEksdCRDs(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("Eksd.InstallEksdComponents() error = nil, wantErr not nil")
	}
}
