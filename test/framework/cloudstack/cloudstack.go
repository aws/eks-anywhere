package cloudstack

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api/cloudstack"
	"github.com/aws/eks-anywhere/test/framework"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	cloudstackDomainVar                = "T_CLOUDSTACK_DOMAIN"
	cloudstackZoneVar                  = "T_CLOUDSTACK_ZONE"
	cloudstackAccountVar               = "T_CLOUDSTACK_ACCOUNT"
	cloudstackNetworkVar               = "T_CLOUDSTACK_NETWORK"
	cloudstackManagementServerVar      = "T_CLOUDSTACK_MANAGEMENT_SERVER"
	cloudstackSshAuthorizedKeyVar      = "T_CLOUDSTACK_SSH_AUTHORIZED_KEY"
	cloudstackTemplateRedhat120Var     = "T_CLOUDSTACK_TEMPLATE_REDHAT_1_20"
	cloudstackTemplateRedhat121Var     = "T_CLOUDSTACK_TEMPLATE_REDHAT_1_21"
	cloudstackComputeOfferingLargeVar  = "T_CLOUDSTACK_COMPUTE_OFFERING_LARGE"
	cloudstackComputeOfferingLargerVar = "T_CLOUDSTACK_COMPUTE_OFFERING_LARGER"
	cloudstackTlsInsecureVar           = "T_CLOUDSTACK_TLS_INSECURE"
	cloudstackK8sControlPlaneHostVar   = "T_CLOUDSTACK_CONTROL_PLANE_HOST"
	cloudstackB64SecretVar             = "T_CLOUDSTACK_B64ENCODED_SECRET"
	podCidrVar                         = "T_CLOUDSTACK_POD_CIDR"
	serviceCidrVar                     = "T_CLOUDSTACK_SERVICE_CIDR"
)

var requiredEnvVars = []string{
	cloudstackDomainVar,
	cloudstackZoneVar,
	cloudstackAccountVar,
	cloudstackNetworkVar,
	cloudstackManagementServerVar,
	cloudstackSshAuthorizedKeyVar,
	cloudstackTemplateRedhat120Var,
	cloudstackTemplateRedhat121Var,
	cloudstackComputeOfferingLargeVar,
	cloudstackComputeOfferingLargerVar,
	cloudstackTlsInsecureVar,
	cloudstackK8sControlPlaneHostVar,
	cloudstackB64SecretVar,
	podCidrVar,
	serviceCidrVar,
}

type CloudStack struct {
	t           *testing.T
	fillers     []cloudstack.CloudStackFiller
	podCidr     string
	serviceCidr string
}

type CloudStackOpt func(*CloudStack)

func UpdateRedhatTemplate120Var() cloudstack.CloudStackFiller {
	return cloudstack.WithStringFromEnvVar(cloudstackTemplateRedhat120Var, cloudstack.WithTemplate)
}

func UpdateRedhatTemplate121Var() cloudstack.CloudStackFiller {
	return cloudstack.WithStringFromEnvVar(cloudstackTemplateRedhat121Var, cloudstack.WithTemplate)
}

func NewCloudStack(t *testing.T, opts ...CloudStackOpt) *CloudStack {
	framework.CheckRequiredEnvVars(t, requiredEnvVars)
	v := &CloudStack{
		t: t,
		fillers: []cloudstack.CloudStackFiller{
			cloudstack.WithStringFromEnvVar(cloudstackDomainVar, cloudstack.WithDomain),
			cloudstack.WithStringFromEnvVar(cloudstackZoneVar, cloudstack.WithZone),
			cloudstack.WithStringFromEnvVar(cloudstackAccountVar, cloudstack.WithAccount),
			cloudstack.WithStringFromEnvVar(cloudstackNetworkVar, cloudstack.WithNetwork),
			cloudstack.WithStringFromEnvVar(cloudstackSshAuthorizedKeyVar, cloudstack.WithSSHAuthorizedKey),
			cloudstack.WithStringFromEnvVar(cloudstackTemplateRedhat120Var, cloudstack.WithTemplate),
			cloudstack.WithStringFromEnvVar(cloudstackComputeOfferingLargeVar, cloudstack.WithComputeOffering),
			cloudstack.WithBoolFromEnvVar(cloudstackTlsInsecureVar, cloudstack.WithTLSInsecure),
		},
	}

	v.podCidr = os.Getenv(podCidrVar)
	v.serviceCidr = os.Getenv(serviceCidrVar)

	for _, opt := range opts {
		opt(v)
	}

	return v
}

func WithRedhat121() CloudStackOpt {
	return func(v *CloudStack) {
		v.fillers = append(v.fillers,
			cloudstack.WithStringFromEnvVar(cloudstackTemplateRedhat121Var, cloudstack.WithTemplate),
			cloudstack.WithOsFamily(v1alpha1.Redhat),
		)
	}
}

func WithRedhat120() CloudStackOpt {
	return func(v *CloudStack) {
		v.fillers = append(v.fillers,
			cloudstack.WithStringFromEnvVar(cloudstackTemplateRedhat120Var, cloudstack.WithTemplate),
			cloudstack.WithOsFamily(v1alpha1.Redhat),
		)
	}
}

func WithCloudStackFillers(fillers ...cloudstack.CloudStackFiller) CloudStackOpt {
	return func(v *CloudStack) {
		v.fillers = append(v.fillers, fillers...)
	}
}

func (v *CloudStack) Name() string {
	return "cloudstack"
}

func (v *CloudStack) Setup() {}

func (v *CloudStack) CustomizeProviderConfig(file string) []byte {
	return v.customizeProviderConfig(file, v.fillers...)
}

func (v *CloudStack) customizeProviderConfig(file string, fillers ...cloudstack.CloudStackFiller) []byte {
	providerOutput, err := cloudstack.AutoFillCloudStackProvider(file, fillers...)
	if err != nil {
		v.t.Fatalf("Error customizing provider config from file: %v", err)
	}
	return providerOutput
}

func (v *CloudStack) WithProviderUpgrade(fillers ...cloudstack.CloudStackFiller) framework.ClusterE2ETestOpt {
	return func(e *framework.ClusterE2ETest) {
		e.ProviderConfigB = v.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (v *CloudStack) ClusterConfigFillers() []api.ClusterFiller {
	return []api.ClusterFiller{
		api.WithPodCidr(os.Getenv(podCidrVar)),
		api.WithServiceCidr(os.Getenv(serviceCidrVar)),
		api.WithControlPlaneCount(1),
		api.WithControlPlaneEndpointIP(os.Getenv(cloudstackK8sControlPlaneHostVar)),
	}
}

func RequiredCloudstackEnvVars() []string {
	return requiredEnvVars
}
