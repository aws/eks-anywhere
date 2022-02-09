package cloudstack

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type CloudStackConfig struct {
	datacenterConfig    *v1alpha1.CloudStackDeploymentConfig
	cpMachineConfig     *v1alpha1.CloudStackMachineConfig
	workerMachineConfig *v1alpha1.CloudStackMachineConfig
	etcdMachineConfig   *v1alpha1.CloudStackMachineConfig
}

type CloudStackFiller func(config CloudStackConfig)

func AutoFillCloudStackProvider(filename string, fillers ...CloudStackFiller) ([]byte, error) {
	var etcdMachineConfig *v1alpha1.CloudStackMachineConfig
	// only to get name of control plane and worker node machine configs
	clusterConfig, err := v1alpha1.GetClusterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}
	if clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return nil, fmt.Errorf("no machineGroupRef defined for control plane")
	}
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) == 0 {
		return nil, fmt.Errorf("no worker nodes defined")
	}
	if clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return nil, fmt.Errorf("no machineGroupRef defined for worker nodes")
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		if clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return nil, fmt.Errorf("no machineGroupRef defined for etcd machines")
		}
	}
	cpName := clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerName := clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudstackDatacenterConfig, err := v1alpha1.GetCloudStackDeploymentConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cloudstack datacenter config from file: %v", err)
	}

	cloudstackMachineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cloudstack machine config from file: %v", err)
	}
	cpMachineConfig, ok := cloudstackMachineConfigs[cpName]
	if !ok {
		return nil, fmt.Errorf("unable to find cloudstack control plane machine config %v", cpName)
	}
	workerMachineConfig, ok := cloudstackMachineConfigs[workerName]
	if !ok {
		return nil, fmt.Errorf("unable to find cloudstack worker node machine config %v", workerName)
	}

	config := CloudStackConfig{
		datacenterConfig:    cloudstackDatacenterConfig,
		cpMachineConfig:     cpMachineConfig,
		workerMachineConfig: workerMachineConfig,
	}

	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig, ok = cloudstackMachineConfigs[clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		if !ok {
			return nil, fmt.Errorf("unable to find cloudstack etcd machine config %v", cpName)
		}
		config.etcdMachineConfig = etcdMachineConfig
	}
	for _, f := range fillers {
		f(config)
	}

	cloudstackDatacenterConfigOutput, err := yaml.Marshal(cloudstackDatacenterConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling cloudstack datacenter config: %v", err)
	}
	cpMachineConfigOutput, err := yaml.Marshal(cpMachineConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling cloudstack control plane machine config: %v", err)
	}
	workerMachineConfigOutput, err := yaml.Marshal(workerMachineConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling cloudstack worker node machine config: %v", err)
	}
	cloudstackConfigOutput := templater.AppendYamlResources(cloudstackDatacenterConfigOutput, cpMachineConfigOutput, workerMachineConfigOutput)
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfigOutput, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return nil, fmt.Errorf("error marshalling cloudstack etcd machine config: %v", err)
		}
		cloudstackConfigOutput = templater.AppendYamlResources(cloudstackConfigOutput, etcdMachineConfigOutput)
	}
	return cloudstackConfigOutput, nil
}

func WithOsFamily(value v1alpha1.OSFamily) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.OSFamily = value
		config.workerMachineConfig.Spec.OSFamily = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.OSFamily = value
		}
	}
}

func WithTLSInsecure(value bool) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Insecure = value
	}
}

func WithTLSThumbprint(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Thumbprint = value
	}
}

func WithTemplate(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.Template = value
		config.workerMachineConfig.Spec.Template = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.Template = value
		}
	}
}

func WithComputeOffering(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.ComputeOffering = value
		config.workerMachineConfig.Spec.ComputeOffering = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.ComputeOffering = value
		}
	}
}

func WithAffinityGroupIds(value []string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.AffinityGroupIds = value
		config.workerMachineConfig.Spec.AffinityGroupIds = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.AffinityGroupIds = value
		}
	}
}

func WithDetails(value map[string]string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.Details = value
		config.workerMachineConfig.Spec.Details = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.Details = value
		}
	}
}

func WithCloudStackConfigNamespace(ns string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Namespace = ns
		config.workerMachineConfig.Namespace = ns
		config.cpMachineConfig.Namespace = ns
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Namespace = ns
		}
	}
}

func WithSSHUsernameAndAuthorizedKey(username string, key string) CloudStackFiller {
	return func(config CloudStackConfig) {
		if len(config.cpMachineConfig.Spec.Users) == 0 {
			config.cpMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
		}
		if len(config.workerMachineConfig.Spec.Users) == 0 {
			config.workerMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
		}
		config.cpMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
			Name:              username,
			SshAuthorizedKeys: []string{key},
		}
		config.workerMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
			Name:              username,
			SshAuthorizedKeys: []string{key},
		}
		if config.etcdMachineConfig != nil {
			if len(config.etcdMachineConfig.Spec.Users) == 0 {
				config.etcdMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{}}
			}
			config.etcdMachineConfig.Spec.Users[0] = v1alpha1.UserConfiguration{
				Name:              username,
				SshAuthorizedKeys: []string{key},
			}
		}
	}
}

func WithSSHAuthorizedKey(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		if len(config.cpMachineConfig.Spec.Users) == 0 {
			config.cpMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{Name: "capc"}}
		}
		if len(config.workerMachineConfig.Spec.Users) == 0 {
			config.workerMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{Name: "capc"}}
		}
		if len(config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys) == 0 {
			config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		if len(config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys) == 0 {
			config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] = value
		config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] = value
		if config.etcdMachineConfig != nil {
			if len(config.cpMachineConfig.Spec.Users) == 0 {
				config.etcdMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{Name: "capc"}}
			}
			if len(config.etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys) == 0 {
				config.etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
			}
			config.etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] = value
		}
	}
}

func WithNetwork(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Network = value
	}
}

func WithDomain(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Domain = value
	}
}

func WithAccount(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Account = value
	}
}

func WithZone(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Zone = value
	}
}

func WithStringFromEnvVar(envVar string, opt func(string) CloudStackFiller) CloudStackFiller {
	return opt(os.Getenv(envVar))
}

func WithBoolFromEnvVar(envVar string, opt func(bool) CloudStackFiller) CloudStackFiller {
	return opt(os.Getenv(envVar) == "true")
}
