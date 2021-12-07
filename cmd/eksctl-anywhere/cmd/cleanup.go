package cmd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

func cleanup(ctx context.Context, deps *dependencies.Dependencies, commandErr *error) {
	close(ctx, deps)

	if commandErr == nil {
		deps.Writer.CleanUpTemp()
	}
}

func close(ctx context.Context, closer types.Closer) {
	if err := closer.Close(ctx); err != nil {
		logger.Error(err, "Closer failed", "closerType", fmt.Sprintf("%T", closer))
	}
}
