package common

import (
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	config "k8s.io/apiserver/pkg/apis/config/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	encryptionConfigurationKind  = "EncryptionConfiguration"
	encryptionProviderVersion    = "v1"
	encryptionProviderNamePrefix = "aws-encryption-provider"
)

var identityProvider = config.ProviderConfiguration{
	Identity: &config.IdentityConfiguration{},
}

// GenerateKMSEncryptionConfiguration takes a list of the EtcdEncryption configs and generates the corresponding Kubernetes EncryptionConfig.
func GenerateKMSEncryptionConfiguration(confs *[]v1alpha1.EtcdEncryption) (string, error) {
	if confs == nil || len(*confs) == 0 {
		return "", nil
	}

	encryptionConf := &config.EncryptionConfiguration{
		TypeMeta: v1.TypeMeta{
			APIVersion: config.SchemeGroupVersion.Identifier(),
			Kind:       encryptionConfigurationKind,
		},
	}

	resourceConfigs := make([]config.ResourceConfiguration, 0, len(*confs))
	for _, conf := range *confs {
		providers := []config.ProviderConfiguration{}
		for _, provider := range conf.Providers {
			provider := config.ProviderConfiguration{
				KMS: &config.KMSConfiguration{
					APIVersion: encryptionProviderVersion,
					Name:       provider.KMS.Name,
					Endpoint:   provider.KMS.SocketListenAddress,
					CacheSize:  provider.KMS.CacheSize,
					Timeout: &v1.Duration{
						Duration: provider.KMS.Timeout.Duration,
					},
				},
			}
			providers = append(providers, provider)
		}
		providers = append(providers, identityProvider)
		resourceConfig := config.ResourceConfiguration{
			Resources: conf.Resources,
			Providers: providers,
		}
		resourceConfigs = append(resourceConfigs, resourceConfig)
	}
	encryptionConf.Resources = resourceConfigs

	marshaledConf, err := yaml.Marshal(encryptionConf)
	if err != nil {
		return "", fmt.Errorf("marshaling encryption config: %v", err)
	}
	return strings.Trim(string(marshaledConf), "\n"), nil
}
