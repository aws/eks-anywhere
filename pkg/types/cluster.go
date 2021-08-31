package types

import "github.com/aws/eks-anywhere/release/api/v1alpha1"

type Cluster struct {
	Name           string
	KubeconfigFile string
}

type InfrastructureBundle struct {
	FolderName string
	Manifests  []v1alpha1.Manifest
}
