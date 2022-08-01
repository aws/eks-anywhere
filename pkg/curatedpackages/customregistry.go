package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/executables"
)

type CustomRegistry struct {
	*executables.Helm
	registry string
}

func NewCustomRegistry(helm *executables.Helm, registry string) *CustomRegistry {
	return &CustomRegistry{
		Helm:     helm,
		registry: registry,
	}
}

func (cm *CustomRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s/%s", cm.registry, ImageRepositoryName), nil
}
