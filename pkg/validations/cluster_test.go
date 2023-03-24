package validations_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type clusterTest struct {
	*WithT
	tlsValidator *mocks.MockTlsValidator
	kubectl      *mocks.MockKubectlClient
	clusterSpec  *cluster.Spec
	certContent  string
	host, port   string
}

type clusterTestOpt func(t *testing.T, ct *clusterTest)

func newTest(t *testing.T, opts ...clusterTestOpt) *clusterTest {
	cTest := &clusterTest{
		WithT:       NewWithT(t),
		clusterSpec: test.NewClusterSpec(),
	}
	for _, opt := range opts {
		opt(t, cTest)
	}
	return cTest
}

func withTLS() clusterTestOpt {
	return func(t *testing.T, ct *clusterTest) {
		ctrl := gomock.NewController(t)
		host := "https://host.h"
		port := "1111"
		ct.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Endpoint: host,
			Port:     port,
		}
		ct.tlsValidator = mocks.NewMockTlsValidator(ctrl)
		ct.certContent = "content"
		ct.host = host
		ct.port = port
	}
}

func withKubectl() clusterTestOpt {
	return func(t *testing.T, ct *clusterTest) {
		ctrl := gomock.NewController(t)
		ct.kubectl = mocks.NewMockKubectlClient(ctrl)
	}
}

func TestValidateCertForRegistryMirrorNoRegistryMirror(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = nil

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorCertInvalid(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent = tt.certContent
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(errors.New("invalid cert"))

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("invalid registry certificate: invalid cert")),
	)
}

func TestValidateCertForRegistryMirrorCertValid(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.CACertContent = tt.certContent
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)
	tt.tlsValidator.EXPECT().ValidateCert(tt.host, tt.port, tt.certContent).Return(nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorNoCertIsSignedByKnownAuthority(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(false, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateCertForRegistryMirrorIsSignedByUnknownAuthority(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.tlsValidator.EXPECT().IsSignedByUnknownAuthority(tt.host, tt.port).Return(true, nil)

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(
		MatchError(ContainSubstring("registry https://host.h is using self-signed certs, please provide the certificate using caCertContent field")),
	)
}

func TestValidateCertForRegistryMirrorInsecureSkip(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify = true

	tt.Expect(validations.ValidateCertForRegistryMirror(tt.clusterSpec, tt.tlsValidator)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorNoRegistryMirror(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = nil

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorNoAuth(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate = false

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateAuthenticationForRegistryMirrorAuthInvalid(t *testing.T) {
	tt := newTest(t, withTLS())
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
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate = true
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")

	tt.Expect(validations.ValidateAuthenticationForRegistryMirror(tt.clusterSpec)).To(Succeed())
}

func TestValidateManagementClusterNameValid(t *testing.T) {
	mgmtName := "test"
	tt := newTest(t, withKubectl())
	tt.clusterSpec.Cluster.Spec.ManagementCluster.Name = mgmtName

	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, managementCluster(mgmtName), mgmtName).Return(anywhereCluster(mgmtName), nil)

	tt.Expect(validations.ValidateManagementClusterName(ctx, tt.kubectl, managementCluster(mgmtName), mgmtName)).To(Succeed())
}

func TestValidateManagementClusterNameEmptyValid(t *testing.T) {
	mgmtName := "test"
	tt := newTest(t, withKubectl())
	tt.clusterSpec.Cluster.Spec.ManagementCluster.Name = mgmtName

	ctx := context.Background()
	mgmtCluster := anywhereCluster(mgmtName)
	mgmtCluster.Spec.ManagementCluster.Name = ""
	tt.kubectl.EXPECT().GetEksaCluster(ctx, managementCluster(mgmtName), mgmtName).Return(anywhereCluster(mgmtName), nil)

	tt.Expect(validations.ValidateManagementClusterName(ctx, tt.kubectl, managementCluster(mgmtName), mgmtName)).To(Succeed())
}

func TestValidateManagementClusterNameNotExist(t *testing.T) {
	mgmtName := "test"
	tt := newTest(t, withKubectl())
	tt.clusterSpec.Cluster.Spec.ManagementCluster.Name = mgmtName

	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, managementCluster(mgmtName), mgmtName).Return(nil, errors.New("test"))

	tt.Expect(validations.ValidateManagementClusterName(ctx, tt.kubectl, managementCluster(mgmtName), mgmtName)).NotTo(Succeed())
}

func TestValidateManagementClusterNameWorkload(t *testing.T) {
	mgmtName := "test"
	tt := newTest(t, withKubectl())
	tt.clusterSpec.Cluster.Spec.ManagementCluster.Name = mgmtName

	ctx := context.Background()
	eksaCluster := anywhereCluster(mgmtName)
	eksaCluster.Spec.ManagementCluster.Name = "mgmtCluster"
	tt.kubectl.EXPECT().GetEksaCluster(ctx, managementCluster(mgmtName), mgmtName).Return(eksaCluster, nil)

	tt.Expect(validations.ValidateManagementClusterName(ctx, tt.kubectl, managementCluster(mgmtName), mgmtName)).NotTo(Succeed())
}

func managementCluster(name string) *types.Cluster {
	return &types.Cluster{
		Name: name,
	}
}

func anywhereCluster(name string) *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: anywherev1.ClusterSpec{
			ManagementCluster: anywherev1.ManagementCluster{
				Name: name,
			},
		},
	}
}

func TestValidateK8s126Support(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube126
	tt.Expect(validations.ValidateK8s126Support(tt.clusterSpec)).To(
		MatchError(ContainSubstring("kubernetes version 1.26 is not enabled. Please set the env variable K8S_1_26_SUPPORT")))
}

func TestValidateK8s126SupportActive(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube126
	features.ClearCache()
	os.Setenv(features.K8s126SupportEnvVar, "true")
	tt.Expect(validations.ValidateK8s126Support(tt.clusterSpec)).To(Succeed())
}
