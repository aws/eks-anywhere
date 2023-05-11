package v1alpha1

import "fmt"

type OSFamily string

const (
	Ubuntu       OSFamily = "ubuntu"
	Bottlerocket OSFamily = "bottlerocket"
	RedHat       OSFamily = "redhat"
)

// UserConfiguration defines the configuration of the user to be added to the VM.
type UserConfiguration struct {
	Name              string   `json:"name"`
	SshAuthorizedKeys []string `json:"sshAuthorizedKeys"`
}

func defaultMachineConfigUsers(defaultUsername string, users []UserConfiguration) []UserConfiguration {
	if len(users) <= 0 {
		users = []UserConfiguration{{}}
	}
	if len(users[0].SshAuthorizedKeys) <= 0 {
		users[0].SshAuthorizedKeys = []string{""}
	}
	if users[0].Name == "" {
		users[0].Name = defaultUsername
	}

	return users
}

func validateMachineConfigUsers(machineConfigName string, machineConfigKind string, users []UserConfiguration) error {
	if len(users) == 0 {
		return fmt.Errorf("users is not set for %s %s, please provide a user", machineConfigKind, machineConfigName)
	}
	if users[0].Name == "" {
		return fmt.Errorf("users[0].name is not set or is empty for %s %s, please provide a username", machineConfigKind, machineConfigName)
	}
	if len(users[0].SshAuthorizedKeys) == 0 || users[0].SshAuthorizedKeys[0] == "" {
		return fmt.Errorf("users[0].SshAuthorizedKeys is not set or is empty for %s %s, please provide a valid ssh authorized key for user %s", machineConfigKind, machineConfigName, users[0].Name)
	}
	return nil
}
