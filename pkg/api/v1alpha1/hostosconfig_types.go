package v1alpha1

import "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

// HostOSConfiguration defines the configuration settings on the host OS.
type HostOSConfiguration struct {
	// +optional
	NTPConfiguration *NTPConfiguration `json:"ntpConfiguration,omitempty"`

	// +optional
	BottlerocketConfiguration *BottlerocketConfiguration `json:"bottlerocketConfiguration,omitempty"`

	// +optional
	CertBundles []certBundle `json:"certBundles,omitempty"`
}

// NTPConfiguration defines the NTP configuration on the host OS.
type NTPConfiguration struct {
	// Servers defines a list of NTP servers to be configured on the host OS.
	Servers []string `json:"servers"`
}

// BottlerocketConfiguration defines the Bottlerocket configuration on the host OS.
// These settings only take effect when the `osFamily` is bottlerocket.
type BottlerocketConfiguration struct {
	// Kubernetes defines the Kubernetes settings on the host OS.
	// +optional
	Kubernetes *v1beta1.BottlerocketKubernetesSettings `json:"kubernetes,omitempty"`

	// Kernel defines the kernel settings for bottlerocket.
	Kernel *v1beta1.BottlerocketKernelSettings `json:"kernel,omitempty"`

	// Boot defines the boot settings for bottlerocket.
	Boot *v1beta1.BottlerocketBootSettings `json:"boot,omitempty"`
}

// Cert defines additional trusted cert bundles on the host OS.
type certBundle struct {
	// Name defines the cert bundle name.
	Name string `json:"name"`

	// Data defines the cert bundle data.
	Data string `json:"data"`
}
