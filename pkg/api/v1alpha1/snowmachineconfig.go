package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	SnowMachineConfigKind   = "SnowMachineConfig"
	DefaultSnowInstanceType = "sbe-c.large"
	DefaultSnowSshKeyName   = "default"
)

// Used for generating yaml for generate clusterconfig command
func NewSnowMachineConfigGenerate(name string) *SnowMachineConfigGenerate {
	return &SnowMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       SnowMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: SnowMachineConfigSpec{
			AMIID:        "",
			InstanceType: DefaultSnowInstanceType,
			SshKeyName:   DefaultSnowSshKeyName,
		},
	}
}

func (s *SnowMachineConfigGenerate) APIVersion() string {
	return s.TypeMeta.APIVersion
}

func (s *SnowMachineConfigGenerate) Kind() string {
	return s.TypeMeta.Kind
}

func (s *SnowMachineConfigGenerate) Name() string {
	return s.ObjectMeta.Name
}
