package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CloudStackDatacenterKind = "CloudStackDatacenterConfig"

// Used for generating yaml for generate clusterconfig command
func NewCloudStackDatacenterConfigGenerate(clusterName string) *CloudStackDatacenterConfigGenerate {
	return &CloudStackDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudStackDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: CloudStackDatacenterConfigSpec{
			Domain: "domain1",
			Zones: []CloudStackZone{
				{
					Network: CloudStackResourceIdentifier{},
				},
			},
			Account: "admin",
		},
	}
}

func (c *CloudStackDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudStackDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudStackDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudStackDatacenterConfig(fileName string) (*CloudStackDatacenterConfig, error) {
	var clusterConfig CloudStackDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
