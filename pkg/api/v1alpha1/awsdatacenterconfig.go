package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AWSDatacenterKind = "AWSDatacenterConfig"

// Used for generating yaml for generate clusterconfig command.
func NewAWSDatacenterConfigGenerate(clusterName string) *AWSDatacenterConfigGenerate {
	return &AWSDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       AWSDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: AWSDatacenterConfigSpec{},
	}
}

func (c *AWSDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *AWSDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *AWSDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetAWSDatacenterConfig(fileName string) (*AWSDatacenterConfig, error) {
	var clusterConfig AWSDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
