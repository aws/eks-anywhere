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
	executable Executable
}

func NewSonobuoy(executable Executable) *Sonobuoy {
	return &Sonobuoy{
		executable: executable,
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
	_, err := k.executable.Execute(ctx, executionArgs...)
	if err != nil {
		return "", fmt.Errorf("error executing sonobuoy: %v", err)
	}

	executionArgs = []string{
		"--context",
		contextName,
		"retrieve",
		"./results",
	}
	var output bytes.Buffer
	output, err = k.executable.Execute(ctx, executionArgs...)
	if err != nil {
		return "", fmt.Errorf("error executing sonobuoy retrieve: %v", err)
	}
	outputFile := strings.TrimSpace(output.String())
	logger.Info("Sonobuoy results file: " + outputFile)

	executionArgs = []string{
		"results",
		outputFile,
	}
	output, err = k.executable.Execute(ctx, executionArgs...)
	if err != nil {
		return "", fmt.Errorf("error executing sonobuoy results command: %v", err)
	}
	return output.String(), err
}
