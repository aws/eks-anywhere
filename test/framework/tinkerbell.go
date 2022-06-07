package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	TinkerbellProviderName               = "tinkerbell"
	tinkerbellBootstrapIPEnvVar          = "T_TINKERBELL_BOOTSTRAP_IP"
	tinkerbellNetworkCidrEnvVar          = "T_TINKERBELL_NETWORK_CIDR"
	tinkerbellImageUbuntu120EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_20"
	tinkerbellImageUbuntu121EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_21"
	tinkerbellImageUbuntu122EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_22"
	tinkerbellInventoryCsvFilePathEnvVar = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey           = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellNetworkCidrEnvVar,
	tinkerbellImageUbuntu120EnvVar,
	tinkerbellImageUbuntu121EnvVar,
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
	bootstrapIP          string
	cidr                 string
	inventoryCsvFilePath string
}

func NewTinkerbell(t *testing.T, opts ...TinkerbellOpt) *Tinkerbell {
	checkRequiredEnvVars(t, requiredTinkerbellEnvVars)
	cidr := os.Getenv(tinkerbellNetworkCidrEnvVar)

	serverIP, err := GenerateUniqueIp(cidr)
	if err != nil {
		t.Fatalf("failed to generate tinkerbell ip from cidr %s: %v", cidr, err)
	}

	tink := &Tinkerbell{
		t: t,
		fillers: []api.TinkerbellFiller{
			api.WithTinkerbellServer(serverIP),
			api.WithStringFromEnvVarTinkerbell(tinkerbellSSHAuthorizedKey, api.WithSSHAuthorizedKeyForAllTinkerbellMachines),
		},
	}

	tink.serverIP = serverIP
	tink.bootstrapIP = os.Getenv(tinkerbellBootstrapIPEnvVar)
	if tink.bootstrapIP == "" {
		t.Fatalf("tinkerbell bootstrap ip is required!")
	}

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
	clusterIP, err := GenerateUniqueIp(t.cidr)
	if err != nil {
		t.t.Fatalf("failed to generate cluster ip from cidr %s: %v", t.cidr, err)
	}
	t.clusterFillers = append(t.clusterFillers, api.WithControlPlaneEndpointIP(clusterIP))

	return t.clusterFillers
}

func WithUbuntu120Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu120EnvVar, api.WithImageUrlForAllTinkerbellMachines),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

func WithUbuntu121Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu121EnvVar, api.WithImageUrlForAllTinkerbellMachines),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

func WithUbuntu122Tinkerbell() TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append(t.fillers,
			api.WithStringFromEnvVarTinkerbell(tinkerbellImageUbuntu122EnvVar, api.WithImageUrlForAllTinkerbellMachines),
			api.WithOsFamilyForAllTinkerbellMachines(anywherev1.Ubuntu),
		)
	}
}

func WithTinkerbellExternalEtcdTopology(count int) TinkerbellOpt {
	return func(t *Tinkerbell) {
		t.fillers = append([]api.TinkerbellFiller{api.WithTinkerbellEtcdMachineConfig()}, t.fillers...)
		t.clusterFillers = append(t.clusterFillers, api.WithExternalEtcdTopology(count), api.WithExternalEtcdMachineRef(anywherev1.TinkerbellMachineConfigKind))
	}
}
