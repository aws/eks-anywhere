package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

const (
	TinkerbellProviderName                  = "tinkerbell"
	tinkerbellBootstrapIPEnvVar             = "T_TINKERBELL_BOOTSTRAP_IP"
	tinkerbellControlPlaneNetworkCidrEnvVar = "T_TINKERBELL_CP_NETWORK_CIDR"
	tinkerbellImageUbuntu122EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_22"
	tinkerbellImageUbuntu123EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_23"
	tinkerbellImageUbuntu124EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_24"
	tinkerbellImageUbuntu125EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_25"
	tinkerbellImageUbuntu126EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_26"
	tinkerbellImageRedHat122EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_22"
	tinkerbellImageRedHat123EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_23"
	tinkerbellImageRedHat124EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_24"
	tinkerbellImageRedHat125EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_25"
	tinkerbellInventoryCsvFilePathEnvVar    = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey              = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
	TinkerbellCIEnvironment                 = "T_TINKERBELL_CI_ENVIRONMENT"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellControlPlaneNetworkCidrEnvVar,
	tinkerbellImageUbuntu122EnvVar,
	tinkerbellImageUbuntu123EnvVar,
	tinkerbellImageUbuntu124EnvVar,
	tinkerbellImageUbuntu125EnvVar,
	tinkerbellImageUbuntu126EnvVar,
	tinkerbellImageRedHat122EnvVar,
	tinkerbellImageRedHat123EnvVar,
	tinkerbellImageRedHat124EnvVar,
	tinkerbellImageRedHat125EnvVar,
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

func UpdateTinkerbellUbuntuTemplate122Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu122EnvVar, api.WithTinkerbellOSImageURL)
}

func UpdateTinkerbellUbuntuTemplate123Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu123EnvVar, api.WithTinkerbellOSImageURL)
}

// UpdateTinkerbellUbuntuTemplate124Var updates the tinkerbell template.
func UpdateTinkerbellUbuntuTemplate124Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu124EnvVar, api.WithTinkerbellOSImageURL)
}

// UpdateTinkerbellUbuntuTemplate125Var updates the tinkerbell template.
func UpdateTinkerbellUbuntuTemplate125Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu125EnvVar, api.WithTinkerbellOSImageURL)
}

// UpdateTinkerbellUbuntuTemplate126Var updates the tinkerbell template.
func UpdateTinkerbellUbuntuTemplate126Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu126EnvVar, api.WithTinkerbellOSImageURL)
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
	return TinkerbellProviderName
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

func WithUbuntu122Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu122EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

func WithUbuntu123Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu123EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu124Tinkerbell tink test with ubuntu 1.24.
func WithUbuntu124Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu124EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu125Tinkerbell tink test with ubuntu 1.25.
func WithUbuntu125Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu125EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

// WithUbuntu126Tinkerbell tink test with ubuntu 1.26.
func WithUbuntu126Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu126EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

// WithRedHat122Tinkerbell tink test with redhat 1.22.
func WithRedHat122Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageRedHat122EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat123Tinkerbell tink test with redhat 1.23.
func WithRedHat123Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageRedHat123EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat124Tinkerbell tink test with redhat 1.24.
func WithRedHat124Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageRedHat124EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.RedHat),
		)
	}
}

// WithRedHat125Tinkerbell tink test with redhat 1.25.
func WithRedHat125Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageRedHat125EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.RedHat),
		)
	}
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
