package executables

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const defaultEksaImage = "public.ecr.aws/l0g8r8j6/eks-anywhere-cli-tools:v0.18.4-eks-a-v0.0.0-dev-build.8100"

type ExecutableBuilder interface {
	Init(ctx context.Context) (Closer, error)
	Build(binaryPath string) Executable
}

type ExecutablesBuilder struct {
	executableBuilder ExecutableBuilder
}

func NewExecutablesBuilder(executableBuilder ExecutableBuilder) *ExecutablesBuilder {
	return &ExecutablesBuilder{
		executableBuilder: executableBuilder,
	}
}

func (b *ExecutablesBuilder) BuildKindExecutable(writer filewriter.FileWriter) *Kind {
	return NewKind(b.executableBuilder.Build(kindPath), writer)
}

func (b *ExecutablesBuilder) BuildClusterAwsAdmExecutable() *Clusterawsadm {
	return NewClusterawsadm(b.executableBuilder.Build(clusterAwsAdminPath))
}

// BuildClusterCtlExecutable builds a new Clusterctl executable.
func (b *ExecutablesBuilder) BuildClusterCtlExecutable(writer filewriter.FileWriter, reader manifests.FileReader) *Clusterctl {
	return NewClusterctl(b.executableBuilder.Build(clusterCtlPath), writer, reader)
}

func (b *ExecutablesBuilder) BuildKubectlExecutable() *Kubectl {
	return NewKubectl(b.executableBuilder.Build(kubectlPath))
}

func (b *ExecutablesBuilder) BuildGovcExecutable(writer filewriter.FileWriter, opts ...GovcOpt) *Govc {
	return NewGovc(b.executableBuilder.Build(govcPath), writer, opts...)
}

// BuildCmkExecutable initializes a Cmk object and returns it.
func (b *ExecutablesBuilder) BuildCmkExecutable(writer filewriter.FileWriter, config *decoder.CloudStackExecConfig) (*Cmk, error) {
	return NewCmk(b.executableBuilder.Build(cmkPath), writer, config)
}

func (b *ExecutablesBuilder) BuildAwsCli() *AwsCli {
	return NewAwsCli(b.executableBuilder.Build(awsCliPath))
}

func (b *ExecutablesBuilder) BuildFluxExecutable() *Flux {
	return NewFlux(b.executableBuilder.Build(fluxPath))
}

func (b *ExecutablesBuilder) BuildTroubleshootExecutable() *Troubleshoot {
	return NewTroubleshoot(b.executableBuilder.Build(troubleshootPath))
}

// BuildHelmExecutable initializes a helm executable and returns it.
func (b *ExecutablesBuilder) BuildHelmExecutable(opts ...helm.Opt) *Helm {
	return NewHelm(b.executableBuilder.Build(helmPath), opts...)
}

// BuildHelm initializes a helm executable and returns it.
func (b *ExecutablesBuilder) BuildHelm(opts ...helm.Opt) helm.Client {
	return b.BuildHelmExecutable(opts...)
}

// BuildDockerExecutable initializes a docker executable and returns it.
func (b *ExecutablesBuilder) BuildDockerExecutable() *Docker {
	return NewDocker(b.executableBuilder.Build(dockerPath))
}

// BuildSSHExecutable initializes a SSH executable and returns it.
func (b *ExecutablesBuilder) BuildSSHExecutable() *SSH {
	return NewSSH(b.executableBuilder.Build(sshPath))
}

// Init initializes the executable builder and returns a Closer
// that needs to be called once the executables are not in used anymore
// The closer will cleanup and free all internal resources.
func (b *ExecutablesBuilder) Init(ctx context.Context) (Closer, error) {
	return b.executableBuilder.Init(ctx)
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

// RunExecutablesInDocker determines if binary executables should be ran
// from a docker container or native binaries from the host path
// It reads MR_TOOLS_DISABLE variable.
func ExecutablesInDocker() bool {
	if env, ok := os.LookupEnv("MR_TOOLS_DISABLE"); ok && strings.EqualFold(env, "true") {
		logger.Info("Warning: eks-a tools image disabled, using client's executables")
		return false
	}
	return true
}

// InitInDockerExecutablesBuilder builds and inits a default ExecutablesBuilder to run executables in a docker container
// that will make use of a long running docker container.
func InitInDockerExecutablesBuilder(ctx context.Context, image string, mountDirs ...string) (*ExecutablesBuilder, Closer, error) {
	b, err := NewInDockerExecutablesBuilder(BuildDockerExecutable(), image, mountDirs...)
	if err != nil {
		return nil, nil, err
	}

	closer, err := b.Init(ctx)
	if err != nil {
		return nil, nil, err
	}

	return b, closer, nil
}

// NewInDockerExecutablesBuilder builds an executables builder for docker.
func NewInDockerExecutablesBuilder(dockerClient DockerClient, image string, mountDirs ...string) (*ExecutablesBuilder, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %v", err)
	}
	mountDirs = append(mountDirs, currentDir)

	dockerContainer := newDockerContainer(image, currentDir, mountDirs, dockerClient)
	dockerExecutableBuilder := NewDockerExecutableBuilder(dockerContainer)

	return NewExecutablesBuilder(dockerExecutableBuilder), nil
}

func NewLocalExecutablesBuilder() *ExecutablesBuilder {
	return NewExecutablesBuilder(newLocalExecutableBuilder())
}

func DefaultEksaImage() string {
	return defaultEksaImage
}

type Closer func(ctx context.Context) error

// Close implements interface types.Closer.
func (c Closer) Close(ctx context.Context) error {
	return c(ctx)
}

// CheckErr just calls the closer and logs an error if present
// It's mostly a helper for defering the close in a oneliner without ignoring the error.
func (c Closer) CheckErr(ctx context.Context) {
	if err := c(ctx); err != nil {
		logger.Error(err, "Failed closing container for executables")
	}
}
