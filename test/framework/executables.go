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

	return kubectl
}

func buildLocalKubectl() *executables.Kubectl {
	return executables.NewLocalExecutablesBuilder().BuildKubectlExecutable()
}

func executableBuilder(t *testing.T, ctx context.Context) *executables.ExecutablesBuilder {
	executableBuilder, close, err := executables.InitInDockerExecutablesBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		t.Fatalf("Unable initialize executable builder: %v", err)
	}
	t.Cleanup(func() {
		if err := close(ctx); err != nil {
			t.Fatal(err)
		}
	})

	return executableBuilder
}

func buildGovc(t *testing.T) *executables.Govc {
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

func buildHelm(t *testing.T) *executables.Helm {
	ctx := context.Background()
	helm := executableBuilder(t, ctx).BuildHelmExecutable(executables.WithInsecure())

	return helm
}
