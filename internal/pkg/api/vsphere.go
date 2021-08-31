package api

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type VSphereConfig struct {
	datacenterConfig    *v1alpha1.VSphereDatacenterConfig
	cpMachineConfig     *v1alpha1.VSphereMachineConfig
	workerMachineConfig *v1alpha1.VSphereMachineConfig
}

type VSphereFiller func(config VSphereConfig)

func AutoFillVSphereProvider(filename string, fillers ...VSphereFiller) ([]byte, error) {
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
	cpName := clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	workerName := clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	vsphereDatacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get vsphere datacenter config from file: %v", err)
	}

	vsphereMachineConfigs, err := v1alpha1.GetVSphereMachineConfigs(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get vsphere machine config from file: %v", err)
	}
	cpMachineConfig, ok := vsphereMachineConfigs[cpName]
	if !ok {
		return nil, fmt.Errorf("unable to find vsphere control plane machine config %v", cpName)
	}
	workerMachineConfig, ok := vsphereMachineConfigs[workerName]
	if !ok {
		return nil, fmt.Errorf("unable to find vsphere worker node machine config %v", workerName)
	}
	config := VSphereConfig{
		datacenterConfig:    vsphereDatacenterConfig,
		cpMachineConfig:     cpMachineConfig,
		workerMachineConfig: workerMachineConfig,
	}
	for _, f := range fillers {
		f(config)
	}

	vsphereDatacenterConfigOutput, err := yaml.Marshal(vsphereDatacenterConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling vsphere datacenter config: %v", err)
	}
	cpMachineConfigOutput, err := yaml.Marshal(cpMachineConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling vsphere control plane machine config: %v", err)
	}
	workerMachineConfigOutput, err := yaml.Marshal(workerMachineConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling vsphere worker node machine config: %v", err)
	}
	vsphereConfigOutput := templater.AppendYamlResources(vsphereDatacenterConfigOutput, cpMachineConfigOutput, workerMachineConfigOutput)

	return vsphereConfigOutput, nil
}

func WithOsFamily(value v1alpha1.OSFamily) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.OSFamily = value
		config.workerMachineConfig.Spec.OSFamily = value
	}
}

func WithNumCPUs(value int) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.NumCPUs = value
		config.workerMachineConfig.Spec.NumCPUs = value
	}
}

func WithDiskGiB(value int) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.DiskGiB = value
		config.workerMachineConfig.Spec.DiskGiB = value
	}
}

func WithMemoryMiB(value int) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.MemoryMiB = value
		config.workerMachineConfig.Spec.MemoryMiB = value
	}
}

func WithTLSInsecure(value bool) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Insecure = value
	}
}

func WithTLSThumbprint(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Thumbprint = value
	}
}

func WithTemplate(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.Template = value
		config.workerMachineConfig.Spec.Template = value
	}
}

func WithStoragePolicyName(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.StoragePolicyName = value
		config.workerMachineConfig.Spec.StoragePolicyName = value
	}
}

func WithSSHUsernameAndAuthorizedKey(username string, key string) VSphereFiller {
	return func(config VSphereConfig) {
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
	}
}

func WithSSHAuthorizedKey(value string) VSphereFiller {
	return func(config VSphereConfig) {
		if len(config.cpMachineConfig.Spec.Users) == 0 {
			config.cpMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{Name: "capv"}}
		}
		if len(config.workerMachineConfig.Spec.Users) == 0 {
			config.workerMachineConfig.Spec.Users = []v1alpha1.UserConfiguration{{Name: "capv"}}
		}
		if len(config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys) == 0 {
			config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		if len(config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys) == 0 {
			config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
		config.cpMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] = value
		config.workerMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] = value
	}
}

func WithServer(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Server = value
	}
}

func WithResourcePool(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.ResourcePool = value
		config.workerMachineConfig.Spec.ResourcePool = value
	}
}

func WithNetwork(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Network = value
	}
}

func WithFolder(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.Folder = value
		config.workerMachineConfig.Spec.Folder = value
	}
}

func WithDatastore(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.cpMachineConfig.Spec.Datastore = value
		config.workerMachineConfig.Spec.Datastore = value
	}
}

func WithDatacenter(value string) VSphereFiller {
	return func(config VSphereConfig) {
		config.datacenterConfig.Spec.Datacenter = value
	}
}

func WithStringFromEnvVar(envVar string, opt func(string) VSphereFiller) VSphereFiller {
	return opt(os.Getenv(envVar))
}

func WithBoolFromEnvVar(envVar string, opt func(bool) VSphereFiller) VSphereFiller {
	return opt(os.Getenv(envVar) == "true")
}
