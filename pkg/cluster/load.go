package cluster

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/types"
)

type kubeConfigCluster struct {
	Name string `json:"name"`
}

type kubeConfigYAML struct {
	Clusters []*kubeConfigCluster `json:"clusters"`
}

func LoadManagement(kubeconfig string) (*types.Cluster, error) {
	if kubeconfig == "" {
		return nil, nil
	}
	kubeConfigBytes, err := os.ReadFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	kc := &kubeConfigYAML{}
	kc.Clusters = []*kubeConfigCluster{}
	err = yaml.Unmarshal(kubeConfigBytes, &kc)
	if err != nil {
		return nil, fmt.Errorf("parsing kubeconfig file: %v", err)
	}

	if len(kc.Clusters) < 1 || len(kc.Clusters[0].Name) == 0 {
		return nil, fmt.Errorf("invalid kubeconfig file: %v", kubeconfig)
	}

	return &types.Cluster{
		Name:               kc.Clusters[0].Name,
		KubeconfigFile:     kubeconfig,
	}, nil
}
