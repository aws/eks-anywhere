package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/executables"
)

type CustomRegistry struct {
	*executables.Helm
	registry string
	username string
	password string
}

func NewCustomRegistry(helm *executables.Helm, registry, username, password string) *CustomRegistry {
	return &CustomRegistry{
		Helm:     helm,
		registry: registry,
		username: username,
		password: password,
	}
}

func (cm *CustomRegistry) Login(ctx context.Context) error {
	err := cm.RegistryLogin(ctx, cm.registry, cm.username, cm.password)
	if err != nil {
		return fmt.Errorf("unable to login to registry %v", err)
	}
	return nil
}

func (cm *CustomRegistry) GetRegistryBaseRef(ctx context.Context) (string, error) {
	return fmt.Sprintf("%s/%s", cm.registry, RepositoryName), nil
}
