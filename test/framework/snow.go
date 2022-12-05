package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	snowAMIIDUbuntu121   = "T_SNOW_AMIID_UBUNTU_1_21"
	snowAMIIDUbuntu122   = "T_SNOW_AMIID_UBUNTU_1_22"
	snowAMIIDUbuntu123   = "T_SNOW_AMIID_UBUNTU_1_23"
	snowDevices          = "T_SNOW_DEVICES"
	snowControlPlaneCidr = "T_SNOW_CONTROL_PLANE_CIDR"
	snowPodCidr          = "T_SNOW_POD_CIDR"
	snowCredentialsFile  = "EKSA_AWS_CREDENTIALS_FILE"
	snowCertificatesFile = "EKSA_AWS_CA_BUNDLES_FILE"
)

var requiredSnowEnvVars = []string{
	snowAMIIDUbuntu121,
	snowAMIIDUbuntu122,
	snowAMIIDUbuntu123,
	snowControlPlaneCidr,
	snowCredentialsFile,
	snowCertificatesFile,
}

type Snow struct {
	t              *testing.T
	fillers        []api.SnowFiller
	clusterFillers []api.ClusterFiller
	cpCidr         string
	podCidr        string
}

type SnowOpt func(*Snow)

func NewSnow(t *testing.T, opts ...SnowOpt) *Snow {
	checkRequiredEnvVars(t, requiredSnowEnvVars)
	s := &Snow{
		t: t,
		fillers: []api.SnowFiller{
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu121, api.WithSnowAMIIDForAllMachines),
		},
	}

	s.cpCidr = os.Getenv(snowControlPlaneCidr)
	s.podCidr = os.Getenv(snowPodCidr)

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Snow) Name() string {
	return "snow"
}

func (s *Snow) Setup() {}

func (s *Snow) CustomizeProviderConfig(file string) []byte {
	return s.customizeProviderConfig(file, s.fillers...)
}

func (s *Snow) CleanupVMs(_ string) error {
	return nil
}

func (s *Snow) ClusterConfigFillers() []api.ClusterFiller {
	ip, err := GenerateUniqueIp(s.cpCidr)
	if err != nil {
		s.t.Fatalf("failed to generate control plane ip for snow [cidr=%s]: %v", s.cpCidr, err)
	}
	s.clusterFillers = append(s.clusterFillers, api.WithControlPlaneEndpointIP(ip))

	if s.podCidr != "" {
		s.clusterFillers = append(s.clusterFillers, api.WithPodCidr(s.podCidr))
	}

	return s.clusterFillers
}

func (s *Snow) customizeProviderConfig(file string, fillers ...api.SnowFiller) []byte {
	providerOutput, err := api.AutoFillSnowProvider(file, fillers...)
	if err != nil {
		s.t.Fatalf("failed to customize provider config from file: %v", err)
	}
	return providerOutput
}

func (s *Snow) WithProviderUpgrade(fillers ...api.SnowFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = s.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func WithSnowUbuntu121() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu121, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

func WithSnowUbuntu122() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu122, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

func WithSnowUbuntu123() SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu123, api.WithSnowAMIIDForAllMachines),
			api.WithSnowStringFromEnvVar(snowDevices, api.WithSnowDevicesForAllMachines),
			api.WithOsFamilyForAllSnowMachines(anywherev1.Ubuntu),
		)
	}
}

func WithSnowWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.SnowMachineConfigFiller) SnowOpt {
	return func(s *Snow) {
		s.fillers = append(s.fillers, snowMachineConfig(name, fillers...))

		s.clusterFillers = append(s.clusterFillers, buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup))
	}
}

func (s *Snow) WithNewSnowWorkerNodeGroup(name string, workerNodeGroup *WorkerNodeGroup, fillers ...api.SnowMachineConfigFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = s.customizeProviderConfig(e.ClusterConfigLocation, snowMachineConfig(name, fillers...))
		var err error
		e.ClusterConfigB, err = api.AutoFillClusterFromYaml(e.ClusterConfigB, buildSnowWorkerNodeGroupClusterFiller(name, workerNodeGroup))
		if err != nil {
			e.T.Fatalf("Error filling cluster config: %v", err)
		}
	}
}

func snowMachineConfig(name string, fillers ...api.SnowMachineConfigFiller) api.SnowFiller {
	f := make([]api.SnowMachineConfigFiller, 0, len(fillers)+2)
	f = append(f,
		api.WithSnowMachineDefaultValues(),
		api.WithSnowDevices(os.Getenv(snowDevices)),
	)
	f = append(f, fillers...)

	return api.WithSnowMachineConfig(name, f...)
}

func buildSnowWorkerNodeGroupClusterFiller(machineConfigName string, workerNodeGroup *WorkerNodeGroup) api.ClusterFiller {
	workerNodeGroup.MachineConfigKind = anywherev1.SnowMachineConfigKind
	workerNodeGroup.MachineConfigName = machineConfigName
	return workerNodeGroup.ClusterFiller()
}

func UpdateSnowUbuntuTemplate121Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu121, api.WithSnowAMIIDForAllMachines)
}

func UpdateSnowUbuntuTemplate122Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu122, api.WithSnowAMIIDForAllMachines)
}

func UpdateSnowUbuntuTemplate123Var() api.SnowFiller {
	return api.WithSnowStringFromEnvVar(snowAMIIDUbuntu123, api.WithSnowAMIIDForAllMachines)
}
