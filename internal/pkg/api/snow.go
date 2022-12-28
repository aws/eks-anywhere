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
}

type SnowFiller func(config SnowConfig)

// SnowToConfigFiller transforms a set of SnowFiller's in a single ClusterConfigFiller.
func SnowToConfigFiller(fillers ...SnowFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		updateSnow(c, fillers...)
	}
}

func updateSnow(config *cluster.Config, fillers ...SnowFiller) {
	sc := SnowConfig{
		datacenterConfig: config.SnowDatacenter,
		machineConfigs:   config.SnowMachineConfigs,
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

func WithSnowInstanceTypeForAllMachines(instanceType anywherev1.SnowInstanceType) SnowFiller {
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
