package api

import (
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
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

// WithDHCP configures one single primary DNI using DHCP for IP allocation.
func WithDHCP() SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.Network.DirectNetworkInterfaces = []anywherev1.SnowDirectNetworkInterface{
			{
				DHCP:    true,
				Primary: true,
			},
		}
	}
}

// WithStaticIP configures one single primary DNI using static ip for IP allocation.
func WithStaticIP(poolName string) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		m.Spec.Network.DirectNetworkInterfaces = []anywherev1.SnowDirectNetworkInterface{
			{
				Primary: true,
				IPPoolRef: &anywherev1.Ref{
					Kind: anywherev1.SnowIPPoolKind,
					Name: poolName,
				},
			},
		}
	}
}

// WithSnowContainersVolumeSize sets the container volume size for a SnowMachineConfig.
func WithSnowContainersVolumeSize(size int64) SnowMachineConfigFiller {
	return func(m *anywherev1.SnowMachineConfig) {
		if m.Spec.ContainersVolume == nil {
			m.Spec.ContainersVolume = &snowv1.Volume{}
		}
		m.Spec.ContainersVolume.Size = size
	}
}
