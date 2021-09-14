package executables

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const defaultEksaImage = "public.ecr.aws/l0g8r8j6/eks-anywhere-cli-tools:v1-21-4-ed8ad899e40f0a0e625b7a49d7d90e6077f97a28"

type ExecutableBuilder struct {
	useDocker bool
	image     string
	mountDir  string
}

func (b *ExecutableBuilder) BuildKindExecutable(writer filewriter.FileWriter) *Kind {
	return NewKind(buildExecutable(kindPath, b.useDocker, b.image, b.mountDir), writer)
}

func (b *ExecutableBuilder) BuildClusterAwsAdmExecutable() *Clusterawsadm {
	return NewClusterawsadm(buildExecutable(clusterAwsAdminPath, b.useDocker, b.image, b.mountDir))
}

func (b *ExecutableBuilder) BuildClusterCtlExecutable(writer filewriter.FileWriter) *Clusterctl {
	return NewClusterctl(buildExecutable(clusterCtlPath, b.useDocker, b.image, b.mountDir), writer)
}

func (b *ExecutableBuilder) BuildKubectlExecutable() *Kubectl {
	return NewKubectl(buildExecutable(kubectlPath, b.useDocker, b.image, b.mountDir))
}

func (b *ExecutableBuilder) BuildGovcExecutable(writer filewriter.FileWriter) *Govc {
	return NewGovc(buildExecutable(govcPath, b.useDocker, b.image, b.mountDir), writer)
}

func (b *ExecutableBuilder) BuildAwsCli() *AwsCli {
	return NewAwsCli(buildExecutable(awsCliPath, b.useDocker, b.image, b.mountDir))
}

func (b *ExecutableBuilder) BuildFluxExecutable() *Flux {
	return NewFlux(buildExecutable(fluxPath, b.useDocker, b.image, b.mountDir))
}

func (b *ExecutableBuilder) BuildTroubleshootExecutable() *Troubleshoot {
	return NewTroubleshoot(buildExecutable(troulbeshootPath, b.useDocker, b.image, b.mountDir))
}

func BuildSonobuoyExecutable() *Sonobuoy {
	return NewSonobuoy(&executable{
		cli: sonobuoyPath,
	})
}

func BuildDockerExecutable() *Docker {
	return NewDocker(&executable{
		cli: dockerPath,
	})
}

func buildExecutable(cli string, useDocker bool, image string, mountDir string) Executable {
	if !useDocker {
		return NewExecutable(cli)
	} else {
		return NewDockerExecutable(cli, image, mountDir)
	}
}

// this is suppose to be only called by executables.builder
func checkMRToolsDisabled() bool {
	if env, ok := os.LookupEnv("MR_TOOLS_DISABLE"); ok && strings.EqualFold(env, "true") {
		logger.Info("Warning: eks-a tools image disabled, using client's executables")
		return true
	}
	return false
}

func NewExecutableBuilder(ctx context.Context, image string) (*ExecutableBuilder, error) {
	useDocker := !checkMRToolsDisabled()
	if useDocker {
		if err := setupDockerDependencies(ctx, image); err != nil {
			return nil, err
		}
	}
	mountDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}
	return &ExecutableBuilder{
		useDocker: useDocker,
		image:     image,
		mountDir:  mountDir,
	}, nil
}

func setupDockerDependencies(ctx context.Context, image string) error {
	if err := BuildDockerExecutable().SetUpCLITools(ctx, image); err != nil {
		return fmt.Errorf("failed to setup eks-a dependencies: %v", err)
	}
	return nil
}

func DefaultEksaImage() string {
	return defaultEksaImage
}
