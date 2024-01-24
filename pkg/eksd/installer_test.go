package eksd_test

import (
	"context"
	"errors"
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
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var testdataFile = "testdata/testdata.yaml"

type installerTest struct {
	*WithT
	ctx           context.Context
	client        *mocks.MockEksdInstallerClient
	reader        *m.MockReader
	clusterSpec   *cluster.Spec
	eksdInstaller *eksd.Installer
	cluster       *types.Cluster
}

func newInstallerTest(t *testing.T) *installerTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockEksdInstallerClient(ctrl)
	reader := m.NewMockReader(ctrl)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundles["1.19"].EksD.Name = "eks-d-1"
	})

	return &installerTest{
		WithT:         NewWithT(t),
		ctx:           context.Background(),
		client:        client,
		reader:        reader,
		eksdInstaller: eksd.NewEksdInstaller(client, reader),
		clusterSpec:   clusterSpec,
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestInstallEksdCRDsSuccess(t *testing.T) {
	tt := newInstallerTest(t)
	tt.clusterSpec.Bundles = bundle()

	tt.reader.EXPECT().ReadFile(testdataFile).Return([]byte("test data"), nil).Times(1)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil)
	if err := tt.eksdInstaller.InstallEksdCRDs(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("Eksd.InstallEksdCRDs() error = %v, wantErr nil", err)
	}
}

func TestInstallEksdManifestSuccess(t *testing.T) {
	tt := newInstallerTest(t)
	tt.eksdInstaller = eksd.NewEksdInstaller(tt.client, tt.reader, eksd.WithRetrier(retrier.NewWithMaxRetries(3, 0)))
	tt.clusterSpec.Bundles = bundle()

	tt.reader.EXPECT().ReadFile(testdataFile).Return([]byte("test data"), nil).Times(2)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(errors.New("error apply")).Times(2)
	tt.client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(tt.ctx, tt.cluster, []byte("test data"), constants.EksaSystemNamespace).Return(nil).Times(2)
	if err := tt.eksdInstaller.InstallEksdManifest(tt.ctx, tt.clusterSpec, tt.cluster); err != nil {
		t.Errorf("Eksd.InstallEksdManifest() error = %v, wantErr nil", err)
	}
}

func TestInstallEksdManifestErrorReadingManifest(t *testing.T) {
	tt := newInstallerTest(t)
	tt.eksdInstaller.SetRetrier(retrier.NewWithMaxRetries(1, 0))
	tt.clusterSpec.Bundles = bundle()
	tt.clusterSpec.Bundles.Spec.VersionsBundles[0].EksD.EksDReleaseUrl = "fake.yaml"

	tt.reader.EXPECT().ReadFile(tt.clusterSpec.Bundles.Spec.VersionsBundles[0].EksD.EksDReleaseUrl).Return([]byte(""), fmt.Errorf("error"))
	if err := tt.eksdInstaller.InstallEksdManifest(tt.ctx, tt.clusterSpec, tt.cluster); err == nil {
		t.Error("Eksd.InstallEksdManifest() error = nil, wantErr not nil")
	}
}

func bundle() *v1alpha1.Bundles {
	return &v1alpha1.Bundles{
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					EksD: v1alpha1.EksDRelease{
						Components:     testdataFile,
						EksDReleaseUrl: testdataFile,
					},
				},
				{
					EksD: v1alpha1.EksDRelease{
						Components:     testdataFile,
						EksDReleaseUrl: testdataFile,
					},
				},
			},
		},
	}
}
