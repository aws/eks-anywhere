package api

import (
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type SnowMachineConfigFiller func(m *anywherev1.SnowMachineConfig)

func FillSnowMachineConfig(m *anywherev1.SnowMachineConfig, fillers ...SnowMachineConfigFiller) {
	for _, f := range fillers {
		f(m)
	}
}

func WithSnowMachineDefaultValues() SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.InstanceType = anywherev1.DefaultSnowInstanceType
		m.Spec.PhysicalNetworkConnector = anywherev1.DefaultSnowPhysicalNetworkConnectorType
		m.Spec.SshKeyName = anywherev1.DefaultSnowSSHKeyName
	}
}

func WithSnowAMIID(id string) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.AMIID = id
	}
}

// WithSnowInstanceType specifies an instance type for the snow machine config.
func WithSnowInstanceType(instanceType string) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.InstanceType = instanceType
	}
}

func WithSnowPhysicalNetworkConnector(connectorType anywherev1.PhysicalNetworkConnectorType) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.PhysicalNetworkConnector = connectorType
	}
}

func WithSnowSshKeyName(keyName string) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.SshKeyName = keyName
	}
}

func WithSnowDevices(devices string) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.Devices = strings.Split(devices, ",")
	}
}
