package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	snowAMIIDUbuntu120   = "T_SNOW_AMIID_UBUNTU_1_20"
	snowAMIIDUbuntu121   = "T_SNOW_AMIID_UBUNTU_1_21"
	snowControlPlaneCidr = "T_SNOW_CONTROL_PLANE_CIDR"
	snowPodCidr          = "T_SNOW_POD_CIDR"
	snowCredentialsFile  = "EKSA_AWS_CREDENTIALS_FILE"
	snowCertificatesFile = "EKSA_AWS_CA_BUNDLES_FILE"
)

var requiredSnowEnvVars = []string{
	snowAMIIDUbuntu120,
	snowAMIIDUbuntu121,
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
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu120, api.WithSnowAMIID),
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

func WithSnowUbuntu121() SnowOpt {
	return func(v *Snow) {
		v.fillers = append(v.fillers,
			api.WithSnowStringFromEnvVar(snowAMIIDUbuntu121, api.WithSnowAMIID),
		)
	}
}
