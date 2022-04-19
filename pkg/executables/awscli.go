package executables

import (
	"context"
	"fmt"
)

const awsCliPath = "aws"

type AwsCli struct {
	Executable
}

func NewAwsCli(executable Executable) *AwsCli {
	return &AwsCli{
		Executable: executable,
	}
}

func (ac *AwsCli) CreateAccessKey(ctx context.Context, username string) (string, error) {
	stdOut, err := ac.Execute(ctx, "iam", "create-access-key", "--user-name", username)
	if err != nil {
		return "", fmt.Errorf("executing iam create-access-key: %v", err)
	}
	return stdOut.String(), nil
}
