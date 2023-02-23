package api

import (
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

type SnowConfig struct {
	datacenterConfig *anywherev1.SnowDatacenterConfig
	machineConfigs   map[string]*anywherev1.SnowMachineConfig
	ipPools          map[string]*anywherev1.SnowIPPool
}

type SnowFiller func(config SnowConfig)

// SnowToConfigFiller transforms a set of SnowFiller's in a single ClusterConfigFiller.
func SnowToConfigFiller(fillers ...SnowFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		updateSnow(c, fillers...)
	}
}

func updateSnow(config *cluster.Config, fillers ...SnowFiller) {
	if config.SnowIPPools == nil {
		config.SnowIPPools = map[string]*anywherev1.SnowIPPool{}
	}

	sc := SnowConfig{
		datacenterConfig: config.SnowDatacenter,
		machineConfigs:   config.SnowMachineConfigs,
		ipPools:          config.SnowIPPools,
	}

	for _, f := range fillers {
		f(sc)
	}
}

func WithSnowStringFromEnvVar(envVar string, opt func(string) SnowFiller) SnowFiller {
	return opt(os.Getenv(envVar))
}

func WithSnowAMIIDForAllMachines(id string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.AMIID = id
		}
	}
}

// WithSnowInstanceTypeForAllMachines specifies an instance type for all the snow machine configs.
func WithSnowInstanceTypeForAllMachines(instanceType string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.InstanceType = instanceType
		}
	}
}

func WithSnowPhysicalNetworkConnectorForAllMachines(connectorType anywherev1.PhysicalNetworkConnectorType) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.PhysicalNetworkConnector = connectorType
		}
	}
}

func WithSnowSshKeyNameForAllMachines(keyName string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.SshKeyName = keyName
		}
	}
}

func WithSnowDevicesForAllMachines(devices string) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Devices = strings.Split(devices, ",")
		}
	}
}

func WithSnowMachineConfig(name string, fillers ...SnowMachineConfigFiller) SnowFiller {
	return func(config SnowConfig) {
		m, ok := config.machineConfigs[name]
		if !ok {
			m = &anywherev1.SnowMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.SnowMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			config.machineConfigs[name] = m
		}

		FillSnowMachineConfig(m, fillers...)
	}
}

// WithOsFamilyForAllSnowMachines sets the OSFamily in the SnowMachineConfig.
func WithOsFamilyForAllSnowMachines(value anywherev1.OSFamily) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.OSFamily = value
		}
	}
}

// WithChangeForAllSnowMachines applies the same change to all SnowMachineConfigs.
func WithChangeForAllSnowMachines(change SnowMachineConfigFiller) SnowFiller {
	return func(config SnowConfig) {
		for _, m := range config.machineConfigs {
			change(m)
		}
	}
}

// WithSnowIPPool sets a SnowIPPool.
func WithSnowIPPool(name, ipStart, ipEnd, gateway, subnet string) SnowFiller {
	return func(config SnowConfig) {
		config.ipPools[name] = &anywherev1.SnowIPPool{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.SnowIPPoolKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: anywherev1.SnowIPPoolSpec{
				Pools: []anywherev1.IPPool{
					{
						IPStart: ipStart,
						IPEnd:   ipEnd,
						Subnet:  subnet,
						Gateway: gateway,
					},
				},
			},
		}
	}
}
