package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
)

func createKubectl(ctx context.Context) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage(executables.DefaultEksaImage()).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
