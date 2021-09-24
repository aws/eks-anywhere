package conformance

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/executables"
)

func RunTests(ctx context.Context, contextName string, args ...string) (string, error) {
	sonobuoy := executables.BuildSonobuoyExecutable()
	return sonobuoy.Run(ctx, contextName, args...)
}

func GetResults(ctx context.Context, contextName string, args ...string) (string, error) {
	sonobuoy := executables.BuildSonobuoyExecutable()
	return sonobuoy.GetResults(ctx, contextName, args...)
}
