package awsiamauth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/cluster"
	cryptomocks "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermock "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestInstallAWSIAMAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	clusterID := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")

	var manifest []byte
	k8s := NewMockKubernetesClient(ctrl)
	k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, cluster *types.Cluster, data []byte) error {
			manifest = data
			return nil
		},
	)
	k8s.EXPECT().GetAPIServerURL(gomock.Any(), gomock.Any()).Return("api-server-url", nil)
	k8s.EXPECT().GetClusterCACert(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("ca-cert"), nil)

	var kubeconfig []byte
	writer := filewritermock.NewMockFileWriter(ctrl)
	writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(fileName string, content []byte, f ...filewriter.FileOptionsFunc) (string, error) {
			kubeconfig = content
			return "some file", nil
		},
	)

	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube124,
				},
			},
		},
		VersionsBundles: test.VersionsBundlesMap(),
		AWSIamConfig: &v1alpha1.AWSIamConfig{
			Spec: v1alpha1.AWSIamConfigSpec{
				AWSRegion:   "test-region",
				BackendMode: []string{"mode1", "mode2"},
				MapRoles: []v1alpha1.MapRoles{
					{
						RoleARN:  "test-role-arn",
						Username: "test",
						Groups:   []string{"group1", "group2"},
					},
				},
				MapUsers: []v1alpha1.MapUsers{
					{
						UserARN:  "test-user-arn",
						Username: "test",
						Groups:   []string{"group1", "group2"},
					},
				},
				Partition: "test",
			},
		},
	}

	installer := awsiamauth.NewInstaller(certs, clusterID, k8s, writer)

	err := installer.InstallAWSIAMAuth(context.Background(), &types.Cluster{}, &types.Cluster{}, spec)
	if err != nil {
		t.Fatal(err)
	}

	test.AssertContentToFile(t, string(kubeconfig), "testdata/InstallAWSIAMAuth-kubeconfig.yaml")
	test.AssertContentToFile(t, string(manifest), "testdata/InstallAWSIAMAuth-manifest.yaml")
}

func TestInstallAWSIAMAuthErrors(t *testing.T) {
	cases := []struct {
		Name           string
		ConfigureMocks func(err error, k8s *MockKubernetesClient, writer *filewritermock.MockFileWriter)
	}{
		{
			Name: "ApplyFails",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, writer *filewritermock.MockFileWriter) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
			},
		},
		{
			Name: "GetAPIServerURLFails",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, writer *filewritermock.MockFileWriter) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				k8s.EXPECT().GetAPIServerURL(gomock.Any(), gomock.Any()).Return("", err)
			},
		},
		{
			Name: "GetClusterCACertFails",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, writer *filewritermock.MockFileWriter) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				k8s.EXPECT().GetAPIServerURL(gomock.Any(), gomock.Any()).Return("api-server-url", nil)
				k8s.EXPECT().GetClusterCACert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, err)
			},
		},
		{
			Name: "WriteFails",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, writer *filewritermock.MockFileWriter) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				k8s.EXPECT().GetAPIServerURL(gomock.Any(), gomock.Any()).Return("api-server-url", nil)
				k8s.EXPECT().GetClusterCACert(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("ca-cert"), nil)
				writer.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).Return("", err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			certs := cryptomocks.NewMockCertificateGenerator(ctrl)
			clusterID := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")

			k8s := NewMockKubernetesClient(ctrl)
			writer := filewritermock.NewMockFileWriter(ctrl)
			tc.ConfigureMocks(errors.New(tc.Name), k8s, writer)

			spec := &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &v1alpha1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster",
						},
						Spec: v1alpha1.ClusterSpec{
							KubernetesVersion: v1alpha1.Kube124,
						},
					},
				},
				VersionsBundles: test.VersionsBundlesMap(),
				AWSIamConfig: &v1alpha1.AWSIamConfig{
					Spec: v1alpha1.AWSIamConfigSpec{
						AWSRegion:   "test-region",
						BackendMode: []string{"mode1", "mode2"},
						MapRoles: []v1alpha1.MapRoles{
							{
								RoleARN:  "test-role-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						MapUsers: []v1alpha1.MapUsers{
							{
								UserARN:  "test-user-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						Partition: "test",
					},
				},
			}

			installer := awsiamauth.NewInstaller(certs, clusterID, k8s, writer)

			err := installer.InstallAWSIAMAuth(context.Background(), &types.Cluster{}, &types.Cluster{}, spec)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.Name) {
				t.Fatalf("expected error to contain %q, got %q", tc.Name, err.Error())
			}
		})
	}
}

func TestCreateAndInstallAWSIAMAuthCASecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	clusterID := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")

	writer := filewritermock.NewMockFileWriter(ctrl)

	var manifest []byte
	k8s := NewMockKubernetesClient(ctrl)
	k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, cluster *types.Cluster, data []byte) error {
			manifest = data
			return nil
		},
	)

	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	certs.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return([]byte("ca-cert"), []byte("ca-key"), nil)

	installer := awsiamauth.NewInstaller(certs, clusterID, k8s, writer)

	err := installer.CreateAndInstallAWSIAMAuthCASecret(context.Background(), &types.Cluster{}, "test-cluster")
	if err != nil {
		t.Fatal(err)
	}

	test.AssertContentToFile(t, string(manifest), "testdata/CreateAndInstallAWSIAMAuthCASecret-manifest.yaml")
}

func TestCreateAndInstallAWSIAMAuthCASecretErrors(t *testing.T) {
	cases := []struct {
		Name           string
		ConfigureMocks func(err error, k8s *MockKubernetesClient, certs *cryptomocks.MockCertificateGenerator)
	}{
		{
			Name: "ApplyError",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, certs *cryptomocks.MockCertificateGenerator) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
			},
		},
		{
			Name: "GenerateIamAuthSelfSignCertKeyPairError",
			ConfigureMocks: func(err error, k8s *MockKubernetesClient, certs *cryptomocks.MockCertificateGenerator) {
				k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				certs.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return(nil, nil, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			clusterID := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")

			writer := filewritermock.NewMockFileWriter(ctrl)

			k8s := NewMockKubernetesClient(ctrl)
			k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(tc.Name))

			certs := cryptomocks.NewMockCertificateGenerator(ctrl)
			certs.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return([]byte("ca-cert"), []byte("ca-key"), nil)

			installer := awsiamauth.NewInstaller(certs, clusterID, k8s, writer)

			err := installer.CreateAndInstallAWSIAMAuthCASecret(context.Background(), &types.Cluster{}, "test-cluster")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.Name) {
				t.Fatalf("expected error to contain %q, got %q", tc.Name, err.Error())
			}
		})
	}
}

func TestUpgradeAWSIAMAuth(t *testing.T) {
	clusterID := uuid.Nil

	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	writer := filewritermock.NewMockFileWriter(ctrl)

	k8s := NewMockKubernetesClient(ctrl)

	var manifest []byte
	k8s.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, cluster *types.Cluster, data []byte) error {
			manifest = data
			return nil
		},
	)

	installer := awsiamauth.NewInstaller(certs, clusterID, k8s, writer)

	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube123,
				},
			},
		},
		VersionsBundles: test.VersionsBundlesMap(),
		AWSIamConfig: &v1alpha1.AWSIamConfig{
			Spec: v1alpha1.AWSIamConfigSpec{
				AWSRegion:   "test-region",
				BackendMode: []string{"mode1", "mode2"},
				MapRoles: []v1alpha1.MapRoles{
					{
						RoleARN:  "test-role-arn",
						Username: "test",
						Groups:   []string{"group1", "group2"},
					},
				},
				MapUsers: []v1alpha1.MapUsers{
					{
						UserARN:  "test-user-arn",
						Username: "test",
						Groups:   []string{"group1", "group2"},
					},
				},
				Partition: "test",
			},
		},
	}

	err := installer.UpgradeAWSIAMAuth(context.Background(), &types.Cluster{}, spec)
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
	test.AssertContentToFile(t, string(manifest), "testdata/UpgradeAWSIAMAuth-manifest.yaml")
}
