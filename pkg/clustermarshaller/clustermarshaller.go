package clustermarshaller

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func MarshalClusterSpec(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) ([]byte, error) {
	marshallables := make([]v1alpha1.Marshallable, 0, 5+len(machineConfigs))
	marshallables = append(marshallables,
		clusterSpec.Cluster.ConvertConfigToConfigGenerateStruct(),
		datacenterConfig.Marshallable(),
	)

	for _, machineConfig := range machineConfigs {
		marshallables = append(marshallables, machineConfig.Marshallable())
	}

	if clusterSpec.GitOpsConfig != nil {
		marshallables = append(marshallables, clusterSpec.GitOpsConfig.ConvertConfigToConfigGenerateStruct())
	}

	if clusterSpec.OIDCConfig != nil {
		marshallables = append(marshallables, clusterSpec.OIDCConfig.ConvertConfigToConfigGenerateStruct())
	}
	if clusterSpec.AddOnAWSIamConfig != nil {
		marshallables = append(marshallables, clusterSpec.AddOnAWSIamConfig.ConvertConfigToConfigGenerateStruct())
	}

	resources := make([][]byte, 0, len(marshallables))
	for _, marshallable := range marshallables {
		resource, err := yaml.Marshal(marshallable)
		if err != nil {
			return nil, fmt.Errorf("failed marshalling resource for cluster spec: %v", err)
		}
		resources = append(resources, resource)
	}
	return templater.AppendYamlResources(resources...), nil
}

func WriteClusterConfig(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig, writer filewriter.FileWriter) error {
	resourcesSpec, err := MarshalClusterSpec(clusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		return err
	}
	if filePath, err := writer.Write(fmt.Sprintf("%s-eks-a-cluster.yaml", clusterSpec.Cluster.ObjectMeta.Name), resourcesSpec, filewriter.PersistentFile); err != nil {
		err = fmt.Errorf("error writing eks-a cluster config file into %s: %v", filePath, err)
		return err
	}

	return nil
}
