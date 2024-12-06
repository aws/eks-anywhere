package constants

import "time"

// Namespace constants.
const (
	EksaSystemNamespace                     = "eksa-system"
	EksaDiagnosticsNamespace                = "eksa-diagnostics"
	EksaControllerManagerDeployment         = "eksa-controller-manager"
	CapdSystemNamespace                     = "capd-system"
	CapcSystemNamespace                     = "capc-system"
	CapiKubeadmBootstrapSystemNamespace     = "capi-kubeadm-bootstrap-system"
	CapiKubeadmControlPlaneSystemNamespace  = "capi-kubeadm-control-plane-system"
	CapiSystemNamespace                     = "capi-system"
	CapiWebhookSystemNamespace              = "capi-webhook-system"
	CapvSystemNamespace                     = "capv-system"
	CaptSystemNamespace                     = "capt-system"
	CapaSystemNamespace                     = "capa-system"
	CapasSystemNamespace                    = "capas-system"
	CapxSystemNamespace                     = "capx-system"
	CertManagerNamespace                    = "cert-manager"
	DefaultNamespace                        = "default"
	EtcdAdmBootstrapProviderSystemNamespace = "etcdadm-bootstrap-provider-system"
	EtcdAdmControllerSystemNamespace        = "etcdadm-controller-system"
	KubeNodeLeaseNamespace                  = "kube-node-lease"
	KubePublicNamespace                     = "kube-public"
	KubeSystemNamespace                     = "kube-system"
	AdotPrometheusNamespace                 = "observability"
	MetallbNamespace                        = "metallb-system"
	EmissaryNamespace                       = "emissary-system"
	HarborNamespace                         = "harbor"
	LocalPathStorageNamespace               = "local-path-storage"
	EtcdAdmBootstrapProviderName            = "bootstrap-etcdadm-bootstrap"
	EtcdadmControllerProviderName           = "bootstrap-etcdadm-controller"
	DefaultHttpsPort                        = "443"
	DefaultWorkerNodeGroupName              = "md-0"
	DefaultNodeCidrMaskSize                 = 24

	VSphereProviderName    = "vsphere"
	DockerProviderName     = "docker"
	AWSProviderName        = "aws"
	SnowProviderName       = "snow"
	TinkerbellProviderName = "tinkerbell"
	CloudStackProviderName = "cloudstack"
	NutanixProviderName    = "nutanix"
	// DefaultNutanixPrismCentralPort is the default port for Nutanix Prism Central.
	DefaultNutanixPrismCentralPort = 9440

	VSphereCredentialsName = "vsphere-credentials"
	NutanixCredentialsName = "nutanix-credentials"
	EksaLicenseName        = "eksa-license"
	EksaPackagesName       = "eksa-packages"
	// UpgraderConfigMapName is the name of config map that stores the upgrader images.
	UpgraderConfigMapName = "in-place-upgrade"
	// KubeVipConfigMapName is the name of config map that stores the kube-vip config.
	KubeVipConfigMapName = "kube-vip-in-place-upgrade"
	// KubeVipManifestName is the name of kube-vip spec file.
	KubeVipManifestName = "kube-vip.yaml"

	CloudstackAnnotationSuffix = "cloudstack.anywhere.eks.amazonaws.com/v1alpha1"

	FailureDomainLabelName = "cluster.x-k8s.io/failure-domain"

	// ClusterctlMoveLabelName adds the clusterctl move label.
	ClusterctlMoveLabelName = "clusterctl.cluster.x-k8s.io/move"

	// CloudstackFailureDomainPlaceholder Provider specific keywork placeholder.
	CloudstackFailureDomainPlaceholder = "ds.meta_data.failuredomain"

	// DefaultCoreEKSARegistry is the default registry for eks-a core artifacts.
	DefaultCoreEKSARegistry = "public.ecr.aws"

	// DefaultCuratedPackagesRegistry matches the default registry for curated packages in all regions.
	DefaultCuratedPackagesRegistryRegex = `783794618700\.dkr\.ecr\..*\.amazonaws\.com`

	// DefaultcuratedPackagesRegistry is a containerd compatible registry format that matches all AWS regions.
	DefaultCuratedPackagesRegistry = "783794618700.dkr.ecr.*.amazonaws.com"

	// Provider specific env vars.
	VSphereUsernameKey     = "VSPHERE_USERNAME"
	VSpherePasswordKey     = "VSPHERE_PASSWORD"
	GovcUsernameKey        = "GOVC_USERNAME"
	GovcPasswordKey        = "GOVC_PASSWORD"
	SnowCredentialsKey     = "AWS_B64ENCODED_CREDENTIALS"
	SnowCertsKey           = "AWS_B64ENCODED_CA_BUNDLES"
	NutanixUsernameKey     = "NUTANIX_USER"
	NutanixPasswordKey     = "NUTANIX_PASSWORD"
	EksaNutanixUsernameKey = "EKSA_NUTANIX_USERNAME"
	EksaNutanixPasswordKey = "EKSA_NUTANIX_PASSWORD"
	RegistryUsername       = "REGISTRY_USERNAME"
	RegistryPassword       = "REGISTRY_PASSWORD"

	SecretKind             = "Secret"
	ConfigMapKind          = "ConfigMap"
	ClusterResourceSetKind = "ClusterResourceSet"

	NutanixMachineConfigKind = "NutanixMachineConfig"

	BottlerocketDefaultUser = "ec2-user"
	UbuntuDefaultUser       = "capv"

	// DefaultUnhealthyMachineTimeout is the default timeout for an unhealthy machine health check.
	DefaultUnhealthyMachineTimeout = 5 * time.Minute
	// DefaultNodeStartupTimeout is the default timeout for a machine without a node to be considered to have failed machine health check.
	DefaultNodeStartupTimeout = 10 * time.Minute
	// DefaultTinkerbellNodeStartupTimeout is the default node start up timeout for Tinkerbell.
	DefaultTinkerbellNodeStartupTimeout = 20 * time.Minute
	// DefaultMaxUnhealthy is the default maxUnhealthy value for machine health checks.
	DefaultMaxUnhealthy = "100%"
	// DefaultWorkerMaxUnhealthy is the default maxUnhealthy value for worker node machine health checks.
	DefaultWorkerMaxUnhealthy = "40%"
)

type Operation int

const (
	Create  Operation = 0
	Upgrade Operation = 1
	Delete  Operation = 2
)

// EKSACLIFieldManager is the owner name for fields applied by the EKS-A CLI.
const EKSACLIFieldManager = "eks-a-cli"

// SupportedProviders is the list of supported providers for generating EKS-A cluster spec.
var SupportedProviders = []string{
	VSphereProviderName,
	CloudStackProviderName,
	TinkerbellProviderName,
	DockerProviderName,
	NutanixProviderName,
	SnowProviderName,
}
