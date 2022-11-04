package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TinkerbellDatacenterKind = "TinkerbellDatacenterConfig"

// NewTinkerbellDatacenterConfigGenerate Used for generating yaml for generate clusterconfig command.
func NewTinkerbellDatacenterConfigGenerate(clusterName string) *TinkerbellDatacenterConfigGenerate {
	return &TinkerbellDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: TinkerbellDatacenterConfigSpec{},
	}
}

func (c *TinkerbellDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetTinkerbellDatacenterConfig(fileName string) (*TinkerbellDatacenterConfig, error) {
	var clusterConfig TinkerbellDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
