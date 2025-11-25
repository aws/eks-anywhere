package api

import (
	"os"

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

// WithTags add provided tags to all machines.
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

// WithStringFromEnvVar returns a VSphereMachineConfigFiller function with the value from an envVar passed to it.
func WithStringFromEnvVar(envVar string, opt func(string) VSphereMachineConfigFiller) VSphereMachineConfigFiller {
	return opt(os.Getenv(envVar))
}

func setSSHKeyForFirstUser(m *anywherev1.VSphereMachineConfig, key string) {
	m.SetUserDefaults()
	m.Spec.Users[0].SshAuthorizedKeys = []string{key}
}
