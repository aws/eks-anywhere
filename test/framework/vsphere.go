package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	filereader "github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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
	t              *testing.T
	testsConfig    vsphereConfig
	fillers        []api.VSphereFiller
	clusterFillers []api.ClusterFiller
	cidr           string
	GovcClient     *executables.Govc
	devRelease     *releasev1.EksARelease
	templatesCache map[string]string
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
		templatesCache: make(map[string]string),
	}

	v.cidr = os.Getenv(cidrVar)

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// WithRedHat121VSphere vsphere test with redhat 1.21.
func WithRedHat121VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube121)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat122VSphere vsphere test with redhat 1.22.
func WithRedHat122VSphere() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.RedHat, anywherev1.Kube122)),
			api.WithOsFamilyForAllMachines(anywherev1.RedHat),
		)
	}
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

// WithUbuntu122 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.22
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu122() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube122)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu121 returns a VSphereOpt that adds API fillers to use a Ubuntu vSphere template for k8s 1.21
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu121() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube121)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		)
	}
}

func WithBottleRocket121() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube121)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		)
	}
}

func WithBottleRocket122() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube122)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
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

// WithUbuntu121 returns a cluster config filler that sets the kubernetes version of the cluster to 1.21
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu121() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube121)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
}

// WithUbuntu122 returns a cluster config filler that sets the kubernetes version of the cluster to 1.22
// as well as the right ubuntu template and osFamily for all VSphereMachineConfigs.
func (v *VSphere) WithUbuntu122() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube122)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube122)),
			api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
		),
	)
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

// WithBottleRocket121 returns a cluster config filler that sets the kubernetes version of the cluster to 1.21
// as well as the right botllerocket template and osFamily for all VSphereMachaineConfigs.
func (v *VSphere) WithBottleRocket121() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube121)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
		),
	)
}

// WithBottleRocket122 returns a cluster config filler that sets the kubernetes version of the cluster to 1.22
// as well as the right botllerocket template and osFamily for all VSphereMachaineConfigs.
func (v *VSphere) WithBottleRocket122() api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(anywherev1.Kube122)),
		api.VSphereToConfigFiller(
			api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube122)),
			api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
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

func WithPrivateNetwork() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithVSphereStringFromEnvVar(vspherePrivateNetworkVar, api.WithNetwork),
		)
		v.cidr = os.Getenv(privateNetworkCidrVar)
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
	return api.JoinClusterConfigFillers(
		api.VSphereToConfigFiller(vSphereMachineConfig(name, fillers...)),
		api.ClusterToConfigFiller(buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup)),
	)
}

// WithWorkerNodeGroupConfiguration returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration item to the cluster config.
func (v *VSphere) WithWorkerNodeGroupConfiguration(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	return api.ClusterToConfigFiller(buildVSphereWorkerNodeGroupClusterFiller(name, workerNodeGroup))
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

// Ubuntu121Template returns vsphere filler for 1.21 Ubuntu.
func (v *VSphere) Ubuntu121Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube121))
}

// Ubuntu122Template returns vsphere filler for 1.22 Ubuntu.
func (v *VSphere) Ubuntu122Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube122))
}

// Ubuntu123Template returns vsphere filler for 1.23 Ubuntu.
func (v *VSphere) Ubuntu123Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube123))
}

// Ubuntu124Template returns vsphere filler for 1.24 Ubuntu.
func (v *VSphere) Ubuntu124Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Ubuntu, anywherev1.Kube124))
}

// Bottlerocket121Template returns vsphere filler for 1.21 BR.
func (v *VSphere) Bottlerocket121Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube121))
}

// Bottlerocket122Template returns vsphere filler for 1.22 BR.
func (v *VSphere) Bottlerocket122Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube122))
}

// Bottlerocket123Template returns vsphere filler for 1.23 BR.
func (v *VSphere) Bottlerocket123Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube123))
}

// Bottlerocket124Template returns vsphere filler for 1.24 BR.
func (v *VSphere) Bottlerocket124Template() api.VSphereFiller {
	return api.WithTemplateForAllMachines(v.templateForDevRelease(anywherev1.Bottlerocket, anywherev1.Kube124))
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
	return v.templateForRelease(osFamily, v.getDevRelease(), kubeVersion)
}

// templateForRelease tries to find a suitable template for a particular eks-a release, k8s version and OS family.
// It follows these steps:
//
// 1. Look for explicit configuration through an env var: "T_VSPHERE_TEMPLATE_{osFamily}_{eks-d version}".
// This should be used for explicit configuration, mostly in local development for overrides.
//
// 2. If not present, look for a template if the default templates folder: "/SDDC-Datacenter/vm/Templates/{eks-d version}-{osFamily}"
// This is what should be used most of the time in CI, the explicit configuration is not present but the right template has already been
// imported to vSphere.
//
// 3. If the template doesn't exist, default to the value of the default template env vars: eg. "T_VSPHERE_TEMPLATE_UBUNTU_1_20".
// This is a catch all condition. Mostly for edge cases where the bundle has been updated with a new eks-d version, but the
// the new template hasn't been imported yet. It also preserves backwards compatibility.
func (v *VSphere) templateForRelease(osFamily anywherev1.OSFamily, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) string {
	v.t.Helper()
	osFamilyStr := string(osFamily)
	versionsBundle := readVersionsBundles(v.t, release, kubeVersion)
	eksDName := versionsBundle.EksD.Name

	templateEnvVarName := envVarForVSphereTemplate(osFamilyStr, eksDName)
	cacheKey := templateEnvVarName
	if template, ok := v.templatesCache[cacheKey]; ok {
		v.t.Logf("Template for release found in cache, using %s vSphere template.", template)
		return template
	}

	template, ok := os.LookupEnv(templateEnvVarName)
	if ok && template != "" {
		v.t.Logf("Env var %s is set, using %s vSphere template", templateEnvVarName, template)
		v.templatesCache[cacheKey] = template
		return template
	}
	v.t.Logf("Env var %s not is set, trying default generated template name", templateEnvVarName)

	// Env var is not set, try default template name
	folder := v.testsConfig.TemplatesFolder
	if folder == "" {
		v.t.Log("vSphere templates folder is not configured, can't continue template search.")
	} else {
		template = defaultNameForVSphereTemplate(folder, osFamilyStr, eksDName)
		foundTemplate, err := v.GovcClient.SearchTemplate(context.Background(), v.testsConfig.Datacenter, template)
		if err != nil {
			v.t.Fatalf("Failed checking if default template exists: %v", err)
		}

		if foundTemplate != "" {
			v.t.Logf("Default template for release exists, using %s vSphere template.", template)
			v.templatesCache[cacheKey] = template
			return template
		}
		v.t.Logf("Default template %s for release doesn't exit.", template)
	}

	// Default template doesn't exist, try legacy generic env var
	// It is not guaranteed that this template will work for the given release, if they don't match the
	// same ekd-d release, the test will fail. This is just a catch all last try for cases where the new template
	// hasn't been imported with its own name but the default one matches the same eks-d release.
	templateEnvVarName = defaultEnvVarForTemplate(osFamilyStr, kubeVersion)
	template, ok = os.LookupEnv(templateEnvVarName)
	if !ok || template == "" {
		v.t.Fatalf("Env var %s for default template is not set, can't determine which template to use", templateEnvVarName)
	}

	v.t.Logf("Env var %s is set, using %s vSphere template. There are no guarantees this template will be valid. Cluster validation might fail.", templateEnvVarName, template)

	v.templatesCache[cacheKey] = template
	return template
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

func optionToSetTemplateForRelease(osFamily anywherev1.OSFamily, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithTemplateForAllMachines(v.templateForRelease(osFamily, release, kubeVersion)),
		)
	}
}

func envVarForVSphereTemplate(osFamily, eksDName string) string {
	return fmt.Sprintf("T_VSPHERE_TEMPLATE_%s_%s", strings.ToUpper(osFamily), strings.ToUpper(strings.ReplaceAll(eksDName, "-", "_")))
}

func defaultNameForVSphereTemplate(templatesFolder, osFamily, eksDName string) string {
	return filepath.Join(templatesFolder, fmt.Sprintf("%s-%s", strings.ToLower(eksDName), strings.ToLower(osFamily)))
}

func defaultEnvVarForTemplate(osFamily string, kubeVersion anywherev1.KubernetesVersion) string {
	if osFamily == "bottlerocket" {
		// This is only to maintain backwards compatibility with old env var naming
		osFamily = "br"
	}
	return fmt.Sprintf("T_VSPHERE_TEMPLATE_%s_%s", strings.ToUpper(osFamily), strings.ReplaceAll(string(kubeVersion), ".", "_"))
}

func readVersionsBundles(t testing.TB, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion) *releasev1.VersionsBundle {
	reader := filereader.NewReader(filereader.WithUserAgent("eks-a-e2e-tests"))
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

func (v *VSphere) ClusterValidations() []ClusterValidation {
	return []ClusterValidation{
		validateCSI,
	}
}

func validateCSI(ctx context.Context, vc ClusterValidatorConfig) error {
	clusterClient := vc.ClusterClient
	if vc.ClusterSpec.Cluster.IsManaged() {
		clusterClient = vc.ManagementClusterClient
	}

	yaml := vc.ClusterSpec.Config.VSphereDatacenter
	datacenter := &anywherev1.VSphereDatacenterConfig{}
	key := types.NamespacedName{Namespace: vc.ClusterSpec.Cluster.Namespace, Name: vc.ClusterSpec.Cluster.Name}
	err := clusterClient.Get(ctx, key, datacenter)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}

	disableCSI := datacenter.Spec.DisableCSI
	yamlCSI := yaml.Spec.DisableCSI
	if disableCSI != yamlCSI {
		return fmt.Errorf("cilium config does not match. %t and %t", disableCSI, yamlCSI)
	}

	return nil
}
