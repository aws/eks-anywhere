package api

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type CloudStackMachineConfigFiller func(m *anywherev1.CloudStackMachineConfig)

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
			Name: v1alpha1.DefaultCloudStackUser,
		}}
	}

	m.Spec.Users[0].SshAuthorizedKeys = []string{key}
}
