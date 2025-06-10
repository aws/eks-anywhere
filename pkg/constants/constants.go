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

	// Certificate renewal component types.
	EtcdComponent         = "etcd"
	ControlPlaneComponent = "control-plane"

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

	// SignatureAnnotation applied to the bundle during bundle manifest signing.
	SignatureAnnotation = "anywhere.eks.amazonaws.com/signature"
	// ExcludesAnnotation applied to the bundle during bundle manifest signing.
	ExcludesAnnotation = "anywhere.eks.amazonaws.com/excludes"
	// SignatureAnnotation applied to the bundle during eks distro manifest signing.
	EKSDistroSignatureAnnotation = "distro.eks.amazonaws.com/signature"
	// ExcludesAnnotation applied to the bundle during eks distro manifest signing.
	EKSDistroExcludesAnnotation = "distro.eks.amazonaws.com/excludes"
	// Excludes is a base64-encoded, newline-delimited list of JSON/YAML paths to remove
	// from the Bundles manifest prior to computing the digest. You can add or remove
	// fields depending on your signing requirements.
	// We are excluding some fields from the versionbundle object from signing/verifying the signature to allow users to override images.
	// To check the fields we are excluding for signing/verifying the signature base64 decode the Excludes field.
	Excludes = "LnNwZWMudmVyc2lvbnNCdW5kbGVzW10uYm9vdHN0cmFwCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmJvdHRsZXJvY2tldEhvc3RDb250YWluZXJzCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNlcnRNYW5hZ2VyCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNpbGl1bQouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5jbG91ZFN0YWNrCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNsdXN0ZXJBUEkKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uY29udHJvbFBsYW5lCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmRvY2tlcgouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5la3NhCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmVrc0QuY29tcG9uZW50cwouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5la3NELm1hbmlmZXN0VXJsCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmV0Y2RhZG1Cb290c3RyYXAKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uZXRjZGFkbUNvbnRyb2xsZXIKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uZmx1eAouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5oYXByb3h5Ci5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmtpbmRuZXRkCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLm51dGFuaXgKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10ucGFja2FnZUNvbnRyb2xsZXIKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uc25vdwouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS50aW5rZXJiZWxsCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLnVwZ3JhZGVyCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLnZTcGhlcmU="
	// EKSDistroExcludes is a base64-encoded, newline-delimited list of JSON/YAML paths to remove
	// from the EKS Distro manifest prior to computing the digest. You can add or remove
	// fields depending on your signing requirements.
	EKSDistroExcludes = "Cg=="
	// KMSPublicKey to verify bundle manifest signature.
	KMSPublicKey = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFZU/Z6VVMU9HioT7rGkPdJg3frC2xyQZhWFIrz5HeZEfeQ2nAdnJMLrs2Qr3V9xVrJrHA54wnIHDoPGbEhojqg=="
	// KMSPublicKey to verify eks distro manifest signature.
	EKSDistroKMSPublicKey = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfqjUloL1WPdxB5JPOY18mIgjx0zD0SyWe9N1Fxjv7A8JwyKQycqirw2us7zoxoC5bnWD4lT53uv8skx3E/cHaw=="
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

// AlwaysExcludedFields contains a list of string that need to be excluded while getting a digest of bundle to check the signature validation.
var AlwaysExcludedFields = []string{
	".status",
	".metadata.creationTimestamp",
	".metadata.annotations",
	".metadata.generation",
	".metadata.managedFields",
	".metadata.uid",
	".metadata.resourceVersion",
	".metadata.namespace",
}
