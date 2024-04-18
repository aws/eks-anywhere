package e2e

import (
	"fmt"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
)

func (e *E2ESession) loginToPublicECR() error {
	e.logger.V(1).Info("Logging in to public ECR")

	command := "aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws"
	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("sign in to public ecr: %v", err)
	}

	e.logger.V(1).Info("Logged in to public ECR")

	return nil
}
