package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/workflow"
)

const CreateBootstrapClusterTaskName = "CreateBootstrapCluster"

type CreateBootstrapCluster struct{}

func (task *CreateBootstrapCluster) RunTask(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (task *CreateBootstrapCluster) GetName() workflow.TaskName {
	return CreateBootstrapClusterTaskName
}
