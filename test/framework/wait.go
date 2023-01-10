package framework

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

func (e *ClusterE2ETest) WaitForControlPlaneReady() {
	e.T.Log("Waiting for control plane to be ready")
	err := retrier.New(5 * time.Minute).Retry(func() error {
		return e.KubectlClient.ValidateControlPlaneNodes(context.Background(), e.Cluster(), e.ClusterName)
	})
	if err != nil {
		e.T.Fatal(err)
	}
}
