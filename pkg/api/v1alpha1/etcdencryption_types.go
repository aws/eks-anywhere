package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EtcdEncryption defines the configuration for ETCD encryption.
type EtcdEncryption struct {
	Providers []EtcdEncryptionProvider `json:"providers"`
	// Resources defines a list of objects and custom resources definitions that should be encrypted.
	Resources []string `json:"resources"`
}

// EtcdEncryptionProvider defines the configuration for ETCD encryption providers.
// Currently only KMS provider is supported.
type EtcdEncryptionProvider struct {
	// KMS defines the configuration for KMS Encryption provider.
	KMS *KMS `json:"kms"`
}

// KMS defines the configuration for KMS Encryption provider.
type KMS struct {
	// CacheSize defines the maximum number of encrypted objects to be cached in memory. The default value is 1000.
	// You can set this to a negative value to disable caching.
	CacheSize *int32 `json:"cachesize,omitempty"`
	// Name defines the name of KMS plugin to be used.
	Name string `json:"name"`
	// SocketListenAddress defines a UNIX socket address that the KMS provider listens on.
	SocketListenAddress string `json:"socketListenAddress"`
	// Timeout for kube-apiserver to wait for KMS plugin. Default is 3s.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}
