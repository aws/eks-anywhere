package framework

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	nutanixEndpoint = "NUTANIX_ENDPOINT"
	nutanixPort     = "NUTANIX_PORT"
	nutanixUser     = "NUTANIX_USER"
	nutanixPwd      = "NUTANIX_PASSWORD"
	nutanixInsecure = "NUTANIX_INSECURE"

	nutanixVCPUsPerSocket = "NUTANIX_VCPUS_PER_SOCKET"
)

var requiredNutanixEnvVars = []string{
	nutanixEndpoint,
	nutanixPort,
	nutanixUser,
	nutanixPwd,
	nutanixInsecure,

	nutanixVCPUsPerSocket,
}

type Nutanix struct {
	t              *testing.T
	fillers        []api.NutanixFiller
	clusterFillers []api.ClusterFiller
	cpCidr         string
	podCidr        string
}

type NutanixOpt func(*Nutanix)

func NewNutanix(t *testing.T, opts ...NutanixOpt) *Nutanix {
	checkRequiredEnvVars(t, requiredNutanixEnvVars)
	nutanixProvider := &Nutanix{
		t: t,
		fillers: []api.NutanixFiller{
			api.WithNutanixStringFromEnvVar(nutanixEndpoint, api.WithNutanixEndpoint),
			api.WithNutanixIntFromEnvVar(nutanixPort, api.WithNutanixPort),
			api.WithNutanixStringFromEnvVar(nutanixUser, api.WithNutanixUser),
			api.WithNutanixStringFromEnvVar(nutanixPwd, api.WithNutanixPwd),
			// api.WithNutanixStringFromEnvVar(nutanixInsecure, api.WithNutanixInsure),
			api.WithNutanixInt32FromEnvVar(nutanixVCPUsPerSocket, api.WithNutanixVCPUsPerSocket),
		},
	}

	// s.cpCidr = os.Getenv(nutanixControlPlaneCidr)
	// s.podCidr = os.Getenv(nutanixPodCidr)

	for _, opt := range opts {
		opt(nutanixProvider)
	}

	return nutanixProvider
}

func (s *Nutanix) Name() string {
	return "nutanix"
}

func (s *Nutanix) Setup() {}

func (s *Nutanix) CustomizeProviderConfig(file string) []byte {
	return s.customizeProviderConfig(file, s.fillers...)
}

func (s *Nutanix) ClusterConfigFillers() []api.ClusterFiller {
	// ip, err := GenerateUniqueIp(s.cpCidr)
	// if err != nil {
	// 	s.t.Fatalf("failed to generate control plane ip for nutanix [cidr=%s]: %v", s.cpCidr, err)
	// }
	// s.clusterFillers = append(s.clusterFillers, api.WithControlPlaneEndpointIP(ip))

	// if s.podCidr != "" {
	// 	s.clusterFillers = append(s.clusterFillers, api.WithPodCidr(s.podCidr))
	// }

	return s.clusterFillers
}

func (s *Nutanix) customizeProviderConfig(file string, fillers ...api.NutanixFiller) []byte {
	providerOutput, err := api.AutoFillNutanixProvider(file, fillers...)
	if err != nil {
		s.t.Fatalf("failed to customize provider config from file: %v", err)
	}
	return providerOutput
}

func WithNutanixUbuntu121() NutanixOpt {
	return func(v *Nutanix) {
		v.fillers = append(v.fillers,
			api.WithNutanixInt32FromEnvVar(nutanixVCPUsPerSocket, api.WithNutanixVCPUsPerSocket),
		)
	}
}
