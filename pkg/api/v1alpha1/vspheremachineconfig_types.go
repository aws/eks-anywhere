package v1alpha1

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VSphereMachineConfigSpec defines the desired state of VSphereMachineConfig.
type VSphereMachineConfigSpec struct {
	DiskGiB             int                  `json:"diskGiB,omitempty"`
	Datastore           string               `json:"datastore"`
	Folder              string               `json:"folder"`
	NumCPUs             int                  `json:"numCPUs"`
	MemoryMiB           int                  `json:"memoryMiB"`
	OSFamily            OSFamily             `json:"osFamily"`
	ResourcePool        string               `json:"resourcePool"`
	StoragePolicyName   string               `json:"storagePolicyName,omitempty"`
	Template            string               `json:"template,omitempty"`
	Users               []UserConfiguration  `json:"users,omitempty"`
	TagIDs              []string             `json:"tags,omitempty"`
	CloneMode           CloneMode            `json:"cloneMode,omitempty"`
	HostOSConfiguration *HostOSConfiguration `json:"hostOSConfiguration,omitempty"`
}

func (c *VSphereMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *VSphereMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *VSphereMachineConfig) SetManagedBy(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

// IsManaged returns true if the vspheremachineconfig is associated with a workload cluster.
func (c *VSphereMachineConfig) IsManaged() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *VSphereMachineConfig) OSFamily() OSFamily {
	return c.Spec.OSFamily
}

func (c *VSphereMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *VSphereMachineConfig) GetName() string {
	return c.Name
}

// VSphereMachineConfigStatus defines the observed state of VSphereMachineConfig.
type VSphereMachineConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VSphereMachineConfig is the Schema for the vspheremachineconfigs API.
type VSphereMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereMachineConfigSpec   `json:"spec,omitempty"`
	Status VSphereMachineConfigStatus `json:"status,omitempty"`
}

func (c *VSphereMachineConfig) ConvertConfigToConfigGenerateStruct() *VSphereMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &VSphereMachineConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func (c *VSphereMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

func (c *VSphereMachineConfig) SetDefaults() {
	if c.Spec.OSFamily == "" {
		logger.Info("Warning: OS family not specified in machine config specification. Defaulting to Bottlerocket.")
		c.Spec.OSFamily = Bottlerocket
	}

	if len(c.Spec.Folder) <= 0 {
		logger.Info("VSphereMachineConfig Folder is not set or is empty. Defaulting to root vSphere folder.")
	}

	if c.Spec.MemoryMiB <= 0 {
		logger.V(1).Info("VSphereMachineConfig MemoryMiB is not set or is empty. Defaulting to 8192.", "machineConfig", c.Name)
		c.Spec.MemoryMiB = 8192
	}

	if c.Spec.MemoryMiB < 2048 {
		logger.Info("Warning: VSphereMachineConfig MemoryMiB should not be less than 2048. Defaulting to 2048. Recommended memory is 8192.", "machineConfig", c.Name)
		c.Spec.MemoryMiB = 2048
	}

	if c.Spec.NumCPUs <= 0 {
		logger.V(1).Info("VSphereMachineConfig NumCPUs is not set or is empty. Defaulting to 2.", "machineConfig", c.Name)
		c.Spec.NumCPUs = 2
	}
}

// SetUserDefaults initializes Spec.Users for the VSphereMachineConfig with default values.
// This only runs in the CLI, as we support do support user defaults through the webhook.
func (c *VSphereMachineConfig) SetUserDefaults() {
	if len(c.Spec.Users) <= 0 {
		c.Spec.Users = []UserConfiguration{{}}
	}

	if len(c.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		c.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	if len(c.Spec.Users) == 0 || c.Spec.Users[0].Name == "" {
		if c.Spec.OSFamily == Bottlerocket {
			c.Spec.Users[0].Name = constants.BottlerocketDefaultUser
		} else {
			c.Spec.Users[0].Name = constants.UbuntuDefaultUser
		}
		logger.V(1).Info("SSHUsername is not set or is empty for VSphereMachineConfig, using default", "c", c.Name, "user", c.Spec.Users[0].Name)
	}
}

func (c *VSphereMachineConfig) Validate() error {
	if len(c.Spec.Datastore) <= 0 {
		return fmt.Errorf("VSphereMachineConfig %s datastore is not set or is empty", c.Name)
	}
	if len(c.Spec.ResourcePool) <= 0 {
		return fmt.Errorf("VSphereMachineConfig %s VM resourcePool is not set or is empty", c.Name)
	}
	if err := validateOSFamilyUser(c); err != nil {
		return err
	}
	if err := validateHostOSConfig(c.Spec.HostOSConfiguration, c.Spec.OSFamily); err != nil {
		return fmt.Errorf("HostOSConfiguration is invalid for VSphereMachineConfig %s: %v", c.Name, err)
	}

	return nil
}

// ValidateUsers verifies a VSphereMachineConfig object must have a users with ssh authorized keys.
// This validation only runs in VSphereMachineConfig validation webhook, as we support
// auto-generate and import ssh key when creating a cluster via CLI.
func (c *VSphereMachineConfig) ValidateUsers() error {
	if len(c.Spec.Users) == 0 {
		return fmt.Errorf("users is not set for VSphereMachineConfig %s, please provide a user", c.Name)
	}
	if err := validateOSFamilyUser(c); err != nil {
		return err
	}
	if c.Spec.Users[0].Name == "" {
		return fmt.Errorf("users[0].name is not set or is empty for VSphereMachineConfig %s, please provide a username", c.Name)
	}
	if len(c.Spec.Users[0].SshAuthorizedKeys) == 0 || c.Spec.Users[0].SshAuthorizedKeys[0] == "" {
		return fmt.Errorf("users[0].SshAuthorizedKeys is not set or is empty for VSphereMachineConfig %s, please provide a valid ssh authorized key for user %s", c.Name, c.Spec.Users[0].Name)
	}
	return nil
}

func validateOSFamilyUser(machineConfig *VSphereMachineConfig) error {
	if machineConfig.Spec.OSFamily != Bottlerocket {
		return nil
	}
	if machineConfig.Spec.Users == nil || machineConfig.Spec.Users[0].Name != constants.BottlerocketDefaultUser {
		return fmt.Errorf("users[0].name %s is invalid. Please use 'ec2-user' for Bottlerocket", machineConfig.Spec.Users[0].Name)
	}
	return nil
}

// ValidateHasTemplate verifies that a VSphereMachineConfig object has a template.
// Specifying a template is required when submitting an object via webhook,
// as we only support auto-importing templates when creating a cluster via CLI.
func (c *VSphereMachineConfig) ValidateHasTemplate() error {
	if c.Spec.Template == "" {
		return fmt.Errorf("template field is required")
	}
	return nil
}

// +kubebuilder:object:generate=false

// VSphereMachineConfigGenerate Same as VSphereMachineConfig except stripped down for generation of yaml file during generate clusterconfig.
type VSphereMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec VSphereMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereMachineConfigList contains a list of VSphereMachineConfig.
type VSphereMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereMachineConfig{}, &VSphereMachineConfigList{})
}
