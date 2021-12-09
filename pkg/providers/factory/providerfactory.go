package factory

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

type ProviderFactory struct {
	DockerClient                docker.ProviderClient
	DockerKubectlClient         docker.ProviderKubectlClient
	VSphereGovcClient           vsphere.ProviderGovcClient
	VSphereKubectlClient        vsphere.ProviderKubectlClient
	CloudStackCloudMonkeyClient cloudstack.ProviderCloudMonkeyClient
	CloudStackKubectlClient     cloudstack.ProviderKubectlClient
	TinkerbellKubectlClient     tinkerbell.ProviderKubectlClient
	Writer                      filewriter.FileWriter
	ClusterResourceSetManager   vsphere.ClusterResourceSetManager
}

func (p *ProviderFactory) BuildProvider(clusterConfigFileName string, clusterConfig *v1alpha1.Cluster, skipIpCheck bool, hardwareConfigFile string) (providers.Provider, error) {
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
	case v1alpha1.CloudStackDeploymentKind:
		datacenterConfig, err := v1alpha1.GetCloudStackDeploymentConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get deployment config from file %s: %v", clusterConfigFileName, err)
		}
		machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFileName, err)
		}
		return cloudstack.NewProvider(datacenterConfig, machineConfigs, clusterConfig, p.CloudStackCloudMonkeyClient, p.CloudStackKubectlClient, p.Writer, time.Now, skipIpCheck, p.ClusterResourceSetManager), nil
	case v1alpha1.TinkerbellDatacenterKind:
		datacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		machineConfigs, err := v1alpha1.GetTinkerbellMachineConfigs(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFileName, err)
		}
		return tinkerbell.NewProvider(datacenterConfig, machineConfigs, clusterConfig, p.TinkerbellKubectlClient, time.Now, hardwareConfigFile), nil
	case v1alpha1.DockerDatacenterKind:
		datacenterConfig, err := v1alpha1.GetDockerDatacenterConfig(clusterConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFileName, err)
		}
		return docker.NewProvider(datacenterConfig, p.DockerClient, p.DockerKubectlClient, time.Now), nil
	}
	return nil, errors.New("valid providers include: " + constants.DockerProviderName + ", " + constants.VSphereProviderName)
}
