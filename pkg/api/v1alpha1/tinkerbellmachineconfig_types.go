package v1alpha1

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TinkerbellMachineConfigSpec defines the desired state of TinkerbellMachineConfig
type TinkerbellMachineConfigSpec struct {
	OSFamily OSFamily            `json:"osFamily"`
	Users    []UserConfiguration `json:"users,omitempty"`
}
