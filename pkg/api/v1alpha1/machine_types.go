package v1alpha1

type OSFamily string

const (
	Ubuntu       OSFamily = "ubuntu"
	Bottlerocket OSFamily = "bottlerocket"
	Redhat       OSFamily = "redhat"
)

// UserConfiguration defines the configuration of the user to be added to the VM
type UserConfiguration struct {
	Name              string   `json:"name"`
	SshAuthorizedKeys []string `json:"sshAuthorizedKeys"`
}
