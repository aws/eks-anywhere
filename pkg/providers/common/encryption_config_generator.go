package common

import (
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverv1 "k8s.io/apiserver/pkg/apis/apiserver/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	encryptionConfigurationKind  = "EncryptionConfiguration"
	encryptionProviderVersion    = "v1"
	encryptionProviderNamePrefix = "aws-encryption-provider"
)

var identityProvider = apiserverv1.ProviderConfiguration{
	Identity: &apiserverv1.IdentityConfiguration{},
}

// GenerateKMSEncryptionConfiguration takes a list of the EtcdEncryption configs and generates the corresponding Kubernetes Encryptionapiserverv1.
func GenerateKMSEncryptionConfiguration(confs *[]v1alpha1.EtcdEncryption) (string, error) {
	if confs == nil || len(*confs) == 0 {
		return "", nil
	}
	encryptionConf := &apiserverv1.EncryptionConfiguration{
		TypeMeta: v1.TypeMeta{
			APIVersion: apiserverv1.SchemeGroupVersion.Identifier(),
			Kind:       encryptionConfigurationKind,
		},
	}

	resourceConfigs := make([]apiserverv1.ResourceConfiguration, 0, len(*confs))
	for _, conf := range *confs {
		providers := []apiserverv1.ProviderConfiguration{}
		for _, provider := range conf.Providers {
			provider := apiserverv1.ProviderConfiguration{
				KMS: &apiserverv1.KMSConfiguration{
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
		resourceConfig := apiserverv1.ResourceConfiguration{
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
