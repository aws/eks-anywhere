package api

import (
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type CloudStackConfig struct {
	datacenterConfig *anywherev1.CloudStackDatacenterConfig
	machineConfigs   map[string]*anywherev1.CloudStackMachineConfig
}

type CloudStackFiller func(config CloudStackConfig)

func AutoFillCloudStackProvider(filename string, fillers ...CloudStackFiller) ([]byte, error) {
	config, err := cluster.ParseConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

	cloudStackConfig := CloudStackConfig{
		datacenterConfig: config.CloudStackDatacenter,
		machineConfigs:   config.CloudStackMachineConfigs,
	}

	for _, f := range fillers {
		f(cloudStackConfig)
	}

	resources := make([]interface{}, 0, len(cloudStackConfig.machineConfigs)+1)
	resources = append(resources, cloudStackConfig.datacenterConfig)
	for _, m := range cloudStackConfig.machineConfigs {
		resources = append(resources, m)
	}

	yamlResources := make([][]byte, 0, len(resources))
	for _, r := range resources {
		yamlContent, err := yaml.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshalling cloudstack resource: %v", err)
		}

		yamlResources = append(yamlResources, yamlContent)
	}

	return templater.AppendYamlResources(yamlResources...), nil
}

func WithCloudStackComputeOfferingForAllMachines(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.ComputeOffering.Name = value
		}
	}
}

func WithCloudStackManagementServer(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.AvailabilityZones[0].ManagementApiEndpoint = value
	}
}

func WithCloudStackAz(az anywherev1.CloudStackAvailabilityZone) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.AvailabilityZones = append(config.datacenterConfig.Spec.AvailabilityZones, az)
	}
}

func RemoveCloudStackAzs() CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.AvailabilityZones = make([]anywherev1.CloudStackAvailabilityZone, 0)
	}
}

func WithCloudStackAffinityGroupIds(value []string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.AffinityGroupIds = value
		}
	}
}

func WithUserCustomDetails(value map[string]string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.UserCustomDetails = value
		}
	}
}

func WithSymlinks(value map[string]string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Symlinks = value
		}
	}
}

func WithCloudStackTemplateForAllMachines(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			m.Spec.Template.Name = value
		}
	}
}

func WithCloudStackConfigNamespace(ns string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Namespace = ns
		for _, m := range config.machineConfigs {
			m.Namespace = ns
		}
	}
}

func WithCloudStackSSHAuthorizedKey(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, m := range config.machineConfigs {
			if len(m.Spec.Users) == 0 {
				m.Spec.Users = []anywherev1.UserConfiguration{{Name: "capc"}}
			}
			m.Spec.Users[0].SshAuthorizedKeys = []string{value}
		}
	}
}

func WithCloudStackDomain(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, az := range config.datacenterConfig.Spec.AvailabilityZones {
			az.Domain = value
		}
	}
}

func WithCloudStackAccount(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		for _, az := range config.datacenterConfig.Spec.AvailabilityZones {
			az.Account = value
		}
	}
}

func WithCloudStackZone(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.AvailabilityZones[0].Zone.Name = value
	}
}

func WithCloudStackNetwork(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.AvailabilityZones[0].Zone.Network.Name = value
	}
}

func WithCloudStackStringFromEnvVar(envVar string, opt func(string) CloudStackFiller) CloudStackFiller {
	return opt(os.Getenv(envVar))
}

func WithCloudStackAzFromEnvVars(cloudstackAccountVar, cloudstackDomainVar, cloudstackZoneVar, cloudstackNetworkVar, cloudstackManagementServerVar string, opt func(zone anywherev1.CloudStackAvailabilityZone) CloudStackFiller) CloudStackFiller {
	az := anywherev1.CloudStackAvailabilityZone{
		Name:           strings.ToLower(fmt.Sprintf("az-%s", os.Getenv(cloudstackZoneVar))),
		CredentialsRef: "global",
		Zone: anywherev1.CloudStackZone{
			Name: os.Getenv(cloudstackZoneVar),
			Network: anywherev1.CloudStackResourceIdentifier{
				Name: os.Getenv(cloudstackNetworkVar),
			},
		},
		Domain:                os.Getenv(cloudstackDomainVar),
		Account:               os.Getenv(cloudstackAccountVar),
		ManagementApiEndpoint: os.Getenv(cloudstackManagementServerVar),
	}
	return opt(az)
}

func WithCloudStackMachineConfig(name string, fillers ...CloudStackMachineConfigFiller) CloudStackFiller {
	return func(config CloudStackConfig) {
		m, ok := config.machineConfigs[name]
		if !ok {
			m = &anywherev1.CloudStackMachineConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.CloudStackMachineConfigKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			config.machineConfigs[name] = m
		}

		FillCloudStackMachineConfig(m, fillers...)
	}
}
