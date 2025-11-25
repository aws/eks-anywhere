package framework

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// OS refers to the Operating System to be used for Machine configs.
type OS string

const (
	// DockerOS corresponds to a dummy Docker OS that will be ignored when creating the cluster config.
	DockerOS OS = "docker"
	// Ubuntu2204 corresponds to Ubuntu 22.04 OS.
	Ubuntu2204 OS = "ubuntu-2204"
	// Ubuntu2004 corresponds to Ubuntu 20.04 OS. We don't add the version number in the string to facilitate backwards compatibility.
	Ubuntu2004 OS = "ubuntu"
	// Bottlerocket1 corresponds to Bottlerocket OS. We don't add the version number in the string to facilitate backwards compatibility.
	Bottlerocket1 OS = "bottlerocket"
	// RedHat9 corresponds to Red Hat 9 OS.
	RedHat9 OS = "redhat-9"
	// RedHat8 corresponds to Red Hat 8 OS. We don't add the version number in the string to facilitate backwards compatibility.
	RedHat8 OS = "redhat"
)

var osFamiliesForOS = map[OS]anywherev1.OSFamily{
	Ubuntu2204:    anywherev1.Ubuntu,
	Ubuntu2004:    anywherev1.Ubuntu,
	Bottlerocket1: anywherev1.Bottlerocket,
	RedHat8:       anywherev1.RedHat,
	RedHat9:       anywherev1.RedHat,
}
