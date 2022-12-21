package artifacts

import (
	"context"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Noop struct{}

func (*Noop) Download(ctx context.Context, bundles *releasev1.Bundles) {}

func (*Noop) Push(ctx context.Context, bundles *releasev1.Bundles) {}
