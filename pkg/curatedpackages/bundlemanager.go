package curatedpackages

import (
	"context"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

type Manager interface {
	LatestBundle(ctx context.Context, baseRef string) (
		*packagesv1.PackageBundle, error)
}
