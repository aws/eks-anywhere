package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const SnowDatacenterKind = "SnowDatacenterConfig"

// Used for generating yaml for generate clusterconfig command
func NewSnowDatacenterConfigGenerate(clusterName string) *SnowDatacenterConfigGenerate {
	return &SnowDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       SnowDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: SnowDatacenterConfigSpec{},
	}
}

func (s *SnowDatacenterConfigGenerate) APIVersion() string {
	return s.TypeMeta.APIVersion
}

func (s *SnowDatacenterConfigGenerate) Kind() string {
	return s.TypeMeta.Kind
}

func (s *SnowDatacenterConfigGenerate) Name() string {
	return s.ObjectMeta.Name
}

func GetSnowDatacenterConfig(fileName string) (*SnowDatacenterConfig, error) {
	var clusterConfig SnowDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
