package v1alpha1

import (
	"net/url"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

var (
	// DefaultKMSCacheSize is the default cache size for KMS provider (1000).
	DefaultKMSCacheSize = ptr.Int32(1000)
	// DefaultKMSTimeout is the default timeout for KMS provider (3s).
	DefaultKMSTimeout = metav1.Duration{Duration: time.Second * 3}
)

// ValidateEtcdEncryptionConfig validates the etcd encryption configuration.
func ValidateEtcdEncryptionConfig(config *[]EtcdEncryption) error {
	if config == nil {
		return nil
	}

	if len(*config) == 0 {
		return errors.New("etcdEncryption cannot be empty")
	}

	if len(*config) != 1 {
		return errors.New("etcdEncryption config is invalid, only 1 encryption config is supported currently")
	}

	for i, c := range *config {
		if len(c.Providers) == 0 {
			return errors.Errorf("etcdEncryption[%d].providers cannot be empty", i)
		}
		if len(c.Providers) != 1 {
			return errors.Errorf("etcdEncryption[%d].providers in invalid, only 1 encryption provider is currently supported", i)
		}
		for j, p := range c.Providers {
			if err := validateKMSConfig(p.KMS); err != nil {
				return errors.Errorf("etcdEncryption[%d].providers[%d] is invalid: %v", i, j, err)
			}
		}
		if len(c.Resources) == 0 {
			return errors.Errorf("etcdEncryption[%d].resources cannot be empty", i)
		}
	}

	return nil
}

func validateKMSConfig(kms *KMS) error {
	if kms == nil {
		return errors.New("kms cannot be nil")
	}
	if len(kms.Name) == 0 {
		return errors.New("kms.name cannot be empty")
	}
	if len(kms.SocketListenAddress) == 0 {
		return errors.New("kms.socketListenAddress cannot be empty")
	}
	u, err := url.Parse(kms.SocketListenAddress)
	if err != nil {
		return errors.Errorf("kms.socketListenAddress is malformed: %v", err)
	}
	if u.Scheme != "unix" {
		return errors.Errorf("kms.socketListenAddress has unsupported scheme: %v", u.Scheme)
	}
	return nil
}

func setEtcdEncryptionConfigDefaults(cluster *Cluster) error {
	if cluster.Spec.EtcdEncryption == nil {
		return nil
	}

	for _, c := range *cluster.Spec.EtcdEncryption {
		for _, p := range c.Providers {
			setKMSConfigDefauts(p.KMS)
		}
	}
	return nil
}

func setKMSConfigDefauts(kms *KMS) {
	if kms != nil {
		if kms.CacheSize == nil {
			kms.CacheSize = DefaultKMSCacheSize
		}
		if kms.Timeout == nil {
			kms.Timeout = &DefaultKMSTimeout
		}
	}
}
