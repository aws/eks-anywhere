package v1alpha1

type OSFamily string

const (
	Ubuntu       OSFamily = "ubuntu"
	Bottlerocket OSFamily = "bottlerocket"
	RedHat       OSFamily = "redhat"
)

// UserConfiguration defines the configuration of the user to be added to the VM
type UserConfiguration struct {
	Name              string   `json:"name"`
	SshAuthorizedKeys []string `json:"sshAuthorizedKeys"`
}

type UpgradeRolloutStrategy struct {
	Type			string		`json:"type,omitempty"`
	RollingUpdateParams	*RollingUpdate	`json:"rollingUpdateParams,omitempty"`
}

type RollingUpdate struct {
	MaxSurge	int	`json:"maxSurge,omitempty"`
	MaxUnavailable	int	`json:"maxUnavailable,omitempty"`
}
