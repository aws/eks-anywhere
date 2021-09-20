package clustermarshaller

import (
	"fmt"

	"sigs.k8s.io/yaml"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func MarshalClusterSpec(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) ([]byte, error) {
	copiedStruct := copyToClusterGenerateStruct(clusterSpec.Cluster)
	clusterObj, err := yaml.Marshal(copiedStruct)
	// clusterObj, err := yaml.Marshal(clusterSpec.Cluster)
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
		copiedGitOps := copyToGitOpsConfigGenerateStruct(clusterSpec.GitOpsConfig)
		gitopsObj, err := yaml.Marshal(copiedGitOps)
		// gitopsObj, err := yaml.Marshal(clusterSpec.GitOpsConfig)
		if err != nil {
			return nil, fmt.Errorf("error outputting gitops config yaml: %v", err)
		}
		resources = append(resources, gitopsObj)
	}
	if clusterSpec.OIDCConfig != nil {
		copiedOIDC := copyToOIDCConfigGenerateStruct(clusterSpec.OIDCConfig)
		oidcObj, err := yaml.Marshal(copiedOIDC)
		// oidcObj, err := yaml.Marshal(clusterSpec.OIDCConfig)
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

func copyToGitOpsConfigGenerateStruct(gitopsConfig *eksav1alpha1.GitOpsConfig) *eksav1alpha1.GitOpsConfigGenerate {
	config := &eksav1alpha1.GitOpsConfigGenerate{
		TypeMeta: gitopsConfig.TypeMeta,
		ObjectMeta: eksav1alpha1.ObjectMeta{
			Name:        gitopsConfig.Name,
			Annotations: gitopsConfig.Annotations,
			Namespace:   gitopsConfig.Namespace,
		},
		Spec: gitopsConfig.Spec,
	}

	return config
}

func copyToOIDCConfigGenerateStruct(oidcConfig *eksav1alpha1.OIDCConfig) *eksav1alpha1.OIDCConfigGenerate {
	config := &eksav1alpha1.OIDCConfigGenerate{
		TypeMeta: oidcConfig.TypeMeta,
		ObjectMeta: eksav1alpha1.ObjectMeta{
			Name:        oidcConfig.Name,
			Annotations: oidcConfig.Annotations,
			Namespace:   oidcConfig.Namespace,
		},
		Spec: oidcConfig.Spec,
	}

	return config
}

func copyToClusterGenerateStruct(cluster *eksav1alpha1.Cluster) *eksav1alpha1.ClusterGenerate {
	config := &eksav1alpha1.ClusterGenerate{
		TypeMeta: cluster.TypeMeta,
		ObjectMeta: eksav1alpha1.ObjectMeta{
			Name:        cluster.Name,
			Annotations: cluster.Annotations,
			Namespace:   cluster.Namespace,
		},
		Spec: cluster.Spec,
	}

	return config
}
