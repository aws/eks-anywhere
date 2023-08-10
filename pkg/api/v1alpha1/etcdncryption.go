package v1alpha1

import "github.com/pkg/errors"

func validateEtcdEncryptionConfig(config *[]EtcdEncryption) error {
	if config == nil {
		return nil
	}

	if len(*config) != 1 {
		return errors.New("invalid number of encryption providers, currently only 1 encryption provider is supported")
	}

	return nil
}
