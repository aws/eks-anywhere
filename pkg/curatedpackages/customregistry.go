package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/helm"
)

type CustomRegistry struct {
	helm.Client
	registry string
}

// NewCustomRegistry returns a new CustomRegistry.
func NewCustomRegistry(helm helm.Client, registry string) *CustomRegistry {
	return &CustomRegistry{
		Client:   helm,
		registry: registry,
	}
}

func (cm *CustomRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s/%s", cm.registry, ImageRepositoryName), nil
}
