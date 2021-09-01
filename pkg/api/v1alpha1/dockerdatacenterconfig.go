package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DockerDatacenterKind = "DockerDatacenterConfig"

// Used for generating yaml for generate clusterconfig command
func NewDockerDatacenterConfigGenerate(clusterName string) *DockerDatacenterConfigGenerate {
	return &DockerDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       DockerDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name:      clusterName,
			Namespace: "anywhere-system",
		},
		Spec: DockerDatacenterConfigSpec{},
	}
}

func (c *DockerDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *DockerDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *DockerDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func (c *DockerDatacenterConfigGenerate) Namespace() string {
	return c.ObjectMeta.Namespace
}

func GetDockerDatacenterConfig(fileName string) (*DockerDatacenterConfig, error) {
	var clusterConfig DockerDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}
