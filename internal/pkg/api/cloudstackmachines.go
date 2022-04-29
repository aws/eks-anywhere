package api

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type CloudStackMachineConfigFiller func(m *anywherev1.CloudStackMachineConfig)

const DefaultCloudstackUser = "capc"

func FillCloudStackMachineConfig(m *anywherev1.CloudStackMachineConfig, fillers ...CloudStackMachineConfigFiller) {
	for _, f := range fillers {
		f(m)
	}
}

func WithCloudStackComputeOffering(value string) CloudStackMachineConfigFiller {
	return func(m *anywherev1.CloudStackMachineConfig) {
		m.Spec.ComputeOffering.Name = value
	}
}

func WithCloudStackSSHKey(value string) CloudStackMachineConfigFiller {
	return func(m *anywherev1.CloudStackMachineConfig) {
		setCloudStackSSHKeyForFirstUser(m, value)
	}
}

func setCloudStackSSHKeyForFirstUser(m *anywherev1.CloudStackMachineConfig, key string) {
	if len(m.Spec.Users) == 0 {
		m.Spec.Users = []anywherev1.UserConfiguration{{
			Name: DefaultCloudstackUser,
		}}
	}

	m.Spec.Users[0].SshAuthorizedKeys = []string{key}
}
