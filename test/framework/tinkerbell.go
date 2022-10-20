package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	TinkerbellProviderName                  = "tinkerbell"
	tinkerbellBootstrapIPEnvVar             = "T_TINKERBELL_BOOTSTRAP_IP"
	tinkerbellControlPlaneNetworkCidrEnvVar = "T_TINKERBELL_CP_NETWORK_CIDR"
	tinkerbellImageUbuntu121EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_21"
	tinkerbellImageUbuntu122EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_22"
	tinkerbellImageUbuntu123EnvVar          = "T_TINKERBELL_IMAGE_UBUNTU_1_23"
	tinkerbellImageRedHat121EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_21"
	tinkerbellImageRedHat122EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_22"
	tinkerbellImageRedHat123EnvVar          = "T_TINKERBELL_IMAGE_REDHAT_1_23"
	tinkerbellInventoryCsvFilePathEnvVar    = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey              = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
	TinkerbellCIEnvironment                 = "T_TINKERBELL_CI_ENVIRONMENT"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellControlPlaneNetworkCidrEnvVar,
	tinkerbellImageUbuntu121EnvVar,
	tinkerbellImageUbuntu122EnvVar,
	tinkerbellImageUbuntu123EnvVar,
	tinkerbellImageRedHat123EnvVar,
	tinkerbellImageRedHat121EnvVar,
	tinkerbellImageRedHat122EnvVar,
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

func UpdateTinkerbellUbuntuTemplate121Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu121EnvVar, api.WithTinkerbellOSImageURL)
}

func UpdateTinkerbellUbuntuTemplate122Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu122EnvVar, api.WithTinkerbellOSImageURL)
}

func UpdateTinkerbellUbuntuTemplate123Var() api.TinkerbellFiller {
	return api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu123EnvVar, api.WithTinkerbellOSImageURL)
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

func (t *Tinkerbell) CustomizeProviderConfig(file string) []byte {
	return t.customizeProviderConfig(file, t.fillers...)
}

func (t *Tinkerbell) WithProviderUpgrade(fillers ...api.TinkerbellFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = t.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (t *Tinkerbell) CleanupVMs(_ string) error {
	return nil
}

func (t *Tinkerbell) customizeProviderConfig(file string, fillers ...api.TinkerbellFiller) []byte {
	providerOutput, err := api.AutoFillTinkerbellProvider(file, fillers...)
	if err != nil {
		t.t.Fatalf("failed to customize provider config from file: %v", err)
	}
	return providerOutput
}

func (t *Tinkerbell) ClusterConfigFillers() []api.ClusterFiller {
	clusterIP, err := GetIP(t.cidr, ClusterIPPoolEnvVar)
	if err != nil {
		t.t.Fatalf("failed to get cluster ip for test environment: %v", err)
	}

	t.clusterFillers = append(t.clusterFillers, api.WithControlPlaneEndpointIP(clusterIP))

	return t.clusterFillers
}

func WithUbuntu121Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu121EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
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

// WithRedHat121Tinkerbell tink test with redhat 1.21.
func WithRedHat121Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageRedHat121EnvVar, api.WithTinkerbellOSImageURL),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.RedHat),
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
