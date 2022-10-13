package api

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type VSphereMachineConfigFiller func(m *anywherev1.VSphereMachineConfig)

func FillVSphereMachineConfig(m *anywherev1.VSphereMachineConfig, fillers ...VSphereMachineConfigFiller) {
	for _, f := range fillers {
		f(m)
	}
}

func WithVSphereMachineDefaultValues() VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.DiskGiB = anywherev1.DefaultVSphereDiskGiB
		m.Spec.NumCPUs = anywherev1.DefaultVSphereNumCPUs
		m.Spec.MemoryMiB = anywherev1.DefaultVSphereMemoryMiB
		m.Spec.OSFamily = anywherev1.DefaultVSphereOSFamily
	}
}

func WithDatastore(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.Datastore = value
	}
}

func WithFolder(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.Folder = value
	}
}

func WithTags(value []string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.TagIDs = value
	}
}

func WithResourcePool(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.ResourcePool = value
	}
}

func WithStoragePolicyName(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.StoragePolicyName = value
	}
}

func WithTemplate(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		m.Spec.Template = value
	}
}

func WithSSHKey(value string) VSphereMachineConfigFiller {
	return func(m *anywherev1.VSphereMachineConfig) {
		setSSHKeyForFirstUser(m, value)
	}
}

func setSSHKeyForFirstUser(m *anywherev1.VSphereMachineConfig, key string) {
	if len(m.Spec.Users) == 0 {
		m.Spec.Users = []anywherev1.UserConfiguration{{}}
	}

	m.Spec.Users[0].SshAuthorizedKeys = []string{key}
}
