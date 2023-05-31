package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	"github.com/aws/eks-anywhere/pkg/retrier"
	anywheretypes "github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

const (
	vsphereDatacenterVar        = "T_VSPHERE_DATACENTER"
	vsphereDatastoreVar         = "T_VSPHERE_DATASTORE"
	vsphereFolderVar            = "T_VSPHERE_FOLDER"
	vsphereNetworkVar           = "T_VSPHERE_NETWORK"
	vspherePrivateNetworkVar    = "T_VSPHERE_PRIVATE_NETWORK"
	vsphereResourcePoolVar      = "T_VSPHERE_RESOURCE_POOL"
	vsphereServerVar            = "T_VSPHERE_SERVER"
	vsphereSshAuthorizedKeyVar  = "T_VSPHERE_SSH_AUTHORIZED_KEY"
	vsphereStoragePolicyNameVar = "T_VSPHERE_STORAGE_POLICY_NAME"
	vsphereTlsInsecureVar       = "T_VSPHERE_TLS_INSECURE"
	vsphereTlsThumbprintVar     = "T_VSPHERE_TLS_THUMBPRINT"
	vsphereUsernameVar          = "EKSA_VSPHERE_USERNAME"
	vspherePasswordVar          = "EKSA_VSPHERE_PASSWORD"
	cidrVar                     = "T_VSPHERE_CIDR"
	privateNetworkCidrVar       = "T_VSPHERE_PRIVATE_NETWORK_CIDR"
	govcUrlVar                  = "VSPHERE_SERVER"
	govcInsecureVar             = "GOVC_INSECURE"
	govcDatacenterVar           = "GOVC_DATACENTER"
	vsphereTemplateEnvVarPrefix = "T_VSPHERE_TEMPLATE_"
	vsphereTemplatesFolder      = "T_VSPHERE_TEMPLATE_FOLDER"
	vsphereTestTagEnvVar        = "T_VSPHERE_TAG"
)

var requiredEnvVars = []string{
	vsphereDatacenterVar,
	vsphereDatastoreVar,
	vsphereFolderVar,
	vsphereNetworkVar,
	vspherePrivateNetworkVar,
	vsphereResourcePoolVar,
	vsphereServerVar,
	vsphereSshAuthorizedKeyVar,
	vsphereTlsInsecureVar,
	vsphereTlsThumbprintVar,
	vsphereUsernameVar,
	vspherePasswordVar,
	cidrVar,
	privateNetworkCidrVar,
	govcUrlVar,
	govcInsecureVar,
	govcDatacenterVar,
	vsphereTestTagEnvVar,
}

type VSphere struct {
	t                 *testing.T
	testsConfig       vsphereConfig
	fillers           []api.VSphereFiller
	clusterFillers    []api.ClusterFiller
	cidr              string
	GovcClient        *executables.Govc
	devRelease        *releasev1.EksARelease
	templatesRegistry *templateRegistry
}

type vsphereConfig struct {
	Datacenter        string
	Datastore         string
	Folder            string
	Network           string
	ResourcePool      string
	Server            string
	SSHAuthorizedKey  string
	StoragePolicyName string
	TLSInsecure       bool
	TLSThumbprint     string
	TemplatesFolder   string
}

// VSphereOpt is construction option for the E2E vSphere provider.
type VSphereOpt func(*VSphere)

func NewVSphere(t *testing.T, opts ...VSphereOpt) *VSphere {
	checkRequiredEnvVars(t, requiredEnvVars)
	c := buildGovc(t)
	config, err := readVSphereConfig()
	if err != nil {
		t.Fatalf("Failed reading vSphere tests config: %v", err)
	}
	v := &VSphere{
		t:           t,
		GovcClient:  c,
		testsConfig: config,
		fillers: []api.VSphereFiller{
			api.WithDatacenter(config.Datacenter),
			api.WithDatastoreForAllMachines(config.Datastore),
			api.WithFolderForAllMachines(config.Folder),
			api.WithNetwork(config.Network),
			api.WithResourcePoolForAllMachines(config.ResourcePool),
			api.WithServer(config.Server),
			api.WithSSHAuthorizedKeyForAllMachines(config.SSHAuthorizedKey),
			api.WithStoragePolicyNameForAllMachines(config.StoragePolicyName),
			api.WithTLSInsecure(config.TLSInsecure),
			api.WithTLSThumbprint(config.TLSThumbprint),
		},
	}

	v.cidr = os.Getenv(cidrVar)
	v.templatesRegistry = &templateRegistry{cache: map[string]string{}, generator: v}
	for _, opt := range opts {
		opt(v)
	}

	return v
}

// WithRedHat123VSphere vsphere test with redhat 1.23.
func WithRedHat123VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube123)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat124VSphere vsphere test with redhat 1.24.
func WithRedHat124VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube124)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat125VSphere vsphere test with redhat 1.25.
func WithRedHat125VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube125)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat126VSphere vsphere test with redhat 1.26.
func WithRedHat126VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube126)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat127VSphere vsphere test with redhat 1.27.
func WithRedHat127VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube127)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithUbuntu127 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.27
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu127() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube127)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu126 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.26
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu126() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube126)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu125 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.25
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu125() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube125)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu124 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.24
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu124() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube124)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu123 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.23
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu123() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube123)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

func WithBottleRocket123() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube123)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

// WithUbuntu123 returns a cluster config filler that sets the kubernetes version of the cluster to 1.23
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu123() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube123)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube123)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
}

// WithUbuntu124 returns a cluster config filler that sets the kubernetes version of the cluster to 1.24
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu124() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube124)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube124)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
}

// WithUbuntu125 returns a cluster config filler that sets the kubernetes version of the cluster to 1.25
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu125() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube125)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube125)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
}

// WithUbuntu126 returns a cluster config filler that sets the kubernetes version of the cluster to 1.26
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu126() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube126)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube126)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
}

// WithBottleRocket123 returns a cluster config filler that sets the kubernetes version of the cluster to 1.23
// as well as the right botllerocket template and osFamily for all VSphereMachaineConfigs.
func (v *VSphere) WithBottleRocket123() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube123)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube123)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		),
	)
}

// WithBottleRocket124 returns a cluster config filler that sets the kubernetes version of the cluster to 1.24
// as well as the right botllerocket template and osFamily for all VSphereMachaineConfigs.
func (v *VSphere) WithBottleRocket124() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube124)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube124)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		),
	)
}

// WithBottleRocket124 returns br 124 var.
func WithBottleRocket124() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube124)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

// WithBottleRocket125 returns br 1.25 var.
func WithBottleRocket125() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube125)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

// WithBottleRocket126 returns br 1.26 var.
func WithBottleRocket126() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube126)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

// WithBottleRocket127 returns br 1.27 var.
func WithBottleRocket127() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube127)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

func WithPrivateNetwork() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithVSphereStringFromEnvVar(vspherePrivateNetworkVar, api.WithNetwork),
		)
		v.cidr = os.Getenv(privateNetworkCidrVar)
	}
}

// WithLinkedCloneMode sets clone mode to LinkedClone for all the machine.
func WithLinkedCloneMode() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithCloneModeForAllMachines(anywherev1.LinkedClone),
		)
	}
}

// WithFullCloneMode sets clone mode to FullClone for all the machine.
func WithFullCloneMode() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithCloneModeForAllMachines(anywherev1.FullClone),
		)
	}
}

// WithDiskGiBForAllMachines sets diskGiB for all the machines.
func WithDiskGiBForAllMachines(value int) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithDiskGiBForAllMachines(value),
		)
	}
}

// WithNTPServersForAllMachines sets NTP servers for all the machines.
func WithNTPServersForAllMachines() VSphereOpt {
	return func(v *VSphere) {
		checkRequiredEnvVars(v.t, RequiredNTPServersEnvVars())
		v.fillers = append(v.fillers,
			api.WithNTPServersForAllMachines(GetNTPServersFromEnv()),
		)
	}
}

// WithBottlerocketKubernetesSettingsForAllMachines sets Bottlerocket Kubernetes settings for all the machines.
func WithBottlerocketKubernetesSettingsForAllMachines() VSphereOpt {
	return func(v *VSphere) {
		checkRequiredEnvVars(v.t, RequiredBottlerocketKubernetesSettingsEnvVars())
		unsafeSysctls, clusterDNSIPS, maxPods, err := GetBottlerocketKubernetesSettingsFromEnv()
		if err != nil {
			v.t.Fatalf("failed to get bottlerocket kubernetes settings from env: %v", err)
		}
		config := &anywherev1.BottlerocketConfiguration{
			Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
				AllowedUnsafeSysctls: unsafeSysctls,
				ClusterDNSIPs:        clusterDNSIPS,
				MaxPods:              maxPods,
			},
		}
		v.fillers = append(v.fillers,
			api.WithBottlerocketConfigurationForAllMachines(config),
		)
	}
}

// WithSSHAuthorizedKeyForAllMachines sets SSH authorized keys for all the machines.
func WithSSHAuthorizedKeyForAllMachines(sshKey string) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers, api.WithSSHAuthorizedKeyForAllMachines(sshKey))
	}
}

// WithVSphereTags with vsphere tags option.
func WithVSphereTags() VSphereOpt {
	return func(v *VSphere) {
		tags := []string{os.Getenv(vsphereTestTagEnvVar)}
		v.fillers = append(v.fillers,
			api.WithTagsForAllMachines(tags),
		)
	}
}

func WithVSphereWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.VSphereMachineConfigFiller) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers, vSphereMachineConfig(name, fillers...))

		v.clusterFillers = append(v.clusterFillers, buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup))
	}
}

// WithWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding VSphereMachineConfig to the cluster config.
func (v *VSphere) WithWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.VSphereMachineConfigFiller) api.ClusterConfigFiller {
	machineConfigFillers := append([]api.VSphereMachineConfigFiller{updateMachineSSHAuthorizedKey()}, fillers...)
	return api.JoinClusterConfigFillers(
		api.VSphereToConfigFiller(vSphereMachineConfig(name, machineConfigFillers...)),
		api.ClusterToConfigFiller(buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
	)
}

// WithWorkerNodeGroupConfiguration returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration item to the cluster config.
func (v *VSphere) WithWorkerNodeGroupConfiguration(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.ClusterToConfigFiller(buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup))
}

// updateMachineSSHAuthorizedKey updates a vsphere machine configs SSHAuthorizedKey.
func updateMachineSSHAuthorizedKey() api.VSphereMachineConfigFiller {
	return api.WithStringFromEnvVar(vsphereSshAuthorizedKeyVar, api.WithSSHKey)
}

// WithVSphereFillers adds VSphereFiller to the provider default fillers.
func WithVSphereFillers(fillers ...api.VSphereFiller) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers, fillers...)
	}
}

// Name returns the provider name. It satisfies the test framework Provider.
func (v *VSphere) Name() string {
	return "vsphere"
}

// Setup does nothing. It satisfies the test framework Provider.
func (v *VSphere) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (v *VSphere) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (v *VSphere) ClusterConfigUpdates() []api.ClusterConfigFiller {
	clusterIP, err := GetIP(v.cidr, ClusterIPPoolEnvVar)
	if err != nil {
		v.t.Fatalf("failed to get cluster ip for test environment: %v", err)
	}

	f := make([]api.ClusterFiller, 0, len(v.clusterFillers)+1)
	f = append(f, v.clusterFillers...)
	f = append(f, api.WithControlPlaneEndpointIP(clusterIP))

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.VSphereToConfigFiller(v.fillers...)}
}

// CleanupVMs deletes all the VMs owned by the test EKS-A cluster. It satisfies the test framework Provider.
func (v *VSphere) CleanupVMs(clusterName string) error {
	return cleanup.CleanUpVsphereTestResources(context.Background(), clusterName)
}

func (v *VSphere) WithProviderUpgrade(fillers ...api.VSphereFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.VSphereToConfigFiller(fillers...))
	}
}

func (v *VSphere) WithProviderUpgradeGit(fillers ...api.VSphereFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.VSphereToConfigFiller(fillers...))
	}
}

// WithNewVSphereWorkerNodeGroup adds a new worker node group to the cluster config.
func (v *VSphere) WithNewVSphereWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(
			api.ClusterToConfigFiller(buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
		)
	}
}

// Ubuntu123Template returns vsphere filler for 1.23 Ubuntu.
func (v *VSphere) Ubuntu123Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube123))
}

// Ubuntu124Template returns vsphere filler for 1.24 Ubuntu.
func (v *VSphere) Ubuntu124Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube124))
}

// Ubuntu125Template returns vsphere filler for 1.25 Ubuntu.
func (v *VSphere) Ubuntu125Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube125))
}

// Ubuntu126Template returns vsphere filler for 1.26 Ubuntu.
func (v *VSphere) Ubuntu126Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube126))
}

// Ubuntu127Template returns vsphere filler for 1.27 Ubuntu.
func (v *VSphere) Ubuntu127Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube127))
}

// Bottlerocket123Template returns vsphere filler for 1.23 BR.
func (v *VSphere) Bottlerocket123Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube123))
}

// Bottlerocket124Template returns vsphere filler for 1.24 BR.
func (v *VSphere) Bottlerocket124Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube124))
}

// Bottlerocket125Template returns vsphere filler for 1.25 BR.
func (v *VSphere) Bottlerocket125Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube125))
}

// Bottlerocket126Template returns vsphere filler for 1.26 BR.
func (v *VSphere) Bottlerocket126Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube126))
}

// Bottlerocket127Template returns vsphere filler for 1.27 BR.
func (v *VSphere) Bottlerocket127Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube127))
}

func (v *VSphere) getDevRelease() *releasev1.EksARelease {
	v.t.Helper()
	if v.devRelease == nil {
		latestRelease, err := getLatestDevRelease()
		if err != nil {
			v.t.Fatal(err)
		}
		v.devRelease = latestRelease
	}

	return v.devRelease
}

func (v *VSphere) templateForDevRelease(osFamily anywherev1.OSFamily, kubeVersion anywherev1.KubernetesVersion) string {
	v.t.Helper()
	return v.templatesRegistry.templateForRelease(v.t, osFamily, v.getDevRelease(), kubeVersion)
}

func RequiredVsphereEnvVars() []string {
	return requiredEnvVars
}

// VSphereExtraEnvVarPrefixes returns prefixes for env vars that although not always required,
// might be necessary for certain tests.
func VSphereExtraEnvVarPrefixes() []string {
	return []string{
		vsphereTemplateEnvVarPrefix,
	}
}

func vSphereMachineConfig(name string, fillers ...api.VSphereMachineConfigFiller) api.VSphereFiller {
	f := make([]api.VSphereMachineConfigFiller, 0, len(fillers)+6)
	// Need to add these because at this point the default fillers that assign these
	// values to all machines have already ran
	f = append(f,
		api.WithVSphereMachineDefaultValues(),
		api.WithDatastore(os.Getenv(vsphereDatastoreVar)),
		api.WithFolder(os.Getenv(vsphereFolderVar)),
		api.WithResourcePool(os.Getenv(vsphereResourcePoolVar)),
		api.WithStoragePolicyName(os.Getenv(vsphereStoragePolicyNameVar)),
		api.WithSSHKey(os.Getenv(vsphereSshAuthorizedKeyVar)),
	)
	f = append(f, fillers...)

	return api.WithVSphereMachineConfig(name, f...)
}

func buildVSphereWorkerNodeGroupClusterFiller(machineConfigName string, workerNodeGroup *WorkerNodeGroup) api.ClusterFiller {
	// Set worker node group ref to vsphere machine config
	workerNodeGroup.MachineConfigKind = anywherev1.VSphereMachineConfigKind
	workerNodeGroup.MachineConfigName = machineConfigName
	return workerNodeGroup.ClusterFiller()
}

func WithUbuntuForRelease(release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) VSphereOpt {
	return optionToSetTemplateForRelease(anywherev1.Ubuntu, release, kubeVersion)
}

func WithBottlerocketFromRelease(release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) VSphereOpt {
	return optionToSetTemplateForRelease(anywherev1.Bottlerocket, release, kubeVersion)
}

func (v *VSphere) WithBottleRocketForRelease(release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) api.ClusterConfigFiller {
	return api.VSphereToConfigFiller(
		api.WithTemplateForAllMachines(v.templatesRegistry.templateForRelease(v.t, anywherev1.Bottlerocket, v.getDevRelease(), kubeVersion)),
	)
}

func optionToSetTemplateForRelease(osFamily anywherev1.OSFamily, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templatesRegistry.templateForRelease(v.t, osFamily, v.getDevRelease(), kubeVersion)),
		)
	}
}

// envVarForTemplate looks for explicit configuration through an env var: "T_VSPHERE_TEMPLATE_{osFamily}_{eks-d version}"
// eg: T_VSPHERE_TEMPLATE_REDHAT_KUBERNETES_1_23_EKS_22.
func (v *VSphere) envVarForTemplate(osFamily, eksDName string) string {
	return fmt.Sprintf("T_VSPHERE_TEMPLATE_%s_%s", strings.ToUpper(osFamily), strings.ToUpper(strings.ReplaceAll(eksDName, "-", "_")))
}

// defaultNameForTemplate looks for a template with the name path: "{folder}/{eks-d version}-{osFamily}"
// eg: /SDDC-Datacenter/vm/Templates/kubernetes-1-23-eks-22-redhat.
func (v *VSphere) defaultNameForTemplate(osFamily, eksDName string) string {
	folder := v.testsConfig.TemplatesFolder
	if folder == "" {
		v.t.Log("vSphere templates folder is not configured.")
		return ""
	}
	return filepath.Join(folder, fmt.Sprintf("%s-%s", strings.ToLower(eksDName), strings.ToLower(osFamily)))
}

// defaultEnvVarForTemplate returns the value of the default template env vars: "T_VSPHERE_TEMPLATE_{osFamily}_{kubeVersion}"
// eg. T_VSPHERE_TEMPLATE_REDHAT_1_23.
func (v *VSphere) defaultEnvVarForTemplate(osFamily string, kubeVersion anywherev1.KubernetesVersion) string {
	if osFamily == "bottlerocket" {
		// This is only to maintain backwards compatibility with old env var naming
		osFamily = "br"
	}
	return fmt.Sprintf("T_VSPHERE_TEMPLATE_%s_%s", strings.ToUpper(osFamily), strings.ReplaceAll(string(kubeVersion), ".", "_"))
}

// searchTemplate returns template name if the given template exists in the datacenter.
func (v *VSphere) searchTemplate(ctx context.Context, template string) (string, error) {
	foundTemplate, err := v.GovcClient.SearchTemplate(context.Background(), v.testsConfig.Datacenter, template)
	if err != nil {
		return "", err
	}
	return foundTemplate, nil
}

func readVersionsBundles(t testing.TB, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) *releasev1.VersionsBundle {
	reader := newFileReader()
	b, err := releases.ReadBundlesForRelease(reader, release)
	if err != nil {
		t.Fatal(err)
	}

	return bundles.VersionsBundleForKubernetesVersion(b, string(kubeVersion))
}

func readVSphereConfig() (vsphereConfig, error) {
	return vsphereConfig{
		Datacenter:        os.Getenv(vsphereDatacenterVar),
		Datastore:         os.Getenv(vsphereDatastoreVar),
		Folder:            os.Getenv(vsphereFolderVar),
		Network:           os.Getenv(vsphereNetworkVar),
		ResourcePool:      os.Getenv(vsphereResourcePoolVar),
		Server:            os.Getenv(vsphereServerVar),
		SSHAuthorizedKey:  os.Getenv(vsphereSshAuthorizedKeyVar),
		StoragePolicyName: os.Getenv(vsphereStoragePolicyNameVar),
		TLSInsecure:       os.Getenv(vsphereTlsInsecureVar) == "true",
		TLSThumbprint:     os.Getenv(vsphereTlsThumbprintVar),
		TemplatesFolder:   os.Getenv(vsphereTemplatesFolder),
	}, nil
}

// ClusterStateValidations returns a list of provider specific validations.
func (v *VSphere) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{
		clusterf.RetriableStateValidation(
			retrier.NewWithMaxRetries(60, 5*time.Second),
			validations.ValidateCSI,
		),
	}
}

// ValidateNodesDiskGiB validates DiskGiB for all the machines.
func (v *VSphere) ValidateNodesDiskGiB(machines map[string]anywheretypes.Machine, expectedDiskSize int) error {
	v.t.Log("===================== Disk Size Validation Task =====================")
	for _, m := range machines {
		v.t.Log("Verifying disk size for VM", "Virtual Machine", m.Metadata.Name)
		diskSize, err := v.GovcClient.GetVMDiskSizeInGB(context.Background(), m.Metadata.Name, v.testsConfig.Datacenter)
		if err != nil {
			v.t.Fatalf("validating disk size: %v", err)
		}

		v.t.Log("Disk Size in GiB", "Expected", expectedDiskSize, "Actual", diskSize)
		if diskSize != expectedDiskSize {
			v.t.Fatalf("diskGib for node %s did not match the expected disk size. Expected=%dGiB, Actual=%dGiB", m.Metadata.Name, expectedDiskSize, diskSize)
		}
	}
	return nil
}
