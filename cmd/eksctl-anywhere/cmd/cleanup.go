package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

func cleanup(deps *dependencies.Dependencies, commandErr *error) {
	if *commandErr == nil {
		deps.Writer.CleanUpTemp()
	}
}

func close(ctx context.Context, closer types.Closer) {
	if err := closer.Close(ctx); err != nil {
		logger.Error(err, "Closer failed", "closerType", fmt.Sprintf("%T", closer))
	}
}

func cleanupDirectory(directory string) {
	if _, err := os.Stat(directory); err == nil {
		os.RemoveAll(directory)
	}
}
