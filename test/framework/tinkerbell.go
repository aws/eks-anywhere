package framework

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	tinkerbellProviderName                       = "tinkerbell"
	tinkerbellBootstrapIPEnvVar                  = "T_TINKERBELL_BOOTSTRAP_IP"
	tinkerbellControlPlaneNetworkCidrEnvVar      = "T_TINKERBELL_CP_NETWORK_CIDR"
	tinkerbellImageUbuntu124EnvVar               = "T_TINKERBELL_IMAGE_UBUNTU_1_24"
	tinkerbellImageUbuntu125EnvVar               = "T_TINKERBELL_IMAGE_UBUNTU_1_25"
	tinkerbellImageUbuntu126EnvVar               = "T_TINKERBELL_IMAGE_UBUNTU_1_26"
	tinkerbellImageUbuntu127EnvVar               = "T_TINKERBELL_IMAGE_UBUNTU_1_27"
	tinkerbellImageUbuntu128EnvVar               = "T_TINKERBELL_IMAGE_UBUNTU_1_28"
	tinkerbellImageUbuntu2204Kubernetes124EnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_24"
	tinkerbellImageUbuntu2204Kubernetes125EnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_25"
	tinkerbellImageUbuntu2204Kubernetes126EnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_26"
	tinkerbellImageUbuntu2204Kubernetes127EnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_27"
	tinkerbellImageUbuntu2204Kubernetes128EnvVar = "T_TINKERBELL_IMAGE_UBUNTU_2204_1_28"
	tinkerbellImageRedHat124EnvVar               = "T_TINKERBELL_IMAGE_REDHAT_1_24"
	tinkerbellImageRedHat125EnvVar               = "T_TINKERBELL_IMAGE_REDHAT_1_25"
	tinkerbellImageRedHat126EnvVar               = "T_TINKERBELL_IMAGE_REDHAT_1_26"
	tinkerbellImageRedHat127EnvVar               = "T_TINKERBELL_IMAGE_REDHAT_1_27"
	tinkerbellImageRedHat128EnvVar               = "T_TINKERBELL_IMAGE_REDHAT_1_28"
	tinkerbellInventoryCsvFilePathEnvVar         = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey                   = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
	TinkerbellCIEnvironment                      = "T_TINKERBELL_CI_ENVIRONMENT"
	controlPlaneIdentifier                       = "cp"
	workerIdentifier                             = "worker"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellControlPlaneNetworkCidrEnvVar,
	tinkerbellImageUbuntu124EnvVar,
	tinkerbellImageUbuntu125EnvVar,
	tinkerbellImageUbuntu126EnvVar,
	tinkerbellImageUbuntu127EnvVar,
	tinkerbellImageUbuntu128EnvVar,
	tinkerbellImageUbuntu2204Kubernetes124EnvVar,
	tinkerbellImageUbuntu2204Kubernetes125EnvVar,
	tinkerbellImageUbuntu2204Kubernetes126EnvVar,
	tinkerbellImageUbuntu2204Kubernetes127EnvVar,
	tinkerbellImageUbuntu2204Kubernetes128EnvVar,
	tinkerbellImageRedHat124EnvVar,
	tinkerbellImageRedHat125EnvVar,
	tinkerbellImageRedHat126EnvVar,
	tinkerbellImageRedHat127EnvVar,
	tinkerbellImageRedHat128EnvVar,
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

func (t *Tinkerbell) CleanupVMs(_ string) error {
	return nil
}

// WithKubeVersionAndOS returns a cluster config filler that sets the cluster kube version and the right image for all
// tinkerbell machine configs.
func (t *Tinkerbell) WithKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, os OS, release *releasev1.EksARelease) api.ClusterConfigFiller {
	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithKubernetesVersion(kubeVersion)),
		api.TinkerbellToConfigFiller(
			imageForKubeVersionAndOS(kubeVersion, os, ""),
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

func envVarForImage(os OS, kubeVersion anywherev1.KubernetesVersion) string {
	imageEnvVar := fmt.Sprintf("T_TINKERBELL_IMAGE_%s_%s", strings.ToUpper(strings.ReplaceAll(string(os), "-", "_")), strings.ReplaceAll(string(kubeVersion), ".", "_"))
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

// WithUbuntu125Tinkerbell tink test with ubuntu 1.25.
func WithUbuntu125Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube125, Ubuntu2004, "", nil)
}

// WithUbuntu126Tinkerbell tink test with ubuntu 1.26.
func WithUbuntu126Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube126, Ubuntu2004, "", nil)
}

// WithUbuntu127Tinkerbell tink test with ubuntu 1.27.
func WithUbuntu127Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, "", nil)
}

// WithUbuntu128Tinkerbell tink test with ubuntu 1.28.
func WithUbuntu128Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, "", nil)
}

// WithRedHat125Tinkerbell tink test with redhat 1.25.
func WithRedHat125Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube125, RedHat8, "", nil)
}

// WithRedHat126Tinkerbell tink test with redhat 1.26.
func WithRedHat126Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube126, RedHat8, "", nil)
}

// WithRedHat127Tinkerbell tink test with redhat 1.27.
func WithRedHat127Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube127, RedHat8, "", nil)
}

// WithRedHat128Tinkerbell tink test with redhat 1.27.
func WithRedHat128Tinkerbell() TinkerbellOpt {
	return withKubeVersionAndOS(anywherev1.Kube128, RedHat8, "", nil)
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
func imageForKubeVersionAndOS(kubeVersion anywherev1.KubernetesVersion, operatingSystem OS, machineConfigType string) api.TinkerbellFiller {
	var tinkerbellFiller api.TinkerbellFiller
	if machineConfigType == workerIdentifier {
		tinkerbellFiller = api.WithTinkerbellWorkerMachineConfigOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion)), osFamiliesForOS[operatingSystem])
	} else if machineConfigType == controlPlaneIdentifier {
		tinkerbellFiller = api.WithTinkerbellCPMachineConfigOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion)), osFamiliesForOS[operatingSystem])
	} else {
		tinkerbellFiller = api.WithTinkerbellOSImageURL(os.Getenv(envVarForImage(operatingSystem, kubeVersion)))
	}
	return tinkerbellFiller
}

// Ubuntu125Image represents an Ubuntu raw image corresponding to Kubernetes 1.25.
func Ubuntu125Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube125, Ubuntu2004, "")
}

// Ubuntu126Image represents an Ubuntu raw image corresponding to Kubernetes 1.26.
func Ubuntu126Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube126, Ubuntu2004, "")
}

// Ubuntu127Image represents an Ubuntu raw image corresponding to Kubernetes 1.27.
func Ubuntu127Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, "")
}

// Ubuntu128Image represents an Ubuntu raw image corresponding to Kubernetes 1.28.
func Ubuntu128Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, "")
}

// Ubuntu126ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for CP machine config.
func Ubuntu126ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube126, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu127ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.27 and is set for CP machine config.
func Ubuntu127ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu128ImageForCP represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for CP machine config.
func Ubuntu128ImageForCP() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, controlPlaneIdentifier)
}

// Ubuntu126ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for worker machine config.
func Ubuntu126ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube126, Ubuntu2004, workerIdentifier)
}

// Ubuntu127ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.27 and is set for worker machine config.
func Ubuntu127ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2004, workerIdentifier)
}

// Ubuntu128ImageForWorker represents an Ubuntu raw image corresponding to Kubernetes 1.28 and is set for worker machine config.
func Ubuntu128ImageForWorker() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2004, workerIdentifier)
}

// Ubuntu2204Kubernetes126Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.26.
func Ubuntu2204Kubernetes126Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube126, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes127Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.27.
func Ubuntu2204Kubernetes127Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube127, Ubuntu2204, "")
}

// Ubuntu2204Kubernetes128Image represents an Ubuntu 22.04 raw image corresponding to Kubernetes 1.28.
func Ubuntu2204Kubernetes128Image() api.TinkerbellFiller {
	return imageForKubeVersionAndOS(anywherev1.Kube128, Ubuntu2204, "")
}
