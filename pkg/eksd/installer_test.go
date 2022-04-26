package eksd_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	m "github.com/aws/eks-anywhere/internal/test/mocks"
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
	reader      *m.MockReader
	clusterSpec *cluster.Spec
	eksd        *eksd.Installer
	cluster     *types.Cluster
}

func newInstallerTest(t *testing.T) *installerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockEksdInstallerClient(ctrl)
	reader := m.NewMockReader(ctrl)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.EksD.Name = "eks-d-1"
	})

	return &installerTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		client:      client,
		reader:      reader,
		eksd:        eksd.NewEksdInstaller(client, reader),
		clusterSpec: clusterSpec,
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestInstallEksdCRDsSuccess(t *testing.T) {
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	err := os.Unsetenv(features.CloudStackProviderEnvVar)
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)
	if err != nil {
		return
	}

	tt := newInstallerTest(t)
	tt.clusterSpec.VersionsBundle.EksD.Components = "testdata/testdata.yaml"

	tt.reader.EXPECT().ReadFile(tt.clusterSpec.VersionsBundle.EksD.Components).Return([]byte("test data"), nil)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	if err := tt.eksd.InstallEksdCRDs(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("Eksd.InstallEksdCRDs() error = %v, wantErr nil", err)
	}
}

func TestInstallEksdManifestSuccess(t *testing.T) {
	oldCloudstackProviderFeatureValue := os.Getenv(features.CloudStackProviderEnvVar)
	err := os.Unsetenv(features.CloudStackProviderEnvVar)
	defer os.Setenv(features.CloudStackProviderEnvVar, oldCloudstackProviderFeatureValue)
	if err != nil {
		return
	}

	tt := newInstallerTest(t)
	tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl = "testdata/testdata.yaml"

	tt.reader.EXPECT().ReadFile(tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl).Return([]byte("test data"), nil)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	if err := tt.eksd.InstallEksdManifest(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("Eksd.InstallEksdManifest() error = %v, wantErr nil", err)
	}
}

func TestInstallEksdManifestErrorReadingManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl = "fake.yaml"

	tt.reader.EXPECT().ReadFile(tt.clusterSpec.VersionsBundle.EksD.EksDReleaseUrl).Return([]byte(""), fmt.Errorf("error"))
	if err := tt.eksd.InstallEksdManifest(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("Eksd.InstallEksdManifest() error = nil, wantErr not nil")
	}
}
