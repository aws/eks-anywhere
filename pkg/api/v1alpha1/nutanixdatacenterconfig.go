package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
)

const NutanixDatacenterKind = "NutanixDatacenterConfig"

// NewNutanixDatacenterConfigGenerate is used for generating yaml for generate clusterconfig command.
func NewNutanixDatacenterConfigGenerate(clusterName string) *NutanixDatacenterConfigGenerate {
	return &NutanixDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       NutanixDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: NutanixDatacenterConfigSpec{
			Endpoint: "<enter Prism Central Endpoint (FQDN or IP) here>",
			Port:     9440,
			CredentialRef: &Ref{
				Kind: constants.SecretKind,
				Name: constants.NutanixCredentialsName,
			},
		},
	}
}

func (c *NutanixDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *NutanixDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *NutanixDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

// GetNutanixDatacenterConfig parses config in a yaml file and returns a NutanixDatacenterConfig object.
func GetNutanixDatacenterConfig(fileName string) (*NutanixDatacenterConfig, error) {
	var clusterConfig NutanixDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
