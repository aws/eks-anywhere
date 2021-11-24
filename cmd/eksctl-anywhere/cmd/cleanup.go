package cmd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

func close(ctx context.Context, closer types.Closer) {
	if err := closer.Close(ctx); err != nil {
		logger.Error(err, "Closer failed", "closerType", fmt.Sprintf("%T", closer))
	}
}
