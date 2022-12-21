package executables

import (
	"bytes"
	"context"
	"fmt"
)

const containerNamePrefix = "eksa_"

type linuxDockerExecutable struct {
	cli           string
	containerName string
}

// This currently returns a linuxDockerExecutable, but if we support other types of docker executables we can change
// the name of this constructor.
func NewDockerExecutable(cli string, containerName string) Executable {
	return &linuxDockerExecutable{
		cli:           cli,
		containerName: containerName,
	}
}

func (e *linuxDockerExecutable) Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).Run()
}

func (e *linuxDockerExecutable) ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).WithStdIn(in).Run()
}

func (e *linuxDockerExecutable) ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).WithEnvVars(envs).Run()
}

func (e *linuxDockerExecutable) Command(ctx context.Context, args ...string) *Command {
	return NewCommand(ctx, e, args...)
}

func (e *linuxDockerExecutable) Run(cmd *Command) (stdout bytes.Buffer, err error) {
	return execute(cmd.ctx, "docker", cmd.stdIn, cmd.envVars, e.buildCommand(cmd.envVars, e.cli, cmd.args...)...)
}

func (e *linuxDockerExecutable) buildCommand(envs map[string]string, cli string, args ...string) []string {
	var envVars []string
	for k, v := range envs {
		envVars = append(envVars, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	dockerCommands := []string{"exec", "-i"}
	dockerCommands = append(dockerCommands, envVars...)

	dockerCommands = append(dockerCommands, e.containerName, e.cli)
	dockerCommands = append(dockerCommands, args...)

	return dockerCommands
}
