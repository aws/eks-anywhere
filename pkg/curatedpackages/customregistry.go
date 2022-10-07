package curatedpackages

import (
	"context"
	"fmt"
)

// CustomRegistry aids in requesting bundles from OCI registries.
type CustomRegistry struct {
	registry string
}

func NewCustomRegistry(registry string) *CustomRegistry {
	return &CustomRegistry{
		registry: registry,
	}
}

var _ BundleRegistry = (*CustomRegistry)(nil)

// GetRegistryBaseRef implements BundleRegistry
func (cm *CustomRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s/%s", cm.registry, ImageRepositoryName), nil
}
