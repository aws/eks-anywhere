package cluster

import (
	"context"
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, fetch)
	if err != nil {
		return nil, err
	}

	return BuildSpecFromBundles(cluster, bundles)
}

func GetBundlesForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*v1alpha1release.Bundles, error) {
	bundles, err := fetch(ctx, cluster.Name, cluster.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed fetching Bundles for cluster: %v", err)
	}

	return bundles, nil
}

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
	kubeConfigBytes, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	kc := &kubeConfigYAML{}
	kc.Clusters = []*kubeConfigCluster{}
	err = yaml.Unmarshal(kubeConfigBytes, &kc)
	if err != nil {
		return nil, fmt.Errorf("error parsing kubeconfig file: %v", err)
	}
	return &types.Cluster{
		Name:               kc.Clusters[0].Name,
		KubeconfigFile:     kubeconfig,
		ExistingManagement: true,
	}, nil
}
