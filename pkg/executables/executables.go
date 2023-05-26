package executables

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const (
	redactMask = "*****"
)

var redactedEnvKeys = []string{
	constants.VSphereUsernameKey,
	constants.VSpherePasswordKey,
	constants.GovcUsernameKey,
	constants.GovcPasswordKey,
	decoder.CloudStackCloudConfigB64SecretKey,
	eksaGithubTokenEnv,
	githubTokenEnv,
	config.EksaAccessKeyIdEnv,
	config.EksaSecretAccessKeyEnv,
	config.AwsAccessKeyIdEnv,
	config.AwsSecretAccessKeyEnv,
	constants.SnowCredentialsKey,
	constants.SnowCertsKey,
	constants.NutanixUsernameKey,
	constants.NutanixPasswordKey,
	constants.RegistryUsername,
	constants.RegistryPassword,
}

type executable struct {
	cli string
}

type Executable interface {
	Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error)
	ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) // TODO: remove this from interface in favor of Command
	ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (stdout bytes.Buffer, err error)            // TODO: remove this from interface in favor of Command
	Command(ctx context.Context, args ...string) *Command
	Run(cmd *Command) (stdout bytes.Buffer, err error)
}

// this should only be called through the executables.builder.
func NewExecutable(cli string) Executable {
	return &executable{
		cli: cli,
	}
}

func (e *executable) Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).Run()
}

func (e *executable) ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).WithStdIn(in).Run()
}

func (e *executable) ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
	return e.Command(ctx, args...).WithEnvVars(envs).Run()
}

func (e *executable) Command(ctx context.Context, args ...string) *Command {
	return NewCommand(ctx, e, args...)
}

func (e *executable) Run(cmd *Command) (stdout bytes.Buffer, err error) {
	for k, v := range cmd.envVars {
		os.Setenv(k, v)
	}
	return execute(cmd.ctx, e.cli, cmd.stdIn, cmd.envVars, cmd.args...)
}

func (e *executable) Close(ctx context.Context) error {
	return nil
}

func RedactCreds(cmd string, envMap map[string]string) string {
	redactedEnvs := []string{}
	for _, redactedEnvKey := range redactedEnvKeys {
		if env, found := os.LookupEnv(redactedEnvKey); found && env != "" {
			redactedEnvs = append(redactedEnvs, env)
		} else if env, found := envMap[redactedEnvKey]; found && env != "" {
			redactedEnvs = append(redactedEnvs, env)
		}
	}

	for _, redactedEnv := range redactedEnvs {
		cmd = strings.ReplaceAll(cmd, redactedEnv, redactMask)
	}
	return cmd
}

func execute(ctx context.Context, cli string, in []byte, envVars map[string]string, args ...string) (stdout bytes.Buffer, err error) {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, cli, args...)
	logger.V(6).Info("Executing command", "cmd", RedactCreds(cmd.String(), envVars))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if len(in) != 0 {
		cmd.Stdin = bytes.NewReader(in)
	}

	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			if logger.MaxLogging() {
				logger.V(logger.MaxLogLevel).Info(cli, "stderr", stderr.String())
			}
			return stdout, errors.New(stderr.String())
		} else {
			if !logger.MaxLogging() {
				logger.V(8).Info(cli, "stdout", stdout.String())
				logger.V(8).Info(cli, "stderr", stderr.String())
			}
			return stdout, errors.New(fmt.Sprint(err))
		}
	}
	if !logger.MaxLogging() {
		logger.V(8).Info(cli, "stdout", stdout.String())
		logger.V(8).Info(cli, "stderr", stderr.String())
	}
	return stdout, nil
}
