package executables

import (
	"context"
	"fmt"
)

const awsCliPath = "aws"

type AwsCli struct {
	executable Executable
}

func NewAwsCli(executable Executable) *AwsCli {
	return &AwsCli{
		executable: executable,
	}
}

func (ac *AwsCli) CreateAccessKey(ctx context.Context, username string) (string, error) {
	stdOut, err := ac.executable.Execute(ctx, "iam", "create-access-key", "--user-name", username)
	if err != nil {
		return "", fmt.Errorf("error executing iam create-access-key: %v", err)
	}
	return stdOut.String(), nil
}
