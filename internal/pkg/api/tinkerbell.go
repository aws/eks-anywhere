package api

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
)

type TinkerbellConfig struct {
	clusterName      string
	datacenterConfig *anywherev1.TinkerbellDatacenterConfig
	machineConfigs   map[string]*anywherev1.TinkerbellMachineConfig
	templateConfigs  map[string]*anywherev1.TinkerbellTemplateConfig
}

// TinkerbellFiller updates a TinkerbellConfig.
type TinkerbellFiller func(config TinkerbellConfig)

// TinkerbellToConfigFiller transforms a set of TinkerbellFiller's in a single ClusterConfigFiller.
func TinkerbellToConfigFiller(fillers ...TinkerbellFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		updateTinkerbell(c, fillers...)
	}
}

// updateTinkerbell updates the Tinkerbell datacenter, machine configs and
// template configs in the cluster.Config by applying all the fillers.
func updateTinkerbell(config *cluster.Config, fillers ...TinkerbellFiller) {
	tc := TinkerbellConfig{
		clusterName:      config.Cluster.Name,
		datacenterConfig: config.TinkerbellDatacenter,
		machineConfigs:   config.TinkerbellMachineConfigs,
		templateConfigs:  config.TinkerbellTemplateConfigs,
	}

	for _, f := range fillers {
		f(tc)
	}
}

func WithTinkerbellServer(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		config.datacenterConfig.Spec.TinkerbellIP = value
	}
}

func WithTinkerbellOSImageURL(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		config.datacenterConfig.Spec.OSImageURL = value
	}
}

// WithTinkerbellCPMachineConfigOSImageURL sets the OSImageURL & OSFamily for control-plane machine config.
func WithTinkerbellCPMachineConfigOSImageURL(imageURL string, OSFamily anywherev1.OSFamily) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		clusterName := config.clusterName
		cpName := providers.GetControlPlaneNodeName(clusterName)
		cpMachineConfig := config.machineConfigs[cpName]
		cpMachineConfig.Spec.OSImageURL = imageURL
		cpMachineConfig.Spec.OSFamily = OSFamily
		config.machineConfigs[cpName] = cpMachineConfig
	}
}

// WithTinkerbellWorkerMachineConfigOSImageURL sets the OSImageURL & OSFamily for worker machine config.
func WithTinkerbellWorkerMachineConfigOSImageURL(imageURL string, OSFamily anywherev1.OSFamily) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		clusterName := config.clusterName
		workerMachineConfig := config.machineConfigs[clusterName]
		workerMachineConfig.Spec.OSImageURL = imageURL
		workerMachineConfig.Spec.OSFamily = OSFamily
		config.machineConfigs[clusterName] = workerMachineConfig
	}
}

// WithHookImagesURLPath modify HookImagesURL, it's useful for airgapped testing.
func WithHookImagesURLPath(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		config.datacenterConfig.Spec.HookImagesURLPath = value
	}
}

func WithStringFromEnvVarTinkerbell(envVar string, opt func(string) TinkerbellFiller) TinkerbellFiller {
	return opt(os.Getenv(envVar))
}

func WithOsFamilyForAllTinkerbellMachines(value anywherev1.OSFamily) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.OSFamily = value
		}
	}
}

func WithImageUrlForAllTinkerbellMachines(value string) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		for _, t := range config.templateConfigs {
			for _, task := range t.Spec.Template.Tasks {
				for _, action := range task.Actions {
					if action.Name == "stream-image" {
						action.Environment["IMG_URL"] = value
					}
				}
			}
		}
	}
}

func WithSSHAuthorizedKeyForAllTinkerbellMachines(key string) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		for _, m := range config.machineConfigs {
			if len(m.Spec.Users) == 0 {
				m.Spec.Users = []anywherev1.UserConfiguration{{}}
			}

			m.Spec.Users[0].Name = "ec2-user"
			m.Spec.Users[0].SshAuthorizedKeys = []string{key}
		}
	}
}

func WithHardwareSelectorLabels() TinkerbellFiller {
	return func(config TinkerbellConfig) {
		clusterName := config.clusterName
		cpName := providers.GetControlPlaneNodeName(clusterName)
		workerName := clusterName

		cpMachineConfig := config.machineConfigs[cpName]
		cpMachineConfig.Spec.HardwareSelector = map[string]string{HardwareLabelTypeKeyName: ControlPlane}
		config.machineConfigs[cpName] = cpMachineConfig

		workerMachineConfig := config.machineConfigs[workerName]
		workerMachineConfig.Spec.HardwareSelector = map[string]string{HardwareLabelTypeKeyName: Worker}
		config.machineConfigs[workerName] = workerMachineConfig
	}
}

func WithTinkerbellEtcdMachineConfig() TinkerbellFiller {
	return func(config TinkerbellConfig) {
		clusterName := config.clusterName
		name := providers.GetEtcdNodeName(clusterName)

		_, ok := config.machineConfigs[name]
		if !ok {
			m := &anywherev1.TinkerbellMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.TinkerbellMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: anywherev1.TinkerbellMachineConfigSpec{
					HardwareSelector: map[string]string{HardwareLabelTypeKeyName: ExternalEtcd},
					TemplateRef: anywherev1.Ref{
						Name: clusterName,
						Kind: anywherev1.TinkerbellTemplateConfigKind,
					},
				},
			}
			config.machineConfigs[name] = m
		}
	}
}

// RemoveTinkerbellWorkerMachineConfig removes the worker node TinkerbellMachineConfig for single node clusters.
func RemoveTinkerbellWorkerMachineConfig() TinkerbellFiller {
	return func(config TinkerbellConfig) {
		clusterName := config.clusterName
		delete(config.machineConfigs, clusterName)
	}
}

// WithStringFromEnvVarTinkerbellMachineFiller runs a TinkerbellMachineFiller function with an envVar value.
func WithStringFromEnvVarTinkerbellMachineFiller(envVar string, opt func(string) TinkerbellMachineFiller) TinkerbellMachineFiller {
	return opt(os.Getenv(envVar))
}

// TinkerbellMachineFiller updates a TinkerbellMachineConfig.
type TinkerbellMachineFiller func(machineConfig *anywherev1.TinkerbellMachineConfig)

// WithSSHAuthorizedKeyForTinkerbellMachineConfig updates the SSHAuthorizedKey for a TinkerbellMachineConfig.
func WithSSHAuthorizedKeyForTinkerbellMachineConfig(key string) TinkerbellMachineFiller {
	return func(machineConfig *anywherev1.TinkerbellMachineConfig) {
		if len(machineConfig.Spec.Users) == 0 {
			machineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
		}

		machineConfig.Spec.Users[0].Name = "ec2-user"
		machineConfig.Spec.Users[0].SshAuthorizedKeys = []string{key}
	}
}

// WithOsFamilyForTinkerbellMachineConfig updates the OSFamily of a TinkerbellMachineConfig.
func WithOsFamilyForTinkerbellMachineConfig(value anywherev1.OSFamily) TinkerbellMachineFiller {
	return func(machineConfig *anywherev1.TinkerbellMachineConfig) {
		machineConfig.Spec.OSFamily = value
	}
}

// WithCustomTinkerbellMachineConfig generates a TinkerbellMachineConfig from a hardware selector.
func WithCustomTinkerbellMachineConfig(selector string, machineConfigFillers ...TinkerbellMachineFiller) TinkerbellFiller {
	return func(config TinkerbellConfig) {
		if _, ok := config.machineConfigs[selector]; !ok {
			m := &anywherev1.TinkerbellMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.TinkerbellMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: selector,
				},
				Spec: anywherev1.TinkerbellMachineConfigSpec{
					HardwareSelector: map[string]string{HardwareLabelTypeKeyName: selector},
				},
			}

			for _, f := range machineConfigFillers {
				f(m)
			}

			config.machineConfigs[selector] = m
		}
	}
}
