package hardware

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// TypeMeta constants for defining Kubernetes TypeMeta data in Kubernetes objects.
const (
	tinkerbellAPIVersion   = "tinkerbell.org/v1alpha1"
	tinkerbellHardwareKind = "Hardware"
	tinkerbellBMCKind      = "BaseboardManagement"

	secretKind       = "Secret"
	secretAPIVersion = "v1"
)

func newHardwareTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       tinkerbellHardwareKind,
		APIVersion: tinkerbellAPIVersion,
	}
}

func newBaseboardManagementTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       tinkerbellBMCKind,
		APIVersion: tinkerbellAPIVersion,
	}
}

func newSecretTypeMeta() v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       secretKind,
		APIVersion: secretAPIVersion,
	}
}
