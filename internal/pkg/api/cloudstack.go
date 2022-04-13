package api

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type CloudStackConfig struct {
	datacenterConfig    *v1alpha1.CloudStackDatacenterConfig
	cpMachineConfig     *v1alpha1.CloudStackMachineConfig
	workerMachineConfig *v1alpha1.CloudStackMachineConfig
	etcdMachineConfig   *v1alpha1.CloudStackMachineConfig
}

type CloudStackFiller func(config CloudStackConfig)

func AutoFillCloudStackProvider(filename string, fillers ...CloudStackFiller) ([]byte, error) {
	var etcdMachineConfig *v1alpha1.CloudStackMachineConfig
	// only to get name of control plane and worker node machine configs
	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	cpName := clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerName := clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	cloudstackDatacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(filename)
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
			return nil, fmt.Errorf("unable to find cloudstack etcd machine config %s", clusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
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

func WithCloudStackTemplate(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.Template.Name = value
		config.workerMachineConfig.Spec.Template.Name = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.Template.Name = value
		}
	}
}

func WithCloudStackComputeOffering(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.ComputeOffering.Name = value
		config.workerMachineConfig.Spec.ComputeOffering.Name = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.ComputeOffering.Name = value
		}
	}
}

func WithCloudStackManagementServer(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.ManagementApiEndpoint = value
	}
}

func WithCloudStackAffinityGroupIds(value []string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.AffinityGroupIds = value
		config.workerMachineConfig.Spec.AffinityGroupIds = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.AffinityGroupIds = value
		}
	}
}

func WithUserCustomDetails(value map[string]string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.cpMachineConfig.Spec.UserCustomDetails = value
		config.workerMachineConfig.Spec.UserCustomDetails = value
		if config.etcdMachineConfig != nil {
			config.etcdMachineConfig.Spec.UserCustomDetails = value
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

func WithCloudStackSSHUsernameAndAuthorizedKey(username string, key string) CloudStackFiller {
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

func WithCloudStackSSHAuthorizedKey(value string) CloudStackFiller {
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

func WithCloudStackDomain(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Domain = value
	}
}

func WithCloudStackAccount(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Account = value
	}
}

func WithCloudStackZone(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Zones[0].Name = value
	}
}

func WithCloudStackNetwork(value string) CloudStackFiller {
	return func(config CloudStackConfig) {
		config.datacenterConfig.Spec.Zones[0].Network.Name = value
	}
}

func WithCloudStackStringFromEnvVar(envVar string, opt func(string) CloudStackFiller) CloudStackFiller {
	return opt(os.Getenv(envVar))
}
