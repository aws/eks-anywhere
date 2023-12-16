package framework

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

func buildKubectl(t T) *executables.Kubectl {
	ctx := context.Background()
	kubectl := executableBuilder(ctx, t).BuildKubectlExecutable()

	return kubectl
}

func buildLocalKubectl() *executables.Kubectl {
	return executables.NewLocalExecutablesBuilder().BuildKubectlExecutable()
}

func executableBuilder(ctx context.Context, t T) *executables.ExecutablesBuilder {
	executableBuilder, close, err := executables.InitInDockerExecutablesBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		t.Fatalf("Unable to initialize executable builder: %v", err)
	}
	t.Cleanup(func() {
		if err := close(ctx); err != nil {
			t.Fatal(err)
		}
	})

	return executableBuilder
}

func buildGovc(t T) *executables.Govc {
	ctx := context.Background()
	tmpWriter, err := filewriter.NewWriter("unique-ip")
	if err != nil {
		t.Fatalf("Error creating tmp writer")
	}
	govc := executableBuilder(ctx, t).BuildGovcExecutable(tmpWriter)
	t.Cleanup(func() {
		govc.Close(ctx)
	})

	return govc
}

func buildDocker(t T) *executables.Docker {
	return executables.BuildDockerExecutable()
}

func buildHelm(t T) helm.Client {
	ctx := context.Background()
	helm := executableBuilder(ctx, t).BuildHelmExecutable(helm.WithInsecure())

	return helm
}

func buildSSH(t T) *executables.SSH {
	return executables.NewLocalExecutablesBuilder().BuildSSHExecutable()
}

func buildCmk(t T) *executables.Cmk {
	ctx := context.Background()
	tmpWriter, err := filewriter.NewWriter("cmk")
	if err != nil {
		t.Fatalf("Error creating tmp writer")
	}

	execConfig, err := decoder.ParseCloudStackCredsFromEnv()
	if err != nil {
		t.Fatalf("parsing cloudstack credentials from environment: %v", err)
	}

	cmk, err := executableBuilder(ctx, t).BuildCmkExecutable(tmpWriter, execConfig)
	if err != nil {
		t.Fatalf("Error creating cmk client: %v", err)
	}
	t.Cleanup(func() {
		cmk.Close(ctx)
	})

	return cmk
}
