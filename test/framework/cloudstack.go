package framework

import (
	"context"
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	cloudstackDomainVar                = "T_CLOUDSTACK_DOMAIN"
	cloudstackZoneVar                  = "T_CLOUDSTACK_ZONE"
	cloudstackZone2Var                 = "T_CLOUDSTACK_ZONE_2"
	cloudstackZone3Var                 = "T_CLOUDSTACK_ZONE_3"
	cloudstackAccountVar               = "T_CLOUDSTACK_ACCOUNT"
	cloudstackNetworkVar               = "T_CLOUDSTACK_NETWORK"
	cloudstackNetwork2Var              = "T_CLOUDSTACK_NETWORK_2"
	cloudstackNetwork3Var              = "T_CLOUDSTACK_NETWORK_3"
	cloudstackCredentialsVar           = "T_CLOUDSTACK_CREDENTIALS"
	cloudstackCredentials2Var          = "T_CLOUDSTACK_CREDENTIALS_2"
	cloudstackCredentials3Var          = "T_CLOUDSTACK_CREDENTIALS_3"
	cloudstackManagementServerVar      = "T_CLOUDSTACK_MANAGEMENT_SERVER"
	cloudstackManagementServer2Var     = "T_CLOUDSTACK_MANAGEMENT_SERVER_2"
	cloudstackManagementServer3Var     = "T_CLOUDSTACK_MANAGEMENT_SERVER_3"
	cloudstackSshAuthorizedKeyVar      = "T_CLOUDSTACK_SSH_AUTHORIZED_KEY"
	cloudstackTemplateRedhat121Var     = "T_CLOUDSTACK_TEMPLATE_REDHAT_1_21"
	cloudstackTemplateRedhat122Var     = "T_CLOUDSTACK_TEMPLATE_REDHAT_1_22"
	cloudstackTemplateRedhat123Var     = "T_CLOUDSTACK_TEMPLATE_REDHAT_1_23"
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
	cloudstackAccountVar,
	cloudstackDomainVar,
	cloudstackZoneVar,
	cloudstackZone2Var,
	cloudstackZone3Var,
	cloudstackCredentialsVar,
	cloudstackCredentials2Var,
	cloudstackCredentials3Var,
	cloudstackAccountVar,
	cloudstackNetworkVar,
	cloudstackNetwork2Var,
	cloudstackNetwork3Var,
	cloudstackManagementServerVar,
	cloudstackManagementServer2Var,
	cloudstackManagementServer3Var,
	cloudstackSshAuthorizedKeyVar,
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
	t              *testing.T
	fillers        []api.CloudStackFiller
	clusterFillers []api.ClusterFiller
	cidr           string
	podCidr        string
	serviceCidr    string
}

type CloudStackOpt func(*CloudStack)

func UpdateRedhatTemplate121Var() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplateForAllMachines)
}

// UpdateRedhatTemplate122Var updates the CloudStackTemplate for all machines to the one corresponding to K8s 1.22.
func UpdateRedhatTemplate122Var() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat122Var, api.WithCloudStackTemplateForAllMachines)
}

// UpdateRedhatTemplate123Var updates the CloudStackTemplate for all machines to the one corresponding to K8s 1.23.
func UpdateRedhatTemplate123Var() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat123Var, api.WithCloudStackTemplateForAllMachines)
}

func UpdateLargerCloudStackComputeOffering() api.CloudStackFiller {
	return api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargerVar, api.WithCloudStackComputeOfferingForAllMachines)
}

// UpdateAddCloudStackAz3 add availiability zone 3 to the cluster spec.
func UpdateAddCloudStackAz3() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZone3Var, cloudstackCredentials3Var, cloudstackNetwork3Var,
		cloudstackManagementServer3Var, api.WithCloudStackAz)
}

func UpdateAddCloudStackAz2() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZone2Var, cloudstackCredentials2Var, cloudstackNetwork2Var,
		cloudstackManagementServer2Var, api.WithCloudStackAz)
}

func UpdateAddCloudStackAz1() api.CloudStackFiller {
	return api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZoneVar, cloudstackCredentialsVar, cloudstackNetworkVar,
		cloudstackManagementServerVar, api.WithCloudStackAz)
}

func RemoveAllCloudStackAzs() api.CloudStackFiller {
	return api.RemoveCloudStackAzs()
}

func NewCloudStack(t *testing.T, opts ...CloudStackOpt) *CloudStack {
	checkRequiredEnvVars(t, requiredCloudStackEnvVars)
	c := &CloudStack{
		t: t,
		fillers: []api.CloudStackFiller{
			api.RemoveCloudStackAzs(),
			api.WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZoneVar, cloudstackCredentialsVar, cloudstackNetworkVar,
				cloudstackManagementServerVar, api.WithCloudStackAz),
			api.WithCloudStackStringFromEnvVar(cloudstackSshAuthorizedKeyVar, api.WithCloudStackSSHAuthorizedKey),
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplateForAllMachines),
			api.WithCloudStackStringFromEnvVar(cloudstackComputeOfferingLargeVar, api.WithCloudStackComputeOfferingForAllMachines),
		},
	}

	c.cidr = os.Getenv(cloudStackCidrVar)
	c.podCidr = os.Getenv(podCidrVar)
	c.serviceCidr = os.Getenv(serviceCidrVar)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithCloudStackWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.CloudStackMachineConfigFiller) CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers, cloudStackMachineConfig(name, fillers...))

		c.clusterFillers = append(c.clusterFillers, buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup))
	}
}

func WithCloudStackRedhat121() CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat121Var, api.WithCloudStackTemplateForAllMachines),
		)
	}
}

// WithCloudStackRedhat122 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.22.
func WithCloudStackRedhat122() CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat122Var, api.WithCloudStackTemplateForAllMachines),
		)
	}
}

// WithCloudStackRedhat123 returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.23.
func WithCloudStackRedhat123() CloudStackOpt {
	return func(c *CloudStack) {
		c.fillers = append(c.fillers,
			api.WithCloudStackStringFromEnvVar(cloudstackTemplateRedhat123Var, api.WithCloudStackTemplateForAllMachines),
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

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (c *CloudStack) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (c *CloudStack) ClusterConfigUpdates() []api.ClusterConfigFiller {
	controlPlaneIP, err := c.getControlPlaneIP()
	if err != nil {
		c.t.Fatalf("failed to pop cluster ip from test environment: %v", err)
	}

	f := make([]api.ClusterFiller, 0, len(c.clusterFillers)+3)
	f = append(f, c.clusterFillers...)
	f = append(f,
		api.WithPodCidr(os.Getenv(podCidrVar)),
		api.WithServiceCidr(os.Getenv(serviceCidrVar)),
		api.WithControlPlaneEndpointIP(controlPlaneIP))

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.CloudStackToConfigFiller(c.fillers...)}
}

func (c *CloudStack) CleanupVMs(clusterName string) error {
	return cleanup.CleanUpCloudstackTestResources(context.Background(), clusterName, false)
}

func (c *CloudStack) WithProviderUpgrade(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.CloudStackToConfigFiller(fillers...))
	}
}

func (c *CloudStack) WithProviderUpgradeGit(fillers ...api.CloudStackFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.CloudStackToConfigFiller(fillers...))
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

func RequiredCloudstackEnvVars() []string {
	return requiredCloudStackEnvVars
}

func (c *CloudStack) WithNewCloudStackWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.CloudStackMachineConfigFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(
			api.CloudStackToConfigFiller(cloudStackMachineConfig(name, fillers...)),
			api.ClusterToConfigFiller(buildCloudStackWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		)
	}
}

func cloudStackMachineConfig(name string, fillers ...api.CloudStackMachineConfigFiller) api.CloudStackFiller {
	f := make([]api.CloudStackMachineConfigFiller, 0, len(fillers)+2)
	// Need to add these because at this point the default fillers that assign these
	// values to all machines have already ran
	f = append(f,
		api.WithCloudStackComputeOffering(os.Getenv(cloudstackComputeOfferingLargeVar)),
		api.WithCloudStackSSHKey(os.Getenv(cloudstackSshAuthorizedKeyVar)),
	)
	f = append(f, fillers...)

	return api.WithCloudStackMachineConfig(name, f...)
}

func buildCloudStackWorkerNodeGroupClusterFiller(machineConfigName string, workerNodeGroup *WorkerNodeGroup) api.ClusterFiller {
	// Set worker node group ref to cloudstack machine config
	workerNodeGroup.MachineConfigKind = anywherev1.CloudStackMachineConfigKind
	workerNodeGroup.MachineConfigName = machineConfigName
	return workerNodeGroup.ClusterFiller()
}

// ClusterStateValidations returns a list of provider specific validations.
func (c *CloudStack) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}
