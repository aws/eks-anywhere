package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/helm"
)

type CustomRegistry struct {
	helm.ExecuteableClient
	registry string
}

// NewCustomRegistry returns a new CustomRegistry.
func NewCustomRegistry(helm helm.ExecuteableClient, registry string) *CustomRegistry {
	return &CustomRegistry{
		ExecuteableClient: helm,
		registry:          registry,
	}
}

func (cm *CustomRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s/%s", cm.registry, ImageRepositoryName), nil
}
