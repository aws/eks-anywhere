package types

import "github.com/aws/eks-anywhere/release/api/v1alpha1"

type Cluster struct {
	Name               string
	KubeconfigFile     string
	ExistingManagement bool // true is the cluster has EKS Anywhere management components
}

type InfrastructureBundle struct {
	FolderName string
	Manifests  []v1alpha1.Manifest
}
