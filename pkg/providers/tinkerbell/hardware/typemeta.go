package hardware

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// TypeMeta constants for defining Kubernetes TypeMeta data in Kubernetes objects.
const (
	// TODO(pokearu) update API version once upstream is changed.
	rufioAPIVersion        = "bmc.tinkerbell.org/v1alpha1"
	tinkerbellAPIVersion   = "tinkerbell.org/v1alpha1"
	tinkerbellHardwareKind = "Hardware"
	tinkerbellBMCKind      = "Machine"

	secretKind       = "Secret"
	secretAPIVersion = "v1"
)

func newHardwareTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       tinkerbellHardwareKind,
		APIVersion: tinkerbellAPIVersion,
	}
}

func newMachineTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       tinkerbellBMCKind,
		APIVersion: rufioAPIVersion,
	}
}

func newSecretTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       secretKind,
		APIVersion: secretAPIVersion,
	}
}
