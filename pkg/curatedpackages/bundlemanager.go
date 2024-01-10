package curatedpackages

import (
	"context"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

type Manager interface {
	LatestBundle(ctx context.Context, baseRef, kubeMajor, kubeMinor, clusterName string) (*packagesv1.PackageBundle, error)
}
