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
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type clusterTest struct {
	*WithT
	tlsValidator *mocks.MockTlsValidator
	kubectl      *mocks.MockKubectlClient
	provider     *providermocks.MockProvider
	clusterSpec  *cluster.Spec
	certContent  string
	host, port   string
}

type clusterTestOpt func(t *testing.T, ct *clusterTest)

func newTest(t *testing.T, opts ...clusterTestOpt) *clusterTest {
	ctrl := gomock.NewController(t)
	cTest := &clusterTest{
		WithT:       NewWithT(t),
		clusterSpec: test.NewClusterSpec(),
		provider:    providermocks.NewMockProvider(ctrl),
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

func TestValidateOSForRegistryMirrorNoRegistryMirror(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = nil
	tt.Expect(validations.ValidateOSForRegistryMirror(tt.clusterSpec, tt.provider)).To(Succeed())
}

func TestValidateOSForRegistryMirrorInsecureSkipVerifyDisabled(t *testing.T) {
	tt := newTest(t, withTLS())
	tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify = false
	tt.provider.EXPECT().MachineConfigs(tt.clusterSpec).Return([]providers.MachineConfig{})
	tt.Expect(validations.ValidateOSForRegistryMirror(tt.clusterSpec, tt.provider)).To(Succeed())
}

func TestValidateOSForRegistryMirrorInsecureSkipVerifyEnabled(t *testing.T) {
	tests := []struct {
		name           string
		mirrorConfig   *anywherev1.RegistryMirrorConfiguration
		machineConfigs func() []providers.MachineConfig
		wantErr        string
	}{
		{
			name: "insecureSkipVerify no machine configs",
			machineConfigs: func() []providers.MachineConfig {
				return nil
			},
			wantErr: "",
		},
		{
			name: "insecureSkipVerify on provider with ubuntu",
			machineConfigs: func() []providers.MachineConfig {
				configs := make([]providers.MachineConfig, 0, 1)
				configs = append(configs, &anywherev1.VSphereMachineConfig{
					Spec: anywherev1.VSphereMachineConfigSpec{
						OSFamily: anywherev1.Ubuntu,
					},
				})
				return configs
			},
			wantErr: "",
		},
		{
			name: "insecureSkipVerify on provider with bottlerocket",
			machineConfigs: func() []providers.MachineConfig {
				configs := make([]providers.MachineConfig, 0, 1)
				configs = append(configs, &anywherev1.SnowMachineConfig{
					Spec: anywherev1.SnowMachineConfigSpec{
						OSFamily: anywherev1.Bottlerocket,
					},
				})
				return configs
			},
			wantErr: "InsecureSkipVerify is not supported for bottlerocket",
		},
		{
			name: "insecureSkipVerify on provider with redhat",
			machineConfigs: func() []providers.MachineConfig {
				configs := make([]providers.MachineConfig, 0, 1)
				configs = append(configs, &anywherev1.VSphereMachineConfig{
					Spec: anywherev1.VSphereMachineConfigSpec{
						OSFamily: anywherev1.RedHat,
					},
				})
				return configs
			},
			wantErr: "",
		},
	}

	validationTest := newTest(t, func(t *testing.T, ct *clusterTest) {
		ct.clusterSpec = test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
				InsecureSkipVerify: true,
			}
		})
	})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validationTest.provider.EXPECT().MachineConfigs(validationTest.clusterSpec).Return(test.machineConfigs())
			err := validations.ValidateOSForRegistryMirror(validationTest.clusterSpec, validationTest.provider)
			if test.wantErr != "" {
				validationTest.Expect(err).To(MatchError(test.wantErr))
			} else {
				validationTest.Expect(err).To(BeNil())
			}
		})
	}
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

func TestValidateManagementClusterBundlesVersion(t *testing.T) {
	type testParam struct {
		mgmtBundlesName   string
		mgmtBundlesNumber int
		wkBundlesName     string
		wkBundlesNumber   int
		wantErr           string
		errGetEksaCluster error
		errGetBundles     error
	}

	testParams := []testParam{
		{
			mgmtBundlesName:   "bundles-28",
			mgmtBundlesNumber: 28,
			wkBundlesName:     "bundles-27",
			wkBundlesNumber:   27,
			wantErr:           "",
		},
		{
			mgmtBundlesName:   "bundles-28",
			mgmtBundlesNumber: 28,
			wkBundlesName:     "bundles-29",
			wkBundlesNumber:   29,
			wantErr:           "cannot upgrade workload cluster with bundle spec.number 29 while management cluster management-cluster is on older bundle spec.number 28",
		},
		{
			mgmtBundlesName:   "bundles-28",
			mgmtBundlesNumber: 28,
			wkBundlesName:     "bundles-27",
			wkBundlesNumber:   27,
			wantErr:           "failed to reach cluster",
			errGetEksaCluster: errors.New("failed to reach cluster"),
		},
		{
			mgmtBundlesName:   "bundles-28",
			mgmtBundlesNumber: 28,
			wkBundlesName:     "bundles-27",
			wkBundlesNumber:   27,
			wantErr:           "failed to reach cluster",
			errGetBundles:     errors.New("failed to reach cluster"),
		},
	}

	for _, p := range testParams {
		tt := newTest(t, withKubectl())
		mgmtName := "management-cluster"
		mgmtCluster := managementCluster(mgmtName)
		mgmtClusterObject := anywhereCluster(mgmtName)

		mgmtClusterObject.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:      p.mgmtBundlesName,
			Namespace: constants.EksaSystemNamespace,
		}

		tt.clusterSpec.Config.Cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:      p.wkBundlesName,
			Namespace: constants.EksaSystemNamespace,
		}
		wkBundle := &releasev1alpha1.Bundles{
			Spec: releasev1alpha1.BundlesSpec{
				Number: p.wkBundlesNumber,
			},
		}
		tt.clusterSpec.Bundles = wkBundle

		mgmtBundle := &releasev1alpha1.Bundles{
			Spec: releasev1alpha1.BundlesSpec{
				Number: p.mgmtBundlesNumber,
			},
		}

		ctx := context.Background()
		tt.kubectl.EXPECT().GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name).Return(mgmtClusterObject, p.errGetEksaCluster)
		if p.errGetEksaCluster == nil {
			tt.kubectl.EXPECT().GetBundles(ctx, mgmtCluster.KubeconfigFile, mgmtClusterObject.Spec.BundlesRef.Name, mgmtClusterObject.Spec.BundlesRef.Namespace).Return(mgmtBundle, p.errGetBundles)
		}

		if p.wantErr == "" {
			err := validations.ValidateManagementClusterBundlesVersion(ctx, tt.kubectl, mgmtCluster, tt.clusterSpec)
			tt.Expect(err).To(BeNil())
		} else {
			err := validations.ValidateManagementClusterBundlesVersion(ctx, tt.kubectl, mgmtCluster, tt.clusterSpec)
			tt.Expect(err.Error()).To(Equal(p.wantErr))
		}
	}
}

func TestValidateManagementClusterBundlesVersionMissingBundlesRef(t *testing.T) {
	tt := newTest(t, withKubectl())
	wantErr := "management cluster bundlesRef cannot be nil"
	mgmtName := "management-cluster"
	mgmtCluster := managementCluster(mgmtName)
	mgmtClusterObject := anywhereCluster(mgmtName)

	mgmtClusterObject.Spec.BundlesRef = nil
	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name).Return(mgmtClusterObject, nil)

	err := validations.ValidateManagementClusterBundlesVersion(ctx, tt.kubectl, mgmtCluster, tt.clusterSpec)
	tt.Expect(err.Error()).To(Equal(wantErr))
}
