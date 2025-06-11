package framework

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/nutanix"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	nutanixEndpoint                  = "T_NUTANIX_ENDPOINT"
	nutanixPort                      = "T_NUTANIX_PORT"
	nutanixAdditionalTrustBundle     = "T_NUTANIX_ADDITIONAL_TRUST_BUNDLE"
	nutanixInsecure                  = "T_NUTANIX_INSECURE"
	nutanixMachineBootType           = "T_NUTANIX_MACHINE_BOOT_TYPE"
	nutanixMachineMemorySize         = "T_NUTANIX_MACHINE_MEMORY_SIZE"
	nutanixSystemDiskSize            = "T_NUTANIX_SYSTEMDISK_SIZE"
	nutanixMachineVCPUsPerSocket     = "T_NUTANIX_MACHINE_VCPU_PER_SOCKET"
	nutanixMachineVCPUSocket         = "T_NUTANIX_MACHINE_VCPU_SOCKET"
	nutanixPrismElementClusterName   = "T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME"
	nutanixSSHAuthorizedKey          = "T_NUTANIX_SSH_AUTHORIZED_KEY"
	nutanixSubnetName                = "T_NUTANIX_SUBNET_NAME"
	nutanixControlPlaneEndpointIP    = "T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP"
	nutanixControlPlaneCidrVar       = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixPodCidrVar                = "T_NUTANIX_POD_CIDR"
	nutanixServiceCidrVar            = "T_NUTANIX_SERVICE_CIDR"
	nutanixTemplateNameUbuntu129Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_29"
	nutanixTemplateNameUbuntu130Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_30"
	nutanixTemplateNameUbuntu131Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_31"
	nutanixTemplateNameUbuntu132Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_32"
	nutanixTemplateNameUbuntu133Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_33"
	nutanixTemplateNameRedHat129Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_29"
	nutanixTemplateNameRedHat130Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_30"
	nutanixTemplateNameRedHat131Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_31"
	nutanixTemplateNameRedHat132Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_32"
	nutanixTemplateNameRedHat133Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_33"
	nutanixTemplateNameRedHat9129Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_29"
	nutanixTemplateNameRedHat9130Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_30"
	nutanixTemplateNameRedHat9131Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_31"
	nutanixTemplateNameRedHat9132Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_32"
	nutanixTemplateNameRedHat9133Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_33"
)

var requiredNutanixEnvVars = []string{
	constants.EksaNutanixUsernameKey,
	constants.EksaNutanixPasswordKey,
	nutanixEndpoint,
	nutanixPort,
	nutanixAdditionalTrustBundle,
	nutanixMachineBootType,
	nutanixMachineMemorySize,
	nutanixSystemDiskSize,
	nutanixMachineVCPUsPerSocket,
	nutanixMachineVCPUSocket,
	nutanixPrismElementClusterName,
	nutanixSSHAuthorizedKey,
	nutanixSubnetName,
	nutanixPodCidrVar,
	nutanixServiceCidrVar,
	nutanixTemplateNameUbuntu129Var,
	nutanixTemplateNameUbuntu130Var,
	nutanixTemplateNameUbuntu131Var,
	nutanixTemplateNameUbuntu132Var,
	nutanixTemplateNameUbuntu133Var,
	nutanixTemplateNameRedHat129Var,
	nutanixTemplateNameRedHat130Var,
	nutanixTemplateNameRedHat131Var,
	nutanixTemplateNameRedHat132Var,
	nutanixTemplateNameRedHat133Var,
	nutanixTemplateNameRedHat9129Var,
	nutanixTemplateNameRedHat9130Var,
	nutanixTemplateNameRedHat9131Var,
	nutanixTemplateNameRedHat9132Var,
	nutanixTemplateNameRedHat9133Var,
	nutanixInsecure,
}

type Nutanix struct {
	t                      *testing.T
	fillers                []api.NutanixFiller
	clusterFillers         []api.ClusterFiller
	client                 nutanix.PrismClient
	controlPlaneEndpointIP string
	cpCidr                 string
	podCidr                string
	serviceCidr            string
	devRelease             *releasev1.EksARelease
	templatesRegistry      *templateRegistry
}

type NutanixOpt func(*Nutanix)

func NewNutanix(t *testing.T, opts ...NutanixOpt) *Nutanix {
	checkRequiredEnvVars(t, requiredNutanixEnvVars)

	nutanixProvider := &Nutanix{
		t: t,
		fillers: []api.NutanixFiller{
			api.WithNutanixStringFromEnvVar(nutanixEndpoint, api.WithNutanixEndpoint),
			api.WithNutanixIntFromEnvVar(nutanixPort, api.WithNutanixPort),
			api.WithNutanixStringFromEnvVar(nutanixAdditionalTrustBundle, api.WithNutanixAdditionalTrustBundle),
			api.WithNutanixStringFromEnvVar(nutanixMachineMemorySize, api.WithNutanixMachineMemorySize),
			api.WithNutanixStringFromEnvVar(nutanixSystemDiskSize, api.WithNutanixMachineSystemDiskSize),
			api.WithNutanixInt32FromEnvVar(nutanixMachineVCPUsPerSocket, api.WithNutanixMachineVCPUsPerSocket),
			api.WithNutanixInt32FromEnvVar(nutanixMachineVCPUSocket, api.WithNutanixMachineVCPUSocket),
			api.WithNutanixStringFromEnvVar(nutanixSSHAuthorizedKey, api.WithNutanixSSHAuthorizedKey),
			api.WithNutanixBoolFromEnvVar(nutanixInsecure, api.WithNutanixInsecure),
			// Assumption: generated clusterconfig by nutanix provider sets name as id type by default.
			// for uuid specific id type, we will set it thru each specific test so that current CI
			// works as is with name id type for following resources
			api.WithNutanixStringFromEnvVar(nutanixPrismElementClusterName, api.WithNutanixPrismElementClusterName),
			api.WithNutanixStringFromEnvVar(nutanixSubnetName, api.WithNutanixSubnetName),
		},
	}

	nutanixProvider.controlPlaneEndpointIP = os.Getenv(nutanixControlPlaneEndpointIP)
	nutanixProvider.cpCidr = os.Getenv(nutanixControlPlaneCidrVar)
	nutanixProvider.podCidr = os.Getenv(nutanixPodCidrVar)
	nutanixProvider.serviceCidr = os.Getenv(nutanixServiceCidrVar)
	client, err := nutanix.NewPrismClient(os.Getenv(nutanixEndpoint), os.Getenv(nutanixPort), true)
	if err != nil {
		t.Fatalf("Failed to initialize Nutanix Prism Client: %v", err)
	}
	nutanixProvider.client = client
	nutanixProvider.templatesRegistry = &templateRegistry{cache: map[string]string{}, generator: nutanixProvider}

	for _, opt := range opts {
		opt(nutanixProvider)
	}

	return nutanixProvider
}

// RequiredNutanixEnvVars returns a list of environment variables needed for Nutanix tests.
func RequiredNutanixEnvVars() []string {
	return requiredNutanixEnvVars
}

// Name returns the provider name. It satisfies the test framework Provider.
func (n *Nutanix) Name() string {
	return "nutanix"
}

// Setup does nothing. It satisfies the test framework Provider.
func (n *Nutanix) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (n *Nutanix) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// CleanupResources satisfies the test framework Provider.
func (n *Nutanix) CleanupResources(clustername string) error {
	return cleanup.NutanixTestResources(clustername, os.Getenv(nutanixEndpoint), os.Getenv(nutanixPort), true, true)
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (n *Nutanix) ClusterConfigUpdates() []api.ClusterConfigFiller {
	f := make([]api.ClusterFiller, 0, len(n.clusterFillers)+3)
	f = append(f, n.clusterFillers...)
	if n.controlPlaneEndpointIP != "" {
		f = append(f, api.WithControlPlaneEndpointIP(n.controlPlaneEndpointIP))
	} else {
		clusterIP, err := GetIP(n.cpCidr, ClusterIPPoolEnvVar)
		if err != nil {
			n.t.Fatalf("failed to get cluster ip for test environment: %v", err)
		}
		f = append(f, api.WithControlPlaneEndpointIP(clusterIP))
	}

	if n.podCidr != "" {
		f = append(f, api.WithPodCidr(n.podCidr))
	}

	if n.serviceCidr != "" {
		f = append(f, api.WithServiceCidr(n.serviceCidr))
	}

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.NutanixToConfigFiller(n.fillers...)}
}

// WithProviderUpgrade returns a ClusterE2EOpt that updates the cluster config for provider-specific upgrade.
func (n *Nutanix) WithProviderUpgrade(fillers ...api.NutanixFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.NutanixToConfigFiller(fillers...))
	}
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right template for all
// nutanix machine configs.
func (n *Nutanix) WithKubeVersionAndOS(_ anywherev1.KubernetesVersion, _ OS, _ *releasev1.EksARelease, _ ...string) api.ClusterConfigFiller {
	// TODO: Update tests to use this
	panic("Not implemented for Nutanix yet")
}

// WithNewWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding NutanixMachineConfig to the cluster config.
func (n *Nutanix) WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	// TODO: Implement for Nutanix provider
	panic("Not implemented for Nutanix yet")
}

// withNutanixKubeVersionAndOS returns a NutanixOpt that adds API fillers to use a Nutanix template for
// the specified OS family and version (default if not provided), corresponding to a particular
// Kubernetes version, in addition to configuring all machine configs to use this OS family.
func withNutanixKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) NutanixOpt {
	return func(n *Nutanix) {
		n.fillers = append(n.fillers,
			n.templateForKubeVersionAndOS(kubeVersion, os, release),
			api.WithOsFamilyForAllNutanixMachines(osFamiliesForOS[os]),
		)
	}
}

// WithUbuntu129Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.29
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu129Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, nil)
}

// WithUbuntu130Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.30
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu130Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, nil)
}

// WithUbuntu131Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.31
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu131Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, nil)
}

// WithUbuntu132Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.32
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu132Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube132, Ubuntu2004, nil)
}

// WithUbuntu133Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.33
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu133Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube133, Ubuntu2004, nil)
}

// WithRedHat129Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template for k8s 1.29
// and the "redhat" osFamily in all machine configs.
func WithRedHat129Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube129, RedHat8, nil)
}

// WithRedHat130Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template for k8s 1.30
// and the "redhat" osFamily in all machine configs.
func WithRedHat130Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube130, RedHat8, nil)
}

// WithRedHat131Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template for k8s 1.31
// and the "redhat" osFamily in all machine configs.
func WithRedHat131Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube131, RedHat8, nil)
}

// WithRedHat132Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template for k8s 1.32
// and the "redhat" osFamily in all machine configs.
func WithRedHat132Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube132, RedHat8, nil)
}

// WithRedHat133Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template for k8s 1.33
// and the "redhat" osFamily in all machine configs.
func WithRedHat133Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube133, RedHat8, nil)
}

// WithRedHat9Kubernetes128Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.28
// and the "redhat" osFamily in all machine configs.

// WithRedHat9Kubernetes129Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.29
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes129Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube129, RedHat9, nil)
}

// WithRedHat9Kubernetes130Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.30
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes130Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube130, RedHat9, nil)
}

// WithRedHat9Kubernetes131Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.31
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes131Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube131, RedHat9, nil)
}

// WithRedHat9Kubernetes132Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.32
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes132Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube132, RedHat9, nil)
}

// WithRedHat9Kubernetes133Nutanix returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template for k8s 1.33
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes133Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube133, RedHat9, nil)
}

// withNutanixKubeVersionAndOSForUUID returns a NutanixOpt that adds API fillers to use a Nutanix template UUID
// corresponding to the provided OS family and Kubernetes version, in addition to configuring all machine configs
// to use this OS family.
func withNutanixKubeVersionAndOSForUUID(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) NutanixOpt {
	return func(n *Nutanix) {
		useBundlesOverride := getBundlesOverride() == "true"
		name := n.templateForDevRelease(kubeVersion, os, useBundlesOverride)
		n.fillers = append(n.fillers, n.withNutanixUUID(name, osFamiliesForOS[os])...)
	}
}

// WithUbuntu128NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.28
// and the "ubuntu" osFamily in all machine configs.

// WithUbuntu129NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.29
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu129NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube129, Ubuntu2004, nil)
}

// WithUbuntu130NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.30
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu130NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube130, Ubuntu2004, nil)
}

// WithUbuntu131NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.31
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu131NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube131, Ubuntu2004, nil)
}

// WithUbuntu132NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.32
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu132NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube132, Ubuntu2004, nil)
}

// WithUbuntu133NutanixUUID returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template UUID for k8s 1.33
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu133NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube133, Ubuntu2004, nil)
}

// WithRedHat128NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat Nutanix template UUID for k8s 1.28
// and the "redhat" osFamily in all machine configs.

// WithRedHat129NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template UUID for k8s 1.29
// and the "redhat" osFamily in all machine configs.
func WithRedHat129NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube129, RedHat8, nil)
}

// WithRedHat130NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat Nutanix template UUID for k8s 1.30
// and the "redhat" osFamily in all machine configs.
func WithRedHat130NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube130, RedHat8, nil)
}

// WithRedHat131NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat Nutanix template UUID for k8s 1.31
// and the "redhat" osFamily in all machine configs.
func WithRedHat131NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube131, RedHat8, nil)
}

// WithRedHat133NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 8 Nutanix template UUID for k8s 1.33
// and the "redhat" osFamily in all machine configs.
func WithRedHat133NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube133, RedHat8, nil)
}

// WithRedHat9Kubernetes128NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.28
// and the "redhat" osFamily in all machine configs.

// WithRedHat9Kubernetes129NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.28
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes129NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube129, RedHat9, nil)
}

// WithRedHat9Kubernetes130NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.30
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes130NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube130, RedHat9, nil)
}

// WithRedHat9Kubernetes131NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.31
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes131NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube131, RedHat9, nil)
}

// WithRedHat9Kubernetes132NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.32
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes132NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube132, RedHat9, nil)
}

// WithRedHat9Kubernetes133NutanixUUID returns a NutanixOpt that adds API fillers to use a RedHat 9 Nutanix template UUID for k8s 1.33
// and the "redhat" osFamily in all machine configs.
func WithRedHat9Kubernetes133NutanixUUID() NutanixOpt {
	return withNutanixKubeVersionAndOSForUUID(anywherev1.Kube133, RedHat9, nil)
}

func (n *Nutanix) withNutanixUUID(name string, osFamily anywherev1.OSFamily) []api.NutanixFiller {
	uuid, err := n.client.GetImageUUIDFromName(context.Background(), name)
	if err != nil {
		n.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
	}
	return append([]api.NutanixFiller{},
		api.WithNutanixMachineTemplateImageUUID(*uuid),
		api.WithOsFamilyForAllNutanixMachines(osFamily),
	)
}

// WithPrismElementClusterUUID returns a NutanixOpt that adds API fillers to use a PE Cluster UUID.
func WithPrismElementClusterUUID() NutanixOpt {
	return func(n *Nutanix) {
		name := os.Getenv(nutanixPrismElementClusterName)
		uuid, err := n.client.GetClusterUUIDFromName(context.Background(), name)
		if err != nil {
			n.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
		}
		n.fillers = append(n.fillers, api.WithNutanixPrismElementClusterUUID(*uuid))
	}
}

// WithNutanixSubnetUUID returns a NutanixOpt that adds API fillers to use a Subnet UUID.
func WithNutanixSubnetUUID() NutanixOpt {
	return func(n *Nutanix) {
		name := os.Getenv(nutanixSubnetName)
		uuid, err := n.client.GetSubnetUUIDFromName(context.Background(), name)
		if err != nil {
			n.t.Fatalf("Failed to get UUID for image %s: %v", name, err)
		}
		n.fillers = append(n.fillers, api.WithNutanixSubnetUUID(*uuid))
	}
}

// templateForKubeVersionAndOS returns a Nutanix filler for the given OS and Kubernetes version.
func (n *Nutanix) templateForKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.NutanixFiller {
	var template string
	useBundlesOverride := getBundlesOverride() == "true"
	if release == nil {
		template = n.templateForDevRelease(kubeVersion, os, useBundlesOverride)
	} else {
		template = n.templatesRegistry.templateForRelease(n.t, release, kubeVersion, os, useBundlesOverride)
	}
	return api.WithNutanixMachineTemplateImageName(template)
}

// Ubuntu128Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.

// Ubuntu129Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu129Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, nil)
}

// Ubuntu130Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu130Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, nil)
}

// Ubuntu131Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu131Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, nil)
}

// Ubuntu132Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu132Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube132, Ubuntu2004, nil)
}

// Ubuntu133Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu133Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube133, Ubuntu2004, nil)
}

// RedHat128Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.

// RedHat129Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat129Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube129, RedHat8, nil)
}

// RedHat130Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat130Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube130, RedHat8, nil)
}

// RedHat131Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat131Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube131, RedHat8, nil)
}

// RedHat132Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat132Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube132, RedHat8, nil)
}

// RedHat133Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat133Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube133, RedHat8, nil)
}

// RedHat9Kubernetes128Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.

// RedHat9Kubernetes129Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat9Kubernetes129Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube129, RedHat9, nil)
}

// RedHat9Kubernetes130Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat9Kubernetes130Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube130, RedHat9, nil)
}

// RedHat9Kubernetes131Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat9Kubernetes131Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube131, RedHat9, nil)
}

// RedHat9Kubernetes132Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat9Kubernetes132Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube132, RedHat9, nil)
}

// RedHat9Kubernetes133Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) RedHat9Kubernetes133Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube133, RedHat9, nil)
}

// ClusterStateValidations returns a list of provider specific ClusterStateValidations.
func (n *Nutanix) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}

func (n *Nutanix) getDevRelease() *releasev1.EksARelease {
	n.t.Helper()
	if n.devRelease == nil {
		localDevRelease, err := localEksaCLIDevVersionRelease()
		if err != nil {
			n.t.Fatal(err)
		}
		n.t.Log("Using local eksa dev release:", localDevRelease.Version)
		n.devRelease = localDevRelease
	}

	return n.devRelease
}

func (n *Nutanix) templateForDevRelease(kubeVersion anywherev1.KubernetesVersion, os OS, useBundlesOverride bool) string {
	n.t.Helper()
	return n.templatesRegistry.templateForRelease(n.t, n.getDevRelease(), kubeVersion, os, useBundlesOverride)
}

// envVarForTemplate looks for explicit configuration through an env var: "T_NUTANIX_TEMPLATE_{osFamily}_{eks-d version}"
// eg: T_NUTANIX_TEMPLATE_UBUNTU_KUBERNETES_1_27_EKS_22.
func (n *Nutanix) envVarForTemplate(os OS, eksDName string) string {
	return fmt.Sprintf("T_NUTANIX_TEMPLATE_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ToUpper(strings.ReplaceAll(eksDName, "-", "_")))
}

// defaultNameForTemplate looks for a template: "{eks-d version}-{osFamily}"
// eg: kubernetes-1-27-eks-22-ubuntu.
func (n *Nutanix) defaultNameForTemplate(os OS, eksDName string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(eksDName), strings.ToLower(string(os)))
}

// defaultEnvVarForTemplate returns the value of the default template env vars: "T_NUTANIX_TEMPLATE_{osFamily}_{kubeVersion}"
// eg. T_NUTANIX_TEMPLATE_UBUNTU_1_27.
func (n *Nutanix) defaultEnvVarForTemplate(os OS, kubeVersion anywherev1.KubernetesVersion) string {
	return fmt.Sprintf("T_NUTANIX_TEMPLATE_NAME_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ReplaceAll(string(kubeVersion), ".", "_"))
}

// searchTemplate returns template name if the given template exists in Prism Central.
func (n *Nutanix) searchTemplate(ctx context.Context, template string) (string, error) {
	// TODO: implement search functionality for Nutanix templates
	return "", nil
}
