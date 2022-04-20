package framework

import (
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
	cloudStackCidrVar                  = "T_CLOUDSTACK_CIDR"
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
	cloudStackCidrVar,
	podCidrVar,
	serviceCidrVar,
	cloudstackFeatureGateEnvVar,
	cloudstackB64EncodedSecretEnvVar,
}

type CloudStack struct {
	t           *testing.T
	fillers     []api.CloudStackFiller
	cidr        string
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
	c := &CloudStack{
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

	c.cidr = os.Getenv(cidrVar)
	c.podCidr = os.Getenv(podCidrVar)
	c.serviceCidr = os.Getenv(serviceCidrVar)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithRedhat121() CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplate),
		)
	}
}

func WithRedhat120() CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat120Var, api.WithCloudStackTemplate),
		)
	}
}

func WithCloudStackFillers(fillers ...api.CloudStackFiller) CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers, fillers...)
	}
}

func (c *CloudStack) Name() string {
	return "cloudstack"
}

func (c *CloudStack) Setup() {}

func (c *CloudStack) CustomizeProviderConfig(file string) []byte {
	return c.customizeProviderConfig(file, c.fillers...)
}

func (c *CloudStack) customizeProviderConfig(file string, fillers ...api.CloudStackFiller) []byte {
	providerOutput, err := api.AutoFillCloudStackProvider(file, fillers...)
	if err != nil {
		c.t.Fatalf("Error customizing provider config from file: %v", err)
	}
	return providerOutput
}

func (c *CloudStack) WithProviderUpgrade(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = c.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (c *CloudStack) WithProviderUpgradeGit(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = c.customizeProviderConfig(e.clusterConfigGitPath(), fillers...)
	}
}

func (c *CloudStack) getControlPlaneIP() (string, error) {
	value, ok := os.LookupEnv(cloudStackClusterIPPoolEnvVar)
	var clusterIP string
	var err error
	if ok && value != "" {
		clusterIP, err = PopIPFromEnv(cloudStackClusterIPPoolEnvVar)
		if err != nil {
			c.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
		}
	} else {
		clusterIP, err = GenerateUniqueIp(c.cidr)
		if err != nil {
			c.t.Fatalf("failed to generate ip for cloudstack %s: %v", c.cidr, err)
		}
	}
	return clusterIP, nil
}

func (c *CloudStack) ClusterConfigFillers() []api.ClusterFiller {
	controlPlaneIP, err := c.getControlPlaneIP()
	if err != nil {
		c.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
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
