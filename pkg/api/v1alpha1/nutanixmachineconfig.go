package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
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

	defaultNutanixOSFamily          = Ubuntu
	defaultNutanixSystemDiskSizeGi  = "40Gi"
	defaultNutanixMemorySizeGi      = "4Gi"
	defaultNutanixVCPUsPerSocket    = 1
	defaultNutanixVCPUSockets       = 2
	defaultNutanixMachineConfigUser = "nutanix-user"
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
					Name:              defaultNutanixMachineConfigUser,
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

func GetNutanixMachineConfigs(fileName string) (map[string]*NutanixMachineConfig, error) {
	configs := make(map[string]*NutanixMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		config := NutanixMachineConfig{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == NutanixMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == NutanixMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", NutanixMachineConfigKind)
	}
	return configs, nil
}

func setNutanixMachineConfigDefaults(machineConfig *NutanixMachineConfig) {
	if len(machineConfig.Spec.Users) <= 0 {
		machineConfig.Spec.Users = []UserConfiguration{{}}
	}

	if len(machineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	if machineConfig.Spec.OSFamily == "" {
		machineConfig.Spec.OSFamily = defaultNutanixOSFamily
	}

	if len(machineConfig.Spec.Users) == 0 || machineConfig.Spec.Users[0].Name == "" {
		machineConfig.Spec.Users[0].Name = defaultNutanixMachineConfigUser
		logger.V(1).Info("SSHUsername is not set or is empty for NutanixMachineConfig, using default", "machineConfig", machineConfig.Name, "user", machineConfig.Spec.Users[0].Name)
	}
}
