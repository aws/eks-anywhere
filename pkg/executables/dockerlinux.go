package executables

import (
	"bytes"
	"context"
	"fmt"
)

const containerNamePrefix = "eksa_"

type linuxDockerExecutable struct {
	*dockerContainer
	cli string
}

// This currently returns a linuxDockerExecutable, but if we support other types of docker executables we can change
// the name of this constructor
func NewDockerExecutable(cli string, container *dockerContainer) Executable {
	return &linuxDockerExecutable{
		cli:             cli,
		dockerContainer: container,
	}
}

func (e *linuxDockerExecutable) Execute(ctx context.Context, args ...string) (bytes.Buffer, error) {
	var stdout bytes.Buffer
	if command, err := e.buildCommand(map[string]string{}, e.cli, args...); err != nil {
		return stdout, err
	} else {
		return execute(ctx, "docker", nil, command...)
	}
}

func (e *linuxDockerExecutable) ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (bytes.Buffer, error) {
	var stdout bytes.Buffer
	if command, err := e.buildCommand(map[string]string{}, e.cli, args...); err != nil {
		return stdout, err
	} else {
		return execute(ctx, "docker", in, command...)
	}
}

func (e *linuxDockerExecutable) ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (bytes.Buffer, error) {
	var stdout bytes.Buffer
	if command, err := e.buildCommand(envs, e.cli, args...); err != nil {
		return stdout, err
	} else {
		return execute(ctx, "docker", nil, command...)
	}
}

func (e *linuxDockerExecutable) buildCommand(envs map[string]string, cli string, args ...string) ([]string, error) {
	var envVars []string
	for k, v := range envs {
		envVars = append(envVars, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	dockerCommands := []string{"exec", "-i"}
	dockerCommands = append(dockerCommands, envVars...)

	dockerCommands = append(dockerCommands, e.containerName, e.cli)
	dockerCommands = append(dockerCommands, args...)

	return dockerCommands, nil
}
