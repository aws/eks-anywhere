package factory

import (
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

type ProviderFactory struct {
	DockerClient              docker.ProviderClient
	DockerKubectlClient       docker.ProviderKubectlClient
	VSphereGovcClient         vsphere.ProviderGovcClient
	VSphereKubectlClient      vsphere.ProviderKubectlClient
	CloudStackCmkClient       cloudstack.ProviderCmkClient
	CloudStackKubectlClient   cloudstack.ProviderKubectlClient
	TinkerbellKubectlClient   tinkerbell.ProviderKubectlClient
	TinkerbellClients         tinkerbell.TinkerbellClients
	SnowKubectlClient         snow.ProviderKubectlClient
	Writer                    filewriter.FileWriter
	ClusterResourceSetManager vsphere.ClusterResourceSetManager
}

func (p *ProviderFactory) BuildProvider(clusterConfigFileName string, clusterConfig *v1alpha1.Cluster, skipIpCheck bool, hardwareConfigFile string, skipPowerActions bool) (providers.Provider, error) {
	switch clusterConfig.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		machineConfigs, err := v1alpha1.GetVSphereMachineConfigs(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFileName, err)
		}
		return vsphere.NewProvider(datacenterConfig, machineConfigs, clusterConfig, p.VSphereGovcClient, p.VSphereKubectlClient, p.Writer, time.Now, skipIpCheck, p.ClusterResourceSetManager), nil
	case v1alpha1.CloudStackDatacenterKind:
		datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFileName, err)
		}
		return cloudstack.NewProvider(datacenterConfig, machineConfigs, clusterConfig, p.CloudStackKubectlClient, p.CloudStackCmkClient, p.Writer, time.Now, skipIpCheck), nil
	case v1alpha1.SnowDatacenterKind:
		return snow.NewProvider(p.SnowKubectlClient, p.Writer, time.Now), nil
	case v1alpha1.TinkerbellDatacenterKind:
		datacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		machineConfigs, err := v1alpha1.GetTinkerbellMachineConfigs(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFileName, err)
		}
		return tinkerbell.NewProvider(datacenterConfig, machineConfigs, clusterConfig, p.Writer, p.TinkerbellKubectlClient, p.TinkerbellClients, time.Now, skipIpCheck, hardwareConfigFile, skipPowerActions), nil
	case v1alpha1.DockerDatacenterKind:
		datacenterConfig, err := v1alpha1.GetDockerDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		return docker.NewProvider(datacenterConfig, p.DockerClient, p.DockerKubectlClient, time.Now), nil
	}
	return nil, fmt.Errorf("no provider support for datacenter kind: %s", clusterConfig.Spec.DatacenterRef.Kind)
}
