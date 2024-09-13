package framework

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	tinkerbellProviderName                              = "tinkerbell"
	tinkerbellBootstrapIPEnvVar                         = "T_TINKERBELL_BOOTSTRAP_IP"
	tinkerbellControlPlaneNetworkCidrEnvVar             = "T_TINKERBELL_CP_NETWORK_CIDR"
	tinkerbellImageUbuntu127EnvVar                      = "T_TINKERBELL_IMAGE_UBUNTU_1_27"
	tinkerbellImageUbuntu128EnvVar                      = "T_TINKERBELL_IMAGE_UBUNTU_1_28"
	tinkerbellImageUbuntu129EnvVar                      = "T_TINKERBELL_IMAGE_UBUNTU_1_29"
	tinkerbellImageUbuntu130EnvVar                      = "T_TINKERBELL_IMAGE_UBUNTU_1_30"
	tinkerbellImageUbuntu131EnvVar                      = "T_TINKERBELL_IMAGE_UBUNTU_1_31"
	tinkerbellImageUbuntu2204Kubernetes127EnvVar        = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_27"
	tinkerbellImageUbuntu2204Kubernetes128EnvVar        = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_28"
	tinkerbellImageUbuntu2204Kubernetes129EnvVar        = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_29"
	tinkerbellImageUbuntu2204Kubernetes129RTOSEnvVar    = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_29_RTOS"
	tinkerbellImageUbuntu2204Kubernetes129GenericEnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_29_GENERIC"
	tinkerbellImageUbuntu2204Kubernetes130EnvVar        = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_30"
	tinkerbellImageUbuntu2204Kubernetes131EnvVar        = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_31"
	tinkerbellImageRedHat127EnvVar                      = "T_TINKERBELL_IMAGE_REDHAT_1_27"
	tinkerbellImageRedHat128EnvVar                      = "T_TINKERBELL_IMAGE_REDHAT_1_28"
	tinkerbellImageRedHat129EnvVar                      = "T_TINKERBELL_IMAGE_REDHAT_1_29"
	tinkerbellImageRedHat130EnvVar                      = "T_TINKERBELL_IMAGE_REDHAT_1_30"
	tinkerbellImageRedHat131EnvVar                      = "T_TINKERBELL_IMAGE_REDHAT_1_31"
	tinkerbellImageRedHat9127EnvVar                     = "T_TINKERBELL_IMAGE_REDHAT_9_1_27"
	tinkerbellImageRedHat9128EnvVar                     = "T_TINKERBELL_IMAGE_REDHAT_9_1_28"
	tinkerbellImageRedHat9129EnvVar                     = "T_TINKERBELL_IMAGE_REDHAT_9_1_29"
	tinkerbellImageRedHat9130EnvVar                     = "T_TINKERBELL_IMAGE_REDHAT_9_1_30"
	tinkerbellImageRedHat9131EnvVar                     = "T_TINKERBELL_IMAGE_REDHAT_9_1_31"
	tinkerbellInventoryCsvFilePathEnvVar                = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey                          = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
	tinkerbellCIEnvironmentEnvVar                       = "T_TINKERBELL_CI_ENVIRONMENT"
	controlPlaneIdentifier                              = "cp"
	workerIdentifier                                    = "worker"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellControlPlaneNetworkCidrEnvVar,
	tinkerbellImageUbuntu127EnvVar,
	tinkerbellImageUbuntu128EnvVar,
	tinkerbellImageUbuntu129EnvVar,
	tinkerbellImageUbuntu130EnvVar,
	tinkerbellImageUbuntu131EnvVar,
	tinkerbellImageUbuntu2204Kubernetes127EnvVar,
	tinkerbellImageUbuntu2204Kubernetes128EnvVar,
	tinkerbellImageUbuntu2204Kubernetes129EnvVar,
	tinkerbellImageUbuntu2204Kubernetes129RTOSEnvVar,
	tinkerbellImageUbuntu2204Kubernetes129GenericEnvVar,
	tinkerbellImageUbuntu2204Kubernetes130EnvVar,
	tinkerbellImageUbuntu2204Kubernetes131EnvVar,
	tinkerbellImageRedHat127EnvVar,
	tinkerbellImageRedHat128EnvVar,
	tinkerbellImageRedHat129EnvVar,
	tinkerbellImageRedHat130EnvVar,
	tinkerbellImageRedHat131EnvVar,
	tinkerbellImageRedHat9127EnvVar,
	tinkerbellImageRedHat9128EnvVar,
	tinkerbellImageRedHat9129EnvVar,
	tinkerbellImageRedHat9130EnvVar,
	tinkerbellImageRedHat9131EnvVar,
	tinkerbellInventoryCsvFilePathEnvVar,
	tinkerbellSSHAuthorizedKey,
}

func RequiredTinkerbellEnvVars() []string {
	return requiredTinkerbellEnvVars
}

type TinkerbellOpt func(*Tinkerbell)

type Tinkerbell struct {
	t                    *testing.T
	fillers              []api.TinkerbellFiller
	clusterFillers       []api.ClusterFiller
	serverIP             string
	cidr                 string
	inventoryCsvFilePath string
}

// UpdateTinkerbellMachineSSHAuthorizedKey updates a tinkerbell machine configs SSHAuthorizedKey.
func UpdateTinkerbellMachineSSHAuthorizedKey() api.TinkerbellMachineFiller {
	return api.WithStringFromEnvVarTinkerbellMachineFiller(tinkerbellSSHAuthorizedKey, api.WithSSHAuthorizedKeyForTinkerbellMachineConfig)
}

func NewTinkerbell(t *testing.T, opts ...TinkerbellOpt) *Tinkerbell {
	checkRequiredEnvVars(t, requiredTinkerbellEnvVars)
	cidr := os.Getenv(tinkerbellControlPlaneNetworkCidrEnvVar)

	serverIP, err := GetIP(cidr, ClusterIPPoolEnvVar)
	if err != nil {
		t.Fatalf("failed to get tinkerbell ip for test environment: %v", err)
	}

	tink := &Tinkerbell{
		t: t,
		fillers: []api.TinkerbellFiller{
			api.WithTinkerbellServer(serverIP),
			api.WithStringFromEnvVarTinkerbell(tinkerbellSSHAuthorizedKey, api.WithSSHAuthorizedKeyForAllTinkerbellMachines),
			api.WithHardwareSelectorLabels(),
		},
	}

	tink.serverIP = serverIP

	tink.cidr = cidr
	tink.inventoryCsvFilePath = os.Getenv(tinkerbellInventoryCsvFilePathEnvVar)

	for _, opt := range opts {
		opt(tink)
	}

	return tink
}

func (t *Tinkerbell) Name() string {
	return tinkerbellProviderName
}

func (t *Tinkerbell) Setup() {}

// UpdateKubeConfig customizes generated kubeconfig for the provider.
func (t *Tinkerbell) UpdateKubeConfig(content *[]byte, clusterName string) error {
	return nil
}

// ClusterConfigUpdates satisfies the test framework Provider.
func (t *Tinkerbell) ClusterConfigUpdates() []api.ClusterConfigFiller {
	clusterIP, err := GetIP(t.cidr, ClusterIPPoolEnvVar)
	if err != nil {
		t.t.Fatalf("failed to get cluster ip for test environment: %v", err)
	}

	f := make([]api.ClusterFiller, 0, len(t.clusterFillers)+1)
	f = append(f, t.clusterFillers...)
	f = append(f, api.WithControlPlaneEndpointIP(clusterIP))

	return []api.ClusterConfigFiller{api.ClusterToConfigFiller(f...), api.TinkerbellToConfigFiller(t.fillers...)}
}

func (t *Tinkerbell) WithProviderUpgrade(fillers ...api.TinkerbellFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(api.TinkerbellToConfigFiller(fillers...))
	}
}

// CleanupResources runs a clean up the Tinkerbell machines which simply powers them down.
func (t *Tinkerbell) CleanupResources(_ string) error {
	return cleanup.TinkerbellTestResources(t.inventoryCsvFilePath, true)
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right image for all
// tinkerbell machine configs.
func (t *Tinkerbell) WithKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, _ *releasev1.EksARelease, kernelVariant ...string) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
		api.TinkerbellToConfigFiller(
			imageForKubeVersionAndOS(kubeVersion, os, "", kernelVariant...),
			api.WithOsFamilyForAllTinkerbellMachines(osFamiliesForOS[os]),
		),
	)
}

// WithCPKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right image for CP
// tinkerbell machine configs.
func (t *Tinkerbell) WithCPKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.TinkerbellToConfigFiller(
			imageForKubeVersionAndOS(kubeVersion, os, controlPlaneIdentifier),
		),
	)
}

// WithWorkerKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right image for all
// Worker tinkerbell machine configs.
func (t *Tinkerbell) WithWorkerKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.TinkerbellToConfigFiller(
			imageForKubeVersionAndOS(kubeVersion, os, workerIdentifier),
		),
	)
}

// WithNewWorkerNodeGroup returns an api.ClusterFiller that adds a new workerNodeGroupConfiguration and
// a corresponding TinkerbellMachineConfig to the cluster config.
func (t *Tinkerbell) WithNewWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup) api.ClusterConfigFiller {
	// TODO: Implement for Tinkerbell provider
	panic("Not implemented for Tinkerbell yet")
}

func envVarForImage(os OS, kubeVersion anywherev1.KubernetesVersion, kernelVariant ...string) string {
	imageEnvVar := fmt.Sprintf("T_TINKERBELL_IMAGE_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ReplaceAll(string(kubeVersion), ".", "_"))
	if len(kernelVariant) > 0 && kernelVariant[0] != "" {
		imageEnvVar = fmt.Sprintf("%s_%s", imageEnvVar, strings.ToUpper(kernelVariant[0]))
	}
	return imageEnvVar
}

// withKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right image for all
// tinkerbell machine configs.
func withKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, machineConfigType string, release *releasev1.EksARelease) TinkerbellOpt {
	if machineConfigType == controlPlaneIdentifier || machineConfigType == workerIdentifier {
		return func(t *Tinkerbell) {
			t.fillers = append(t.fillers,
				imageForKubeVersionAndOS(kubeVersion, os, machineConfigType),
			)
		}
	}
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			imageForKubeVersionAndOS(kubeVersion, os, ""),
			api.WithOsFamilyForAllTinkerbellMachines(osFamiliesForOS[os]),
		)
	}
}

// WithUbuntu127Tinkerbell tink test with ubuntu 1.27.
func WithUbuntu127Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, "", nil)
}

// WithUbuntu128Tinkerbell tink test with ubuntu 1.28.
func WithUbuntu128Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, "", nil)
}

// WithUbuntu129Tinkerbell tink test with ubuntu 1.29.
func WithUbuntu129Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, "", nil)
}

// WithUbuntu130Tinkerbell tink test with ubuntu 1.30.
func WithUbuntu130Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, "", nil)
}

// WithUbuntu131Tinkerbell tink test with ubuntu 1.31.
func WithUbuntu131Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, "", nil)
}

// WithRedHat127Tinkerbell tink test with redhat 1.27.
func WithRedHat127Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube127, RedHat8, "", nil)
}

// WithRedHat128Tinkerbell tink test with redhat 1.28.
func WithRedHat128Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube128, RedHat8, "", nil)
}

// WithRedHat129Tinkerbell tink test with redhat 1.29.
func WithRedHat129Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube129, RedHat8, "", nil)
}

// WithRedHat130Tinkerbell tink test with redhat 1.30.
func WithRedHat130Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube130, RedHat8, "", nil)
}

// WithRedHat131Tinkerbell tink test with redhat 1.31.
func WithRedHat131Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube131, RedHat8, "", nil)
}

// WithRedHat9127Tinkerbell tink test with redhat9 efi 1.27.
func WithRedHat9127Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube127, RedHat9, "", nil)
}

// WithRedHat9128Tinkerbell tink test with redhat9 efi 1.28.
func WithRedHat9128Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube128, RedHat9, "", nil)
}

// WithRedHat9129Tinkerbell tink test with redhat9 efi 1.29.
func WithRedHat9129Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube129, RedHat9, "", nil)
}

// WithRedHat9130Tinkerbell tink test with redhat9 efi 1.30.
func WithRedHat9130Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube130, RedHat9, "", nil)
}

// WithRedHat9131Tinkerbell tink test with redhat9 efi 1.31.
func WithRedHat9131Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube131, RedHat9, "", nil)
}

func WithBottleRocketTinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Bottlerocket),
		)
	}
}

func WithTinkerbellExternalEtcdTopology(count int) TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append([]api.TinkerbellFiller{api.WithTinkerbellEtcdMachineConfig()}, t.fillers...)
		t.clusterFillers = append(t.clusterFillers, api.WithExternalEtcdTopology(count), api.WithExternalEtcdMachineRef(anywherev1.TinkerbellMachineConfigKind))
	}
}

func WithCustomTinkerbellMachineConfig(selector string) TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append([]api.TinkerbellFiller{api.WithCustomTinkerbellMachineConfig(selector)}, t.fillers...)
	}
}

// ClusterStateValidations returns a list of provider specific validations.
func (t *Tinkerbell) ClusterStateValidations() []clusterf.StateValidation {
	return []clusterf.StateValidation{}
}

// WithOSImageURL Modify OS Image url.
func WithOSImageURL(url string) TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithTinkerbellOSImageURL(url),
		)
	}
}

// WithHookImagesURLPath Modify Hook OS Image url.
func WithHookImagesURLPath(url string) TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithHookImagesURLPath(url),
		)
	}
}

// imageForKubeVersionAndOS sets osImageURL on the appropriate field in the Machine Config based on the machineConfigType string provided else sets it at Data Center config.
func imageForKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, operatingSystem OS, machineConfigType string, kernelVariant ...string) api.TinkerbellFiller {
	var tinkerbellFiller api.TinkerbellFiller
	if machineConfigType == workerIdentifier {
		tinkerbellFiller = api.WithTinkerbellWorkerMachineConfigOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion, kernelVariant...)), osFamiliesForOS[operatingSystem])
	} else if machineConfigType == controlPlaneIdentifier {
		tinkerbellFiller = api.WithTinkerbellCPMachineConfigOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion, kernelVariant...)), osFamiliesForOS[operatingSystem])
	} else {
		tinkerbellFiller = api.WithTinkerbellOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion, kernelVariant...)))
	}
	return tinkerbellFiller
}

// Ubuntu127Image represents an Ubuntu raw image corresponding to Kubernetes 1.27.
func Ubuntu127Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, "")
}

// Ubuntu128Image represents an Ubuntu raw image corresponding to Kubernetes 1.28.
func Ubuntu128Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, "")
}

// Ubuntu129Image represents an Ubuntu raw image corresponding to Kubernetes 1.29.
func Ubuntu129Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, "")
}

// Ubuntu130Image represents an Ubuntu raw image corresponding to Kubernetes 1.30.
func Ubuntu130Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, "")
}

// Ubuntu131Image represents an Ubuntu raw image corresponding to Kubernetes 1.31.
func Ubuntu131Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, "")
}

// Ubuntu127ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.27 and is set for CP machine config.
func Ubuntu127ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu128ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for CP machine config.
func Ubuntu128ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu129ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.29 and is set for CP machine config.
func Ubuntu129ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu130ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.30 and is set for CP machine config.
func Ubuntu130ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu131ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.31 and is set for CP machine config.
func Ubuntu131ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu127ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.27 and is set for worker machine config.
func Ubuntu127ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, workerIdentifier)
}

// Ubuntu128ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for worker machine config.
func Ubuntu128ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, workerIdentifier)
}

// Ubuntu129ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.29 and is set for worker machine config.
func Ubuntu129ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2004, workerIdentifier)
}

// Ubuntu130ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.30 and is set for worker machine config.
func Ubuntu130ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube130, Ubuntu2004, workerIdentifier)
}

// Ubuntu131ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.31 and is set for worker machine config.
func Ubuntu131ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube131, Ubuntu2004, workerIdentifier)
}

// Ubuntu2204Kubernetes127Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.27.
func Ubuntu2204Kubernetes127Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes128Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.28.
func Ubuntu2204Kubernetes128Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes129Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.29.
func Ubuntu2204Kubernetes129Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes129RTOSImage represents an Ubuntu 22.04 RTOS raw image corresponding to Kubernetes 1.29.
func Ubuntu2204Kubernetes129RTOSImage() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2204, "", "rtos")
}

// Ubuntu2204Kubernetes129GenericImage represents an Ubuntu 22.04 Generic raw image corresponding to Kubernetes 1.29.
func Ubuntu2204Kubernetes129GenericImage() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube129, Ubuntu2204, "", "generic")
}

// Ubuntu2204Kubernetes130Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.30.
func Ubuntu2204Kubernetes130Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube130, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes131Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.31.
func Ubuntu2204Kubernetes131Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube131, Ubuntu2204, "")
}
