package framework

import (
	"context"
	"testing"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

func buildKubectl(t *testing.T) *executables.Kubectl {
	ctx := context.Background()
	kubectl := executableBuilder(t, ctx).BuildKubectlExecutable()
	t.Cleanup(func() {
		kubectl.Close(ctx)
	})

	return kubectl
}

func buildLocalKubectl() *executables.Kubectl {
	return executables.NewLocalExecutableBuilder().BuildKubectlExecutable()
}

func executableBuilder(t *testing.T, ctx context.Context) *executables.ExecutableBuilder {
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		t.Fatalf("Unable initialize executable builder: %v", err)
	}

	return executableBuilder
}

func BuildGovc(t *testing.T) *executables.Govc {
	ctx := context.Background()
	tmpWriter, err := filewriter.NewWriter("unique-ip")
	if err != nil {
		t.Fatalf("Error creating tmp writer")
	}
	govc := executableBuilder(t, ctx).BuildGovcExecutable(tmpWriter)
	t.Cleanup(func() {
		govc.Close(ctx)
	})

	return govc
}

func buildDocker(t *testing.T) *executables.Docker {
	return executables.BuildDockerExecutable()
}
