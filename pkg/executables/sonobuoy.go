package executables

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const sonobuoyPath = "./sonobuoy"

type Sonobuoy struct {
	Executable
}

func NewSonobuoy(executable Executable) *Sonobuoy {
	return &Sonobuoy{
		Executable: executable,
	}
}

func (k *Sonobuoy) Run(ctx context.Context, contextName string, args ...string) (string, error) {
	logger.Info("Starting sonobuoy tests")
	executionArgs := []string{
		"--context",
		contextName,
		"run",
		"--mode=certified-conformance",
		"--wait",
	}
	executionArgs = append(executionArgs, args...)
	output, err := k.Execute(ctx, executionArgs...)
	command := strings.Join(executionArgs, " ") + "\n"
	if err != nil {
		return command, fmt.Errorf("executing sonobuoy: %v", err)
	}
	return command + output.String(), err
}

func (k *Sonobuoy) GetResults(ctx context.Context, contextName string, args ...string) (string, error) {
	executionArgs := []string{
		"--context",
		contextName,
		"retrieve",
		"./results",
	}
	var output bytes.Buffer
	output, err := k.Execute(ctx, executionArgs...)
	if err != nil {
		return "", fmt.Errorf("executing sonobuoy retrieve: %v", err)
	}
	outputFile := strings.TrimSpace(output.String())
	logger.Info("Sonobuoy results file: " + outputFile)

	executionArgs = []string{
		"results",
		outputFile,
	}
	output, err = k.Execute(ctx, executionArgs...)
	command := strings.Join(executionArgs, " ") + "\n"
	if err != nil {
		return command, fmt.Errorf("executing sonobuoy results command: %v", err)
	}
	return command + output.String(), err
}
