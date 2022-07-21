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

	VSphereCredentialsName = "vsphere-credentials"
	EksaLicenseName        = "eksa-license"
	EksaPackagesName       = "eksa-packages"

	DefaultRegistry            = "public.ecr.aws"
	CloudstackAnnotationSuffix = "cloudstack.anywhere.eks.amazonaws.com/v1alpha1"
)
