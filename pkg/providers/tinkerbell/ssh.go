package tinkerbell

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

const defaultUsername = "ec2-user"

func ensureMachineConfigsHaveAtLeast1User(machines map[string]*v1alpha1.TinkerbellMachineConfig) {
	for _, machine := range machines {
		if len(machine.Spec.Users) == 0 {
			machine.Spec.Users = []v1alpha1.UserConfiguration{{Name: defaultUsername}}
		}
	}
}

func extractUserConfigurationsWithoutSshKeys(machines map[string]*v1alpha1.TinkerbellMachineConfig) []*v1alpha1.UserConfiguration {
	var users []*v1alpha1.UserConfiguration

	for _, machine := range machines {
		if len(machine.Spec.Users[0].SshAuthorizedKeys) == 0 || len(machine.Spec.Users[0].SshAuthorizedKeys[0]) == 0 {
			users = append(users, &machine.Spec.Users[0])
		}
	}

	return users
}

func applySshKeyToUsers(users []*v1alpha1.UserConfiguration, key string) {
	for _, user := range users {
		if len(user.SshAuthorizedKeys) == 0 {
			user.SshAuthorizedKeys = make([]string, 1)
		}

		user.SshAuthorizedKeys[0] = key
	}
}

func stripCommentsFromSshKeys(machines map[string]*v1alpha1.TinkerbellMachineConfig) error {
	for _, machine := range machines {
		key, err := common.StripSshAuthorizedKeyComment(machine.Spec.Users[0].SshAuthorizedKeys[0])
		if err != nil {
			return fmt.Errorf("TinkerbellMachineConfig name=%v: %v", machine.Name, err)
		}
		machine.Spec.Users[0].SshAuthorizedKeys[0] = key
	}

	return nil
}

func (p *Provider) configureSshKeys() error {
	ensureMachineConfigsHaveAtLeast1User(p.machineConfigs)

	users := extractUserConfigurationsWithoutSshKeys(p.machineConfigs)
	if len(users) > 0 {
		publicAuthorizedKey, err := p.keyGenerator.GenerateSSHAuthKey(p.writer)
		if err != nil {
			return err
		}

		applySshKeyToUsers(users, publicAuthorizedKey)
	}

	if err := stripCommentsFromSshKeys(p.machineConfigs); err != nil {
		return fmt.Errorf("stripping ssh key comment: %v", err)
	}

	return nil
}
