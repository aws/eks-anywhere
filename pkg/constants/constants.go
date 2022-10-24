package constants

// Namespace constants
const (
	EksaSystemNamespace                     = "eksa-system"
	EksaDiagnosticsNamespace                = "eksa-diagnostics"
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
	LocalPathStorageNamespace               = "local-path-storage"
	EtcdAdmBootstrapProviderName            = "bootstrap-etcdadm-bootstrap"
	EtcdadmControllerProviderName           = "bootstrap-etcdadm-controller"
	DefaultHttpsPort                        = "443"
	DefaultWorkerNodeGroupName              = "md-0"

	VSphereProviderName    = "vsphere"
	DockerProviderName     = "docker"
	AWSProviderName        = "aws"
	SnowProviderName       = "snow"
	TinkerbellProviderName = "tinkerbell"
	CloudStackProviderName = "cloudstack"
	NutanixProviderName    = "nutanix"

	VSphereCredentialsName = "vsphere-credentials"
	EksaLicenseName        = "eksa-license"
	EksaPackagesName       = "eksa-packages"

	DefaultRegistry            = "public.ecr.aws"
	CloudstackAnnotationSuffix = "cloudstack.anywhere.eks.amazonaws.com/v1alpha1"

	FailuredomainLabelName = "cluster.x-k8s.io/failure-domain"

	// provider specific keywork placeholder
	CloudstackFailuredomainPlaceholder = "ds.meta_data.failuredomain"

	// provider specific env vars
	VSphereUsernameKey = "VSPHERE_USERNAME"
	VSpherePasswordKey = "VSPHERE_PASSWORD"
	GovcUsernameKey    = "GOVC_USERNAME"
	GovcPasswordKey    = "GOVC_PASSWORD"
	SnowCredentialsKey = "AWS_B64ENCODED_CREDENTIALS"
	SnowCertsKey       = "AWS_B64ENCODED_CA_BUNDLES"
	NutanixUsernameKey = "NUTANIX_USER"
	NutanixPasswordKey = "NUTANIX_PASSWORD"
)

type Operation int

const (
	Create  Operation = 0
	Upgrade Operation = 1
	Delete  Operation = 2
)
