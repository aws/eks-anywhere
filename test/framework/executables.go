package framework

import (
	"context"
	"testing"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

func buildKubectl(t *testing.T) *executables.Kubectl {
	ctx := context.Background()
	return executableBuilder(t, ctx).BuildKubectlExecutable()
}

func executableBuilder(t *testing.T, ctx context.Context) *executables.ExecutableBuilder {
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		t.Fatalf("Unable initialize executable builder: %v", err)
	}

	return executableBuilder
}

func buildGovc(t *testing.T) *executables.Govc {
	ctx := context.Background()
	tmpWriter, err := filewriter.NewWriter("unique-ip")
	if err != nil {
		t.Fatalf("Error creating tmp writer")
	}
	return executableBuilder(t, ctx).BuildGovcExecutable(tmpWriter)
}
