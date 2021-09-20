package clustermarshaller

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func MarshalClusterSpec(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) ([]byte, error) {
	convertedClusterGenerateConfig := clusterSpec.ConvertConfigToConfigGenerateStruct()
	clusterObj, err := yaml.Marshal(convertedClusterGenerateConfig)
	if err != nil {
		return nil, fmt.Errorf("error outputting cluster yaml: %v", err)
	}
	datacenterObj, err := yaml.Marshal(datacenterConfig)
	if err != nil {
		return nil, fmt.Errorf("error outputting datacenter yaml: %v", err)
	}
	resources := [][]byte{clusterObj, datacenterObj}
	for _, m := range machineConfigs {
		mObj, err := yaml.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("error outputting machine yaml: %v", err)
		}
		resources = append(resources, mObj)
	}
	if clusterSpec.GitOpsConfig != nil {
		convertedGitOpsGenerateConfig := clusterSpec.GitOpsConfig.ConvertConfigToConfigGenerateStruct()
		gitopsObj, err := yaml.Marshal(convertedGitOpsGenerateConfig)
		if err != nil {
			return nil, fmt.Errorf("error outputting gitops config yaml: %v", err)
		}
		resources = append(resources, gitopsObj)
	}
	if clusterSpec.OIDCConfig != nil {
		convertedOIDCGenerateConfig := clusterSpec.OIDCConfig.ConvertConfigToConfigGenerateStruct()
		oidcObj, err := yaml.Marshal(convertedOIDCGenerateConfig)
		if err != nil {
			return nil, fmt.Errorf("error outputting oidc config yaml: %v", err)
		}
		resources = append(resources, oidcObj)
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
