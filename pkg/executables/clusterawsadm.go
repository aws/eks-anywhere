package executables

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const clusterAwsAdminPath = "clusterawsadm"

type Clusterawsadm struct {
	executable Executable
}

func NewClusterawsadm(executable Executable) *Clusterawsadm {
	return &Clusterawsadm{executable: executable}
}

func (c *Clusterawsadm) BootstrapIam(ctx context.Context, envs map[string]string, configFile string) error {
	_, err := c.executable.ExecuteWithEnv(ctx, envs, "bootstrap", "iam", "create-cloudformation-stack",
		"--config", configFile)
	if err != nil {
		return fmt.Errorf("error executing bootstrap iam: %v", err)
	}
	return err
}

func (c *Clusterawsadm) BootstrapCreds(ctx context.Context, envs map[string]string) (string, error) {
	stdOut, err := c.executable.ExecuteWithEnv(ctx, envs, "bootstrap", "credentials", "encode-as-profile")
	if err != nil {
		return "", fmt.Errorf("error executing bootstrap credentials: %v", err)
	}
	return stdOut.String(), nil
}

func (c *Clusterawsadm) ListAccessKeys(ctx context.Context, userName string) (string, error) {
	stdOut, err := c.executable.Execute(ctx, "aws", "iam", "list-access-keys", "--user-name", userName)
	if err != nil {
		return "", fmt.Errorf("error listing user keys: %v", err)
	}
	return stdOut.String(), nil
}

func (c *Clusterawsadm) DeleteCloudformationStack(ctx context.Context, envs map[string]string, fileName string) error {
	logger.V(1).Info("Deleting AWS user")
	_, err := c.executable.ExecuteWithEnv(ctx, envs, "bootstrap", "iam", "delete-cloudformation-stack", "--config", fileName)
	if err != nil {
		if strings.Contains(err.Error(), "status code: 400") {
			return nil
		} else {
			return fmt.Errorf("failed to delete user %v", err)
		}
	}
	return nil
}
