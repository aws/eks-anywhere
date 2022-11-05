package validations_test

import (
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type tlsTest struct {
	*WithT
	tlsValidator *mocks.MockTlsValidator
	provider     *providermocks.MockProvider
	clusterSpec  *cluster.Spec
	certContent  string
	host, port   string
}

func newTlsTest(t *testing.T) *tlsTest {
	ctrl := gomock.NewController(t)
	host := "https://host.h"
	port := "1111"
	return &tlsTest{
		WithT:        NewWithT(t),
		tlsValidator: mocks.NewMockTlsValidator(ctrl),
		provider:     providermocks.NewMockProvider(ctrl),
		clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
				Endpoint: host,
				Port:     port,
			}
		}),
		certContent: "content",
		host:        host,
		port:        port,
	}
}

func TestValidateCertForRegistryMirrorNoRegistryMirror(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = nil

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorCertInvalid(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent = tt.certContent
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(errors.New("invalid cert"))

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("invalid registry certificate: invalid cert")),
	)
}

func TestValidateCertForRegistryMirrorCertValid(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent = tt.certContent
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorNoCertIsSignedByKnownAuthority(t *testing.T) {
	tt := newTlsTest(t)
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorIsSignedByUnknownAuthority(t *testing.T) {
	tt := newTlsTest(t)
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(true, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("registry https://host.h is using self-signed certs, please provide the certificate using caCertContent field")),
	)
}

func TestValidateCertForRegistryMirrorInsecureSkip(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify = true

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorNoRegistryMirror(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = nil

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorNoAuth(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate = false

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorAuthInvalid(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate = true
	if err := os.Unsetenv("REGISTRY_USERNAME"); err != nil {
		t.Fatalf(err.Error())
	}
	if err := os.Unsetenv("REGISTRY_PASSWORD"); err != nil {
		t.Fatalf(err.Error())
	}

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(
		MatchError(ContainSubstring("please set REGISTRY_USERNAME env var")))
}

func TestValidateAuthenticationForRegistryMirrorAuthValid(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate = true
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateK8s124Support(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube124
	tt.Expect(validations.ValidateK8s124Support(tt.clusterSpec)).To(
		MatchError(ContainSubstring("kubernetes version 1.24 is not enabled. Please set the env variable K8S_1_24_SUPPORT")))
}

func TestValidateK8s124SupportActive(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube124
	features.ClearCache()
	os.Setenv(features.K8s124SupportEnvVar, "true")
	tt.Expect(validations.ValidateK8s124Support(tt.clusterSpec)).To(Succeed())
}
