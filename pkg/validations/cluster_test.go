package validations_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type tlsTest struct {
	*WithT
	tlsValidator *mocks.MockTlsValidator
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
	tt.tlsValidator.EXPECT().HasSelfSignedCert(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(errors.New("invalid cert"))

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("invalid registry certificate: invalid cert")),
	)
}

func TestValidateCertForRegistryMirrorCertValid(t *testing.T) {
	tt := newTlsTest(t)
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent = tt.certContent
	tt.tlsValidator.EXPECT().HasSelfSignedCert(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorNoCertNoSelfSigned(t *testing.T) {
	tt := newTlsTest(t)
	tt.tlsValidator.EXPECT().HasSelfSignedCert(tt.host, tt.port).Return(false, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorNoCertSelfSigned(t *testing.T) {
	tt := newTlsTest(t)
	tt.tlsValidator.EXPECT().HasSelfSignedCert(tt.host, tt.port).Return(true, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("registry https://host.h is using self-signed certs, please provide the certificate using caCertContent field")),
	)
}
