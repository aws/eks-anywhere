package executables

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const defaultEksaImage = "public.ecr.aws/l0g8r8j6/eks-anywhere-cli-tools:v0.1.0-eks-a-v0.0.0-dev-build.589"

type ExecutableBuilder struct {
	useDocker  bool
	image      string
	mountDirs  []string
	workingDir string
	container  *dockerContainer
}

func (b *ExecutableBuilder) BuildKindExecutable(writer filewriter.FileWriter) *Kind {
	return NewKind(b.buildExecutable(kindPath), writer)
}

func (b *ExecutableBuilder) BuildClusterAwsAdmExecutable() *Clusterawsadm {
	return NewClusterawsadm(b.buildExecutable(clusterAwsAdminPath))
}

func (b *ExecutableBuilder) BuildClusterCtlExecutable(writer filewriter.FileWriter) *Clusterctl {
	return NewClusterctl(b.buildExecutable(clusterCtlPath), writer)
}

func (b *ExecutableBuilder) BuildKubectlExecutable() *Kubectl {
	return NewKubectl(b.buildExecutable(kubectlPath))
}

func (b *ExecutableBuilder) BuildGovcExecutable(writer filewriter.FileWriter) *Govc {
	return NewGovc(b.buildExecutable(govcPath), writer)
}

func (b *ExecutableBuilder) BuildCmkExecutable(writer filewriter.FileWriter, execConfig decoder.CloudStackExecConfig) *Cmk {
	return NewCmk(b.buildExecutable(cmkPath), writer, execConfig)
}

func (b *ExecutableBuilder) BuildTinkExecutable(tinkerbellCertUrl, tinkerbellGrpcAuthority string) *Tink {
	return NewTink(b.buildExecutable(tinkPath), tinkerbellCertUrl, tinkerbellGrpcAuthority)
}

func (b *ExecutableBuilder) BuildAwsCli() *AwsCli {
	return NewAwsCli(b.buildExecutable(awsCliPath))
}

func (b *ExecutableBuilder) BuildFluxExecutable() *Flux {
	return NewFlux(b.buildExecutable(fluxPath))
}

func (b *ExecutableBuilder) BuildTroubleshootExecutable() *Troubleshoot {
	return NewTroubleshoot(b.buildExecutable(troubleshootPath))
}

func (b *ExecutableBuilder) BuildHelmExecutable() *Helm {
	return NewHelm(b.buildExecutable(helmPath))
}

func (b *ExecutableBuilder) Close(ctx context.Context) *Troubleshoot {
	return NewTroubleshoot(b.buildExecutable(troubleshootPath))
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

func (b *ExecutableBuilder) buildExecutable(cli string) Executable {
	if !b.useDocker {
		return NewExecutable(cli)
	} else {
		return NewDockerExecutable(cli, b.container)
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

func NewExecutableBuilder(ctx context.Context, image string, mountDirs ...string) (*ExecutableBuilder, Closer, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("getting current directory: %v", err)
	}

	mountDirs = append(mountDirs, currentDir)

	useDocker := !checkMRToolsDisabled()
	e := &ExecutableBuilder{
		useDocker:  useDocker,
		image:      image,
		mountDirs:  mountDirs,
		workingDir: currentDir,
	}

	if useDocker {
		// We build, init and store the container in the builder so we reuse the same one for all the executables
		container := newDockerContainer(image, e.workingDir, e.mountDirs, BuildDockerExecutable())
		if err := container.init(ctx); err != nil {
			return nil, nil, err
		}
		e.container = container
	}

	return e, e.closer(), nil
}

func (e *ExecutableBuilder) closer() Closer {
	c := e.container

	return func(ctx context.Context) error {
		if c != nil {
			return c.close(ctx)
		}
		return nil
	}
}

func NewLocalExecutableBuilder() *ExecutableBuilder {
	return &ExecutableBuilder{
		useDocker: false,
		image:     "",
	}
}

func DefaultEksaImage() string {
	return defaultEksaImage
}

type Closer func(ctx context.Context) error

// Close implements interface types.Closer
func (c Closer) Close(ctx context.Context) error {
	return c(ctx)
}

// CheckErr just calls the closer and logs an error if present
// It's mostly a helper for defering the close in a oneliner without ignoring the error
func (c Closer) CheckErr(ctx context.Context) {
	if err := c(ctx); err != nil {
		logger.Error(err, "Failed closing container for executables")
	}
}
