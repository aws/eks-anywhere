package types

import "github.com/aws/eks-anywhere/release/api/v1alpha1"

type Cluster struct {
	Name           string
	KubeconfigFile string
}

// DeepCopy creates a new in-memory copy of c.
func (c *Cluster) DeepCopy() *Cluster {
	return &Cluster{
		Name:           c.Name,
		KubeconfigFile: c.KubeconfigFile,
	}
}

type InfrastructureBundle struct {
	FolderName string
	Manifests  []v1alpha1.Manifest
}
