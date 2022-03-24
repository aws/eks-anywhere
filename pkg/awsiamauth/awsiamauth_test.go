package awsiamauth_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/cluster"
	mocks "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	bundlev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	wantManifestContent   = "testdata/want-aws-iam-authenticator.yaml"
	wantSecretContent     = "testdata/want-aws-iam-authenticator-ca-secret.yaml"
	wantKubeconfigContent = "testdata/want-aws-iam-authenticator-kubeconfig.yaml"
)

func TestGenerateManifestSuccess(t *testing.T) {
	s := givenClusterSpec()

	awsIamAuth, _ := newAwsIamAuth(t)

	gotFileContent, err := awsIamAuth.GenerateManifest(s)
	if err != nil {
		t.Fatalf("awsiamauth.GenerateManifest()\n error = %v\n wantErr = nil", err)
	}
	test.AssertContentToFile(t, string(gotFileContent), wantManifestContent)
}

func TestGenerateCertKeyPairSecretSuccess(t *testing.T) {
	awsIamAuth, mockCertgen := newAwsIamAuth(t)

	mockCertgen.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return([]byte{}, []byte{}, nil)

	gotFileContent, err := awsIamAuth.GenerateCertKeyPairSecret()
	if err != nil {
		t.Fatalf("awsiamauth.GenerateCertKeyPairSecret()\n error = %v\n wantErr = nil", err)
	}
	test.AssertContentToFile(t, string(gotFileContent), wantSecretContent)
}

func TestGenerateCertKeyPairSecretFail(t *testing.T) {
	certGenErr := fmt.Errorf("cert gen error")
	wantErr := fmt.Errorf("generating aws-iam-authenticator cert key pair secret: cert gen error")
	awsIamAuth, mockCertgen := newAwsIamAuth(t)

	mockCertgen.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return(nil, nil, certGenErr)

	_, err := awsIamAuth.GenerateCertKeyPairSecret()
	if !reflect.DeepEqual(err, wantErr) {
		t.Fatalf("error = %v\n wantErr = %v", err, wantErr)
	}
}

func TestGenerateAwsIamAuthKubeconfigSuccess(t *testing.T) {
	s := givenClusterSpec()
	serverUrl := "0.0.0.0:0000"
	tlsCrt := "test-ca"

	awsIamAuth, _ := newAwsIamAuth(t)
	gotFileContent, err := awsIamAuth.GenerateAwsIamAuthKubeconfig(s, serverUrl, tlsCrt)
	if err != nil {
		t.Fatalf("awsiamauth.GenerateAwsIamAuthKubeconfig()\n error = %v\n wantErr = nil", err)
	}
	test.AssertContentToFile(t, string(gotFileContent), wantKubeconfigContent)
}

func newAwsIamAuth(t *testing.T) (*awsiamauth.AwsIamAuth, *mocks.MockCertificateGenerator) {
	mockCtrl := gomock.NewController(t)
	mockCertgen := mocks.NewMockCertificateGenerator(mockCtrl)
	mockClusterId := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")
	return awsiamauth.NewAwsIamAuth(mockCertgen, mockClusterId), mockCertgen
}

func givenClusterSpec() *cluster.Spec {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.VersionsBundle.KubeDistro.AwsIamAuthImage = bundlev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-sigs/aws-iam-authenticator:v0.5.2-eks-1-18-11",
		}
		s.AWSIamConfig = &v1alpha1.AWSIamConfig{
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
		}
	})
	return clusterSpec
}
