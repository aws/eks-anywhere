package framework

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
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
	cloudStackClusterIPPoolEnvVar      = "T_CLOUDSTACK_CLUSTER_IP_POOL"
	podCidrVar                         = "T_CLOUDSTACK_POD_CIDR"
	serviceCidrVar                     = "T_CLOUDSTACK_SERVICE_CIDR"
	cloudstackFeatureGateEnvVar        = "CLOUDSTACK_PROVIDER"
	cloudstackB64EncodedSecretEnvVar   = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"
)

var requiredCloudStackEnvVars = []string{
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
	cloudStackClusterIPPoolEnvVar,
	podCidrVar,
	serviceCidrVar,
	cloudstackFeatureGateEnvVar,
	cloudstackB64EncodedSecretEnvVar,
}

type CloudStack struct {
	t           *testing.T
	fillers     []api.CloudStackFiller
	podCidr     string
	serviceCidr string
}

type CloudStackOpt func(*CloudStack)

func UpdateRedhatTemplate120Var() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat120Var, api.WithCloudStackTemplate)
}

func UpdateRedhatTemplate121Var() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplate)
}

func UpdateLargerCloudStackComputeOffering() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargerVar, api.WithCloudStackComputeOffering)
}

func NewCloudStack(t *testing.T, opts ...CloudStackOpt) *CloudStack {
	checkRequiredEnvVars(t, requiredCloudStackEnvVars)
	v := &CloudStack{
		t: t,
		fillers: []api.CloudStackFiller{
			api.WithCloudStackStringFromEnvVar(cloudstackDomainVar, api.WithCloudStackDomain),
			api.WithCloudStackStringFromEnvVar(cloudstackManagementServerVar, api.WithCloudStackManagementServer),
			api.WithCloudStackStringFromEnvVar(cloudstackZoneVar, api.WithCloudStackZone),
			api.WithCloudStackStringFromEnvVar(cloudstackNetworkVar, api.WithCloudStackNetwork),
			api.WithCloudStackStringFromEnvVar(cloudstackAccountVar, api.WithCloudStackAccount),
			api.WithCloudStackStringFromEnvVar(cloudstackSshAuthorizedKeyVar, api.WithCloudStackSSHAuthorizedKey),
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat120Var, api.WithCloudStackTemplate),
			api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargeVar, api.WithCloudStackComputeOffering),
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
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplate),
		)
	}
}

func WithRedhat120() CloudStackOpt {
	return func(v *CloudStack) {
		v.fillers = append(v.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat120Var, api.WithCloudStackTemplate),
		)
	}
}

func WithCloudStackFillers(fillers ...api.CloudStackFiller) CloudStackOpt {
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

func (v *CloudStack) customizeProviderConfig(file string, fillers ...api.CloudStackFiller) []byte {
	providerOutput, err := api.AutoFillCloudStackProvider(file, fillers...)
	if err != nil {
		v.t.Fatalf("Error customizing provider config from file: %v", err)
	}
	return providerOutput
}

func (v *CloudStack) WithProviderUpgrade(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = v.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (v *CloudStack) WithProviderUpgradeGit(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = v.customizeProviderConfig(e.clusterConfigGitPath(), fillers...)
	}
}

func (v *CloudStack) getControlPlaneIP() (string, error) {
	value, ok := os.LookupEnv(cloudStackClusterIPPoolEnvVar)
	if ok && value != "" {
		clusterIP, err := PopIPFromEnv(cloudStackClusterIPPoolEnvVar)
		if err != nil {
			v.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
		}
		return clusterIP, err
	}
	return "", fmt.Errorf("failed to generate ip for cloudstack from IP pool %s", value)
}

func (v *CloudStack) ClusterConfigFillers() []api.ClusterFiller {
	controlPlaneIP, err := v.getControlPlaneIP()
	if err != nil {
		v.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
	}
	return []api.ClusterFiller{
		api.WithPodCidr(os.Getenv(podCidrVar)),
		api.WithServiceCidr(os.Getenv(serviceCidrVar)),
		api.WithControlPlaneEndpointIP(controlPlaneIP),
	}
}

func RequiredCloudstackEnvVars() []string {
	return requiredCloudStackEnvVars
}
