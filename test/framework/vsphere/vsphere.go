package vsphere

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api/vsphere"
	"github.com/aws/eks-anywhere/test/framework"

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
	fillers    []vsphere.VSphereFiller
	cidr       string
	GovcClient *executables.Govc
}

type VSphereOpt func(*VSphere)

func UpdateUbuntuTemplate118Var() vsphere.VSphereFiller {
	return vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu118Var, vsphere.WithTemplate)
}

func UpdateUbuntuTemplate119Var() vsphere.VSphereFiller {
	return vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, vsphere.WithTemplate)
}

func UpdateUbuntuTemplate120Var() vsphere.VSphereFiller {
	return vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu120Var, vsphere.WithTemplate)
}

func UpdateUbuntuTemplate121Var() vsphere.VSphereFiller {
	return vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu121Var, vsphere.WithTemplate)
}

func UpdateBottlerocketTemplate121() vsphere.VSphereFiller {
	return vsphere.WithStringFromEnvVar(vsphereTemplateBR121Var, vsphere.WithTemplate)
}

func NewVSphere(t *testing.T, opts ...VSphereOpt) *VSphere {
	framework.CheckRequiredEnvVars(t, requiredEnvVars)
	c := framework.BuildGovc(t)
	v := &VSphere{
		t:          t,
		GovcClient: c,
		fillers: []vsphere.VSphereFiller{
			vsphere.WithStringFromEnvVar(vsphereDatacenterVar, vsphere.WithDatacenter),
			vsphere.WithStringFromEnvVar(vsphereDatastoreVar, vsphere.WithDatastore),
			vsphere.WithStringFromEnvVar(vsphereFolderVar, vsphere.WithFolder),
			vsphere.WithStringFromEnvVar(vsphereNetworkVar, vsphere.WithNetwork),
			vsphere.WithStringFromEnvVar(vsphereResourcePoolVar, vsphere.WithResourcePool),
			vsphere.WithStringFromEnvVar(vsphereServerVar, vsphere.WithServer),
			vsphere.WithStringFromEnvVar(vsphereSshAuthorizedKeyVar, vsphere.WithSSHAuthorizedKey),
			vsphere.WithStringFromEnvVar(vsphereStoragePolicyNameVar, vsphere.WithStoragePolicyName),
			vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, vsphere.WithTemplate),
			vsphere.WithBoolFromEnvVar(vsphereTlsInsecureVar, vsphere.WithTLSInsecure),
			vsphere.WithStringFromEnvVar(vsphereTlsThumbprintVar, vsphere.WithTLSThumbprint),
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
			vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu121Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu120() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu120Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu119() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu119Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithUbuntu118() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vsphereTemplateUbuntu118Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Ubuntu),
		)
	}
}

func WithBottleRocket120() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vsphereTemplateBR120Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Bottlerocket),
		)
	}
}

func WithBottleRocket121() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vsphereTemplateBR121Var, vsphere.WithTemplate),
			vsphere.WithOsFamily(v1alpha1.Bottlerocket),
		)
	}
}

func WithPrivateNetwork() VSphereOpt {
	return func(v *VSphere) {
		v.fillers = append(v.fillers,
			vsphere.WithStringFromEnvVar(vspherePrivateNetworkVar, vsphere.WithNetwork),
		)
		v.cidr = os.Getenv(privateNetworkCidrVar)
	}
}

func WithVSphereFillers(fillers ...vsphere.VSphereFiller) VSphereOpt {
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

func (v *VSphere) customizeProviderConfig(file string, fillers ...vsphere.VSphereFiller) []byte {
	providerOutput, err := vsphere.AutoFillVSphereProvider(file, fillers...)
	if err != nil {
		v.t.Fatalf("Error customizing provider config from file: %v", err)
	}
	return providerOutput
}

func (v *VSphere) WithProviderUpgrade(fillers ...vsphere.VSphereFiller) framework.ClusterE2ETestOpt {
	return func(e *framework.ClusterE2ETest) {
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
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip := os.Getenv(vsphereHost)
	if len(ip) > 0 && ipgen.IsIPUnique(ip) {
		logger.V(1).Info("Using configured ip: " + ip)
		return ip
	}
	logger.V(1).Info("Generating unique IP for vsphere control plane")
	ip, err := ipgen.GenerateUniqueIP(v.cidr)
	if err != nil {
		v.t.Fatalf("Error getting unique IP for vsphere: %v", err)
	}
	logger.V(2).Info("IP generated successfully", "ip", ip)
	return ip
}
