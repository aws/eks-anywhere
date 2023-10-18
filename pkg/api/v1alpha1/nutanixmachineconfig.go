package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NutanixIdentifierType is an enumeration of different resource identifier types.
type NutanixIdentifierType string

func (c NutanixIdentifierType) String() string {
	return string(c)
}

const (
	// NutanixMachineConfigKind is the kind for a NutanixMachineConfig.
	NutanixMachineConfigKind = "NutanixMachineConfig"

	// NutanixIdentifierUUID is a resource identifier identifying the object by UUID.
	NutanixIdentifierUUID NutanixIdentifierType = "uuid"
	// NutanixIdentifierName is a resource identifier identifying the object by Name.
	NutanixIdentifierName NutanixIdentifierType = "name"

	defaultNutanixOSFamily         = Ubuntu
	defaultNutanixSystemDiskSizeGi = "40Gi"
	defaultNutanixMemorySizeGi     = "4Gi"
	defaultNutanixVCPUsPerSocket   = 1
	defaultNutanixVCPUSockets      = 2

	// DefaultNutanixMachineConfigUser is the default username we set in machine config.
	DefaultNutanixMachineConfigUser string = "eksa"
)

// NutanixResourceIdentifier holds the identity of a Nutanix Prism resource (cluster, image, subnet, etc.)
//
// +union.
type NutanixResourceIdentifier struct {
	// Type is the identifier type to use for this resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=uuid;name
	Type NutanixIdentifierType `json:"type"`

	// uuid is the UUID of the resource in the PC.
	// +optional
	UUID *string `json:"uuid,omitempty"`

	// name is the resource name in the PC
	// +optional
	Name *string `json:"name,omitempty"`
}

// NutanixCategoryIdentifier holds the identity of a Nutanix Prism Central category.
type NutanixCategoryIdentifier struct {
	// key is the Key of the category in the Prism Central.
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`

	// value is the category value linked to the key in the Prism Central.
	// +kubebuilder:validation:Required
	Value string `json:"value,omitempty"`
}

// NutanixMachineConfigGenerateOpt is a functional option that can be passed to NewNutanixMachineConfigGenerate to
// customize the generated machine config
//
// +kubebuilder:object:generate=false
type NutanixMachineConfigGenerateOpt func(config *NutanixMachineConfigGenerate)

// NewNutanixMachineConfigGenerate returns a new instance of NutanixMachineConfigGenerate
// used for generating yaml for generate clusterconfig command.
func NewNutanixMachineConfigGenerate(name string, opts ...NutanixMachineConfigGenerateOpt) *NutanixMachineConfigGenerate {
	enterNameString := "<Enter %s name here>"
	machineConfig := &NutanixMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       NutanixMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: NutanixMachineConfigSpec{
			OSFamily: defaultNutanixOSFamily,
			Users: []UserConfiguration{
				{
					Name:              DefaultNutanixMachineConfigUser,
					SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
				},
			},
			VCPUsPerSocket: defaultNutanixVCPUsPerSocket,
			VCPUSockets:    defaultNutanixVCPUSockets,
			MemorySize:     resource.MustParse(defaultNutanixMemorySizeGi),
			Image:          NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: func() *string { s := fmt.Sprintf(enterNameString, "image"); return &s }()},
			Cluster:        NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: func() *string { s := fmt.Sprintf(enterNameString, "Prism Element cluster"); return &s }()},
			Subnet:         NutanixResourceIdentifier{Type: NutanixIdentifierName, Name: func() *string { s := fmt.Sprintf(enterNameString, "subnet"); return &s }()},
			SystemDiskSize: resource.MustParse(defaultNutanixSystemDiskSizeGi),
		},
	}

	for _, opt := range opts {
		opt(machineConfig)
	}

	return machineConfig
}

func (c *NutanixMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *NutanixMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *NutanixMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func setNutanixMachineConfigDefaults(machineConfig *NutanixMachineConfig) {
	initUser := UserConfiguration{
		Name:              DefaultNutanixMachineConfigUser,
		SshAuthorizedKeys: []string{""},
	}
	if machineConfig.Spec.Users == nil || len(machineConfig.Spec.Users) <= 0 {
		machineConfig.Spec.Users = []UserConfiguration{initUser}
	}

	user := machineConfig.Spec.Users[0]
	if user.Name == "" {
		machineConfig.Spec.Users[0].Name = DefaultNutanixMachineConfigUser
	}

	if user.SshAuthorizedKeys == nil || len(user.SshAuthorizedKeys) <= 0 {
		machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	if machineConfig.Spec.OSFamily == "" {
		machineConfig.Spec.OSFamily = defaultNutanixOSFamily
	}
}

func validateNutanixMachineConfig(c *NutanixMachineConfig) error {
	if err := validateObjectMeta(c.ObjectMeta); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if err := validateNutanixReferences(c); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if err := validateMinimumNutanixMachineSpecs(c); err != nil {
		return fmt.Errorf("NutanixMachineConfig: %v", err)
	}

	if c.Spec.OSFamily != Ubuntu && c.Spec.OSFamily != RedHat {
		return fmt.Errorf(
			"NutanixMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s",
			c.Spec.OSFamily,
			Ubuntu,
			RedHat,
		)
	}

	if err := validateMachineConfigUsers(c.Name, NutanixMachineConfigKind, c.Spec.Users); err != nil {
		return err
	}

	return nil
}

func validateMinimumNutanixMachineSpecs(c *NutanixMachineConfig) error {
	if c.Spec.VCPUSockets < defaultNutanixVCPUSockets {
		return fmt.Errorf("NutanixMachineConfig: vcpu sockets must be greater than or equal to %d", defaultNutanixVCPUSockets)
	}

	if c.Spec.VCPUsPerSocket < defaultNutanixVCPUsPerSocket {
		return fmt.Errorf("NutanixMachineConfig: vcpu per socket must be greater than or equal to %d", defaultNutanixVCPUsPerSocket)
	}

	if c.Spec.MemorySize.Cmp(resource.MustParse(defaultNutanixMemorySizeGi)) < 0 {
		return fmt.Errorf("NutanixMachineConfig: memory size must be greater than or equal to %s", defaultNutanixMemorySizeGi)
	}

	if c.Spec.SystemDiskSize.Cmp(resource.MustParse(defaultNutanixSystemDiskSizeGi)) < 0 {
		return fmt.Errorf("NutanixMachineConfig: system disk size must be greater than %s", defaultNutanixSystemDiskSizeGi)
	}

	return nil
}

func validateNutanixReferences(c *NutanixMachineConfig) error {
	if err := validateNutanixResourceReference(&c.Spec.Subnet, "subnet", c.Name); err != nil {
		return err
	}

	if err := validateNutanixResourceReference(&c.Spec.Cluster, "cluster", c.Name); err != nil {
		return err
	}

	if err := validateNutanixResourceReference(&c.Spec.Image, "image", c.Name); err != nil {
		return err
	}

	if c.Spec.Project != nil {
		if err := validateNutanixResourceReference(c.Spec.Project, "project", c.Name); err != nil {
			return err
		}
	}

	if len(c.Spec.AdditionalCategories) > 0 {
		if err := validateNutanixCategorySlice(c.Spec.AdditionalCategories, c.Name); err != nil {
			return err
		}
	}

	return nil
}

func validateNutanixResourceReference(i *NutanixResourceIdentifier, resource string, mcName string) error {
	if i.Type != NutanixIdentifierName && i.Type != NutanixIdentifierUUID {
		return fmt.Errorf("NutanixMachineConfig: invalid identifier type for %s: %s", resource, i.Type)
	}

	if i.Type == NutanixIdentifierName && i.Name == nil {
		return fmt.Errorf("NutanixMachineConfig: missing %s name: %s", resource, mcName)
	} else if i.Type == NutanixIdentifierUUID && i.UUID == nil {
		return fmt.Errorf("NutanixMachineConfig: missing %s UUID: %s", resource, mcName)
	}

	return nil
}

func validateNutanixCategorySlice(i []NutanixCategoryIdentifier, mcName string) error {
	for _, category := range i {
		if category.Key == "" {
			return fmt.Errorf("NutanixMachineConfig: missing category key: %s", mcName)
		}

		if category.Value == "" {
			return fmt.Errorf("NutanixMachineConfig: missing category value for key %s: %s", category.Key, mcName)
		}
	}

	return nil
}
