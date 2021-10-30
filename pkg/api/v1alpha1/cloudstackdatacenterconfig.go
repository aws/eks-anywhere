package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CloudstackDatacenterKind = "CloudstackDatacenterConfig"

// Used for generating yaml for generate clusterconfig command
func NewCloudstackDatacenterConfigGenerate(clusterName string) *CloudstackDatacenterConfigGenerate {
	return &CloudstackDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudstackDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: CloudstackDatacenterConfigSpec{
			Domain: "domain1",
			Zone: "zone1",
			Network: "net1",
			Account: "admin",
			ControlPlaneEndpoint: "https://127.0.0.1:8080/client/api",
			Insecure: false,

		},
	}
}

func (c *CloudstackDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudstackDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudstackDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudstackDatacenterConfig(fileName string) (*CloudstackDatacenterConfig, error) {
	var clusterConfig CloudstackDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
