package v1alpha1

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
	// SocketListenAddress defines a UNIX socket address that the KMS provider listens on.
	SocketListenAddress string `json:"socketListenAddress"`
}
