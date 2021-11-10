package framework

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
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
	vsphereTemplateUbuntu118Var = "T_VSPHERE_TEMPLATE_UBUNTU_1_18"
	vsphereTemplateUbuntu119Var = "T_VSPHERE_TEMPLATE_UBUNTU_1_19"
	vsphereTemplateUbuntu120Var = "T_VSPHERE_TEMPLATE_UBUNTU_1_20"
	vsphereTemplateUbuntu121Var = "T_VSPHERE_TEMPLATE_UBUNTU_1_21"
	vsphereTemplateBR120Var     = "T_VSPHERE_TEMPLATE_BR_1_20"
	vsphereTemplateBR121Var     = "T_VSPHERE_TEMPLATE_BR_1_21"
	vsphereTlsInsecureVar       = "T_VSPHERE_TLS_INSECURE"
	vsphereTlsThumbprintVar     = "T_VSPHERE_TLS_THUMBPRINT"
	vsphereHost                 = "T_VSPHERE_HOST"
	vsphereUsernameVar          = "EKSA_VSPHERE_USERNAME"
	vspherePasswordVar          = "EKSA_VSPHERE_PASSWORD"
	cidrVar                     = "T_VSPHERE_CIDR"
	privateNetworkCidrVar       = "T_VSPHERE_PRIVATE_NETWORK_CIDR"
	govcUrlVar                  = "GOVC_URL"
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
	vsphereTemplateUbuntu118Var,
	vsphereTemplateUbuntu119Var,
	vsphereTemplateUbuntu120Var,
	vsphereTemplateUbuntu121Var,
	vsphereTemplateBR120Var,
	vsphereTemplateBR121Var,
	vsphereTlsInsecureVar,
	vsphereTlsThumbprintVar,
	vsphereUsernameVar,
	vspherePasswordVar,
	cidrVar,
	privateNetworkCidrVar,
	govcUrlVar,
}

type VSphere struct {
	t          *testing.T
	fillers    []api.VSphereFiller
	cidr       string
	GovcClient *executables.Govc
}

type VSphereOpt func(*VSphere)

func UpdateUbuntuTemplate118Var() api.VSphereFiller {
	return api.WithStringFromEnvVar(vsphereTemplateUbuntu118Var, api.WithTemplate)
}

func UpdateUbuntuTemplate119Var() api.VSphereFiller {
	return api.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, api.WithTemplate)
}

func UpdateUbuntuTemplate120Var() api.VSphereFiller {
	return api.WithStringFromEnvVar(vsphereTemplateUbuntu120Var, api.WithTemplate)
}

func UpdateUbuntuTemplate121Var() api.VSphereFiller {
	return api.WithStringFromEnvVar(vsphereTemplateUbuntu121Var, api.WithTemplate)
}

func UpdateBottlerocketTemplate121() api.VSphereFiller {
	return api.WithStringFromEnvVar(vsphereTemplateBR121Var, api.WithTemplate)
}

func NewVSphere(t *testing.T, opts ...VSphereOpt) *VSphere {
	checkRequiredEnvVars(t, requiredEnvVars)
	c := buildGovc(t)
	v := &VSphere{
		t:          t,
		GovcClient: c,
		fillers: []api.VSphereFiller{
			api.WithStringFromEnvVar(vsphereDatacenterVar, api.WithDatacenter),
			api.WithStringFromEnvVar(vsphereDatastoreVar, api.WithDatastore),
			api.WithStringFromEnvVar(vsphereFolderVar, api.WithFolder),
			api.WithStringFromEnvVar(vsphereNetworkVar, api.WithNetwork),
			api.WithStringFromEnvVar(vsphereResourcePoolVar, api.WithResourcePool),
			api.WithStringFromEnvVar(vsphereServerVar, api.WithServer),
			api.WithStringFromEnvVar(vsphereSshAuthorizedKeyVar, api.WithSSHAuthorizedKey),
			api.WithStringFromEnvVar(vsphereStoragePolicyNameVar, api.WithStoragePolicyName),
			api.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, api.WithTemplate),
			api.WithBoolFromEnvVar(vsphereTlsInsecureVar, api.WithTLSInsecure),
			api.WithStringFromEnvVar(vsphereTlsThumbprintVar, api.WithTLSThumbprint),
		},
	}

	v.cidr = os.Getenv(cidrVar)
	for _, opt := range opts {
		opt(v)
	}

	return v
}

func WithUbuntu121() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateUbuntu121Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu120() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateUbuntu120Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu119() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu118() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateUbuntu118Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithBottleRocket120() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateBR120Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Bottlerocket),
		)
	}
}

func WithBottleRocket121() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vsphereTemplateBR121Var, api.WithTemplate),
			api.WithOsFamily(v1alpha1.Bottlerocket),
		)
	}
}

func WithPrivateNetwork() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			api.WithStringFromEnvVar(vspherePrivateNetworkVar, api.WithNetwork),
		)
		v.cidr = os.Getenv(privateNetworkCidrVar)
	}
}

func WithVSphereFillers(fillers ...api.VSphereFiller) VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers, fillers...)
	}
}

func (v *VSphere) Name() string {
	return "vsphere"
}

func (v *VSphere) Setup() {}

func (v *VSphere) CustomizeProviderConfig(file string) []byte {
	return v.customizeProviderConfig(file, v.fillers...)
}

func (v *VSphere) customizeProviderConfig(file string, fillers ...api.VSphereFiller) []byte {
	providerOutput, err := api.AutoFillVSphereProvider(file, fillers...)
	if err != nil {
		v.t.Fatalf("Error customizing provider config from file: %v", err)
	}
	return providerOutput
}

func (v *VSphere) WithProviderUpgrade(fillers ...api.VSphereFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ProviderConfigB = v.customizeProviderConfig(e.ClusterConfigLocation, fillers...)
	}
}

func (v *VSphere) ClusterConfigFillers() []api.ClusterFiller {
	return []api.ClusterFiller{api.WithControlPlaneEndpointIP(v.generateUniqueIp())}
}

func RequiredVsphereEnvVars() []string {
	return requiredEnvVars
}

func (v *VSphere) generateUniqueIp() string {
	ip := os.Getenv(vsphereHost)
	if len(ip) > 0 {
		logger.V(1).Info("Using configured ip: " + ip)
		return ip
	}
	logger.V(1).Info("Generating unique IP for vsphere control plane")
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(v.cidr)
	if err != nil {
		v.t.Fatalf("Error getting unique IP for vsphere: %v", err)
	}
	logger.V(2).Info("IP generated successfully", "ip", ip)
	return ip
}
