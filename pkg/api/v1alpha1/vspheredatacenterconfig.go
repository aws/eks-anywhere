package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const VSphereDatacenterKind = "VSphereDatacenterConfig"

// Used for generating yaml for generate clusterconfig command
func NewVSphereDatacenterConfigGenerate(clusterName string) *VSphereDatacenterConfigGenerate {
	return &VSphereDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       VSphereDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: VSphereDatacenterConfigSpec{},
	}
}

func (c *VSphereDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *VSphereDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *VSphereDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetVSphereDatacenterConfig(fileName string) (*VSphereDatacenterConfig, error) {
	var clusterConfig VSphereDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
