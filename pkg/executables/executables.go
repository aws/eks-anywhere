package executables

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	redactMask = "*****"
)

var redactedEnvKeys = []string{vSphereUsernameKey, vSpherePasswordKey}

type executable struct {
	cli string
}

type Executable interface {
	Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error)
	ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error)
	ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (stdout bytes.Buffer, err error)
}

// this should only be called through the executables.builder
func NewExecutable(cli string) Executable {
	return &executable{
		cli: cli,
	}
}

func (e *executable) Execute(ctx context.Context, args ...string) (bytes.Buffer, error) {
	return execute(ctx, e.cli, nil, args...)
}

func (e *executable) ExecuteWithStdin(ctx context.Context, in []byte, args ...string) (bytes.Buffer, error) {
	return execute(ctx, e.cli, in, args...)
}

func (e *executable) ExecuteWithEnv(ctx context.Context, envs map[string]string, args ...string) (stdout bytes.Buffer, err error) {
	for k, v := range envs {
		os.Setenv(k, v)
	}
	return e.Execute(ctx, args...)
}

func (e *executable) Close(ctx context.Context) error {
	return nil
}

func redactCreds(cmd string) string {
	redactedEnvs := []string{}
	for _, redactedEnvKey := range redactedEnvKeys {
		if env, found := os.LookupEnv(redactedEnvKey); found {
			redactedEnvs = append(redactedEnvs, env)
		}
	}

	for _, redactedEnv := range redactedEnvs {
		cmd = strings.ReplaceAll(cmd, redactedEnv, redactMask)
	}
	return cmd
}

func execute(ctx context.Context, cli string, in []byte, args ...string) (bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, cli, args...)
	logger.V(6).Info("Executing command", "cmd", redactCreds(cmd.String()))
	cmd.Stdout = &stdout
	if logger.MaxLogging() {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = &stderr
	}
	if len(in) != 0 {
		cmd.Stdin = bytes.NewReader(in)
	}

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
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
