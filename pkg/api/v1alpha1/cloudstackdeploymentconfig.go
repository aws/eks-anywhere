package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CloudStackDeploymentKind = "CloudStackDeploymentConfig"

// Used for generating yaml for generate clusterconfig command
func NewCloudStackDeploymentConfigGenerate(clusterName string) *CloudStackDeploymentConfigGenerate {
	return &CloudStackDeploymentConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudStackDeploymentKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: CloudStackDeploymentConfigSpec{
			Domain:                "domain1",
			Zone:                  "zone1",
			Network:               "net1",
			Account:               "admin",
			ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
			Insecure:              false,
		},
	}
}

func (c *CloudStackDeploymentConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudStackDeploymentConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudStackDeploymentConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudStackDeploymentConfig(fileName string) (*CloudStackDeploymentConfig, error) {
	var clusterConfig CloudStackDeploymentConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
