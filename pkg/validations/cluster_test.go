package validations_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
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

func TestValidateOSForRegistryMirrorNoPublicEcrRegistry(t *testing.T) {
	tt := newTest(t, withTLS())
	tests := []struct {
		name         string
		mirrorConfig *anywherev1.RegistryMirrorConfiguration
	}{
		{
			name: "no public ecr registry",
			mirrorConfig: &anywherev1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.4",
				Port:     "123",
				OCINamespaces: []anywherev1.OCINamespace{
					{
						Registry: "docker.io",
					},
				},
			},
		},
		{
			name: "more than one registry",
			mirrorConfig: &anywherev1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.4",
				Port:     "123",
				OCINamespaces: []anywherev1.OCINamespace{
					{
						Registry: "public.ecr.aws",
					},
					{
						Registry: "docker.io",
					},
				},
			},
		},
	}

	tt.provider.EXPECT().MachineConfigs(tt.clusterSpec).Return([]providers.MachineConfig{
		&anywherev1.VSphereMachineConfig{
			Spec: anywherev1.VSphereMachineConfigSpec{
				OSFamily: anywherev1.Bottlerocket,
			},
		},
	}).MaxTimes(2)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tt.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = test.mirrorConfig
			tt.Expect(validations.ValidateOSForRegistryMirror(tt.clusterSpec, tt.provider)).To(MatchError("public.ecr.aws is the only registry supported in ociNamespaces for bottlerocket"))
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

func TestValidateEksaVersion(t *testing.T) {
	v := test.DevEksaVersion()
	badVersion := anywherev1.EksaVersion("invalid")
	tests := []struct {
		name       string
		wantErr    error
		version    *anywherev1.EksaVersion
		cliVersion string
		workload   bool
	}{
		{
			name:       "Success",
			wantErr:    nil,
			version:    &v,
			cliVersion: "v0.0.0-dev",
		},
		{
			name:       "Bad Cluster version",
			wantErr:    fmt.Errorf("parsing cluster eksa version: invalid major version in semver invalid: strconv.ParseInt: parsing \"\": invalid syntax"),
			version:    &badVersion,
			cliVersion: "v0.0.0-dev",
		},
		{
			name:       "Bad CLI version",
			wantErr:    fmt.Errorf("parsing eksa cli version: invalid major version in semver badvalue: strconv.ParseInt: parsing \"\": invalid syntax"),
			version:    &v,
			cliVersion: "badvalue",
		},
		{
			name:       "Mismatch",
			wantErr:    fmt.Errorf("cluster's eksaVersion does not match EKS-Anywhere CLI's version"),
			version:    &v,
			cliVersion: "v0.0.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTest(t, withKubectl())

			tt.clusterSpec.Cluster.Spec.EksaVersion = tc.version
			ctx := context.Background()

			if tc.workload {
				tt.clusterSpec.Cluster.SetManagedBy("other")
			}

			err := validations.ValidateEksaVersion(ctx, tc.cliVersion, tt.clusterSpec)
			if tc.wantErr != nil {
				tt.Expect(err).To(MatchError(tc.wantErr))
			}
		})
	}
}

func TestValidateEksaVersionSkew(t *testing.T) {
	v := anywherev1.EksaVersion("v0.1.0")
	uv := anywherev1.EksaVersion("v0.2.0")
	devVersion := test.DevEksaVersion()
	badVersion := anywherev1.EksaVersion("invalid")
	tests := []struct {
		name           string
		wantErr        error
		version        *anywherev1.EksaVersion
		upgradeVersion *anywherev1.EksaVersion
	}{
		{
			name:           "Success",
			wantErr:        nil,
			version:        &v,
			upgradeVersion: &uv,
		},
		{
			name:           "Bad Cluster version",
			wantErr:        nil,
			version:        &badVersion,
			upgradeVersion: &v,
		},
		{
			name:           "Bad upgrade version",
			wantErr:        fmt.Errorf("spec.EksaVersion: Invalid value: \"invalid\": EksaVersion is not a valid semver"),
			version:        &v,
			upgradeVersion: &badVersion,
		},
		{
			name:           "Fail",
			wantErr:        fmt.Errorf("spec.EksaVersion: Invalid value: \"v0.1.0\": cannot downgrade from v0.2.0 to v0.1.0: EksaVersion upgrades must be incremental"),
			version:        &uv,
			upgradeVersion: &v,
		},
		{
			name:           "Cluster nil",
			wantErr:        nil,
			version:        nil,
			upgradeVersion: &v,
		},
		{
			name:           "Upgrade nil",
			wantErr:        nil,
			version:        &uv,
			upgradeVersion: nil,
		},
		{
			name:           "Upgrade to dev build success",
			wantErr:        nil,
			version:        &uv,
			upgradeVersion: &devVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTest(t, withKubectl())
			mgmtName := "management-cluster"
			mgmtCluster := managementCluster(mgmtName)
			mgmtClusterObject := anywhereCluster(mgmtName)
			mgmtClusterObject.Spec.BundlesRef = &anywherev1.BundlesRef{}
			mgmtClusterObject.Spec.EksaVersion = tc.version
			tt.clusterSpec.Cluster.Spec.EksaVersion = tc.upgradeVersion
			ctx := context.Background()
			tt.kubectl.EXPECT().GetEksaCluster(ctx, mgmtCluster, tt.clusterSpec.Cluster.Name).Return(mgmtClusterObject, nil)
			tt.kubectl.EXPECT().GetBundles(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(test.Bundle(), nil).AnyTimes()

			err := validations.ValidateEksaVersionSkew(ctx, tt.kubectl, mgmtCluster, tt.clusterSpec)
			if err != nil {
				tt.Expect(err.Error()).To(ContainSubstring(tc.wantErr.Error()))
			} else {
				tt.Expect(tc.wantErr).To(BeNil())
			}
		})
	}
}

func TestValidateManagementClusterEksaVersion(t *testing.T) {
	v := test.DevEksaVersion()
	uv := anywherev1.EksaVersion("v0.1.0")
	badVersion := anywherev1.EksaVersion("invalid")

	tests := []struct {
		name              string
		wantErr           error
		version           *anywherev1.EksaVersion
		managementVersion *anywherev1.EksaVersion
	}{
		{
			name:              "Success",
			wantErr:           nil,
			version:           &v,
			managementVersion: &v,
		},
		{
			name:              "Management with dev build version and workload with release version",
			wantErr:           nil,
			version:           &uv,
			managementVersion: &v,
		},
		{
			name:              "Bad workload version",
			wantErr:           fmt.Errorf("parsing workload"),
			version:           &badVersion,
			managementVersion: &v,
		},
		{
			name:              "Bad management version",
			wantErr:           fmt.Errorf("parsing management"),
			version:           &v,
			managementVersion: &badVersion,
		},
		{
			name:              "Fail",
			wantErr:           fmt.Errorf("cannot upgrade workload cluster to"),
			version:           &uv,
			managementVersion: &v,
		},
		{
			name:              "workload nil",
			wantErr:           fmt.Errorf("cluster has nil EksaVersion"),
			version:           nil,
			managementVersion: &v,
		},
		{
			name:              "managment nil",
			wantErr:           fmt.Errorf("management cluster has nil EksaVersion"),
			version:           &uv,
			managementVersion: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTest(t, withKubectl())
			mgmtName := "management-cluster"
			mgmtCluster := managementCluster(mgmtName)
			mgmtClusterObject := anywhereCluster(mgmtName)
			mgmtClusterObject.Spec.BundlesRef = &anywherev1.BundlesRef{}
			mgmtClusterObject.Spec.EksaVersion = tc.managementVersion
			tt.clusterSpec.Cluster.Spec.EksaVersion = tc.version
			ctx := context.Background()
			tt.kubectl.EXPECT().GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name).Return(mgmtClusterObject, nil)
			tt.kubectl.EXPECT().GetBundles(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(test.Bundle(), nil).AnyTimes()

			err := validations.ValidateManagementClusterEksaVersion(ctx, tt.kubectl, mgmtCluster, tt.clusterSpec)
			if err != nil {
				tt.Expect(err.Error()).To(ContainSubstring(tc.wantErr.Error()))
			}
		})
	}
}

func TestValidateK8s129Support(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube129
	tt.Expect(validations.ValidateK8s129Support(tt.clusterSpec)).To(
		MatchError(ContainSubstring("kubernetes version 1.29 is not enabled. Please set the env variable K8S_1_29_SUPPORT")))
}

func TestValidateK8s129SupportActive(t *testing.T) {
	tt := newTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = anywherev1.Kube129
	features.ClearCache()
	os.Setenv(features.K8s129SupportEnvVar, "true")
	tt.Expect(validations.ValidateK8s129Support(tt.clusterSpec)).To(Succeed())
}

func TestValidateEksaReleaseExistOnManagement(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
		},
		{
			name:    "not present",
			wantErr: fmt.Errorf("not found"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTest(t)
			ctx := context.Background()
			objs := []client.Object{}
			if tc.wantErr == nil {
				objs = append(objs, test.EKSARelease())
			}
			fakeClient := test.NewFakeKubeClient(objs...)
			err := validations.ValidateEksaReleaseExistOnManagement(ctx, fakeClient, tt.clusterSpec.Cluster)
			if err != nil {
				tt.Expect(err.Error()).To(ContainSubstring(tc.wantErr.Error()))
			}
		})
	}
}

func TestValidatePauseAnnotation(t *testing.T) {
	mgmtName := "test"
	tests := []struct {
		name       string
		gotCluster *anywherev1.Cluster
		wantErr    error
	}{
		{
			name: "success",
			gotCluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name:        mgmtName,
					Annotations: map[string]string{},
				},
				Spec: anywherev1.ClusterSpec{
					ManagementCluster: anywherev1.ManagementCluster{
						Name: mgmtName,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "success",
			gotCluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name:        mgmtName,
					Annotations: map[string]string{"anywhere.eks.amazonaws.com/paused": "true"},
				},
				Spec: anywherev1.ClusterSpec{
					ManagementCluster: anywherev1.ManagementCluster{
						Name: mgmtName,
					},
				},
			},
			wantErr: fmt.Errorf("cluster cannot be upgraded with paused cluster controller reconciler"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTest(t, withKubectl())
			ctx := context.Background()

			tt.kubectl.EXPECT().GetEksaCluster(ctx, managementCluster(mgmtName), mgmtName).Return(tc.gotCluster, nil)

			err := validations.ValidatePauseAnnotation(ctx, tt.kubectl, managementCluster(mgmtName), mgmtName)
			if err != nil {
				tt.Expect(err.Error()).To(ContainSubstring(tc.wantErr.Error()))
			} else {
				tt.Expect(tc.wantErr).To(BeNil())
			}
		})
	}
}
