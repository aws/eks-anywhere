package capi

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
)

type Installer interface {
	Install(context.Context, *types.Cluster) error
}

type Install struct{}

func (i Install) RunTask(context.Context) (context.Context, error) {
	return ctx, nil
}
