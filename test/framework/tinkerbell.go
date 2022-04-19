package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	TinkerbellProviderName               = "tinkerbell"
	tinkerbellServerEnvVar               = "T_TINKERBELL_IP"
	tinkerbellNetworkCidrEnvVar          = "T_TINKERBELL_NETWORK_CIDR"
	tinkerbellCertURLEnvVar              = "T_TINKERBELL_CERT_URL"
	tinkerbellHegelURLEnvVar             = "T_TINKERBELL_HEGEL_URL"
	tinkerbellGRPCAuthEnvVar             = "T_TINKERBELL_GRPC_AUTHORITY"
	tinkerbellPBnJGRPCAuthEnvVar         = "T_TINKERBELL_PBNJ_GRPC_AUTHORITY"
	tinkerbellImageUbuntu120EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_20"
	tinkerbellImageUbuntu121EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_21"
	tinkerbellImageUbuntu122EnvVar       = "T_TINKERBELL_IMAGE_UBUNTU_1_22"
	tinkerbellInventoryCsvFilePathEnvVar = "T_TINKERBELL_INVENTORY_CSV"
	tinkerbellSSHAuthorizedKey           = "T_TINKERBELL_SSH_AUTHORIZED_KEY"
)

var requiredTinkerbellEnvVars = []string{
	tinkerbellServerEnvVar,
	tinkerbellNetworkCidrEnvVar,
	tinkerbellCertURLEnvVar,
	tinkerbellHegelURLEnvVar,
	tinkerbellGRPCAuthEnvVar,
	tinkerbellImageUbuntu120EnvVar,
	tinkerbellImageUbuntu121EnvVar,
	tinkerbellInventoryCsvFilePathEnvVar,
	tinkerbellSSHAuthorizedKey,
}

type TinkerbellOpt func(*Tinkerbell)

type Tinkerbell struct {
	t                    *testing.T
	fillers              []api.TinkerbellFiller
	clusterFillers       []api.ClusterFiller
	cidr                 string
	inventoryCsvFilePath string
}

func NewTinkerbell(t *testing.T, opts ...TinkerbellOpt) *Tinkerbell {
	checkRequiredEnvVars(t, requiredTinkerbellEnvVars)
	tink := &Tinkerbell{
		t: t,
		fillers: []api.TinkerbellFiller{
			api.WithStringFromEnvVarTinkerbell(tinkerbellServerEnvVar, api.WithTinkerbellServer),
			api.WithStringFromEnvVarTinkerbell(tinkerbellHegelURLEnvVar, api.WithTinkerbellHegelURL),
			api.WithStringFromEnvVarTinkerbell(tinkerbellCertURLEnvVar, api.WithTinkerbellCertURL),
			api.WithStringFromEnvVarTinkerbell(tinkerbellGRPCAuthEnvVar, api.WithTinkerbellGRPCAuthEndpoint),
			api.WithStringFromEnvVarTinkerbell(tinkerbellPBnJGRPCAuthEnvVar, api.WithTinkerbellPBnJGRPCAuthEndpoint),
			api.WithStringFromEnvVarTinkerbell(tinkerbellSSHAuthorizedKey, api.WithSSHAuthorizedKeyForAllTinkerbellMachines),
		},
	}

	tink.cidr = os.Getenv(tinkerbellNetworkCidrEnvVar)
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
		t.t.Fatalf("failed to generate ip for tinkerbell cidr %s: %v", t.cidr, err)
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
