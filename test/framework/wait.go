package framework

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

func (e *ClusterE2ETest) WaitForControlPlaneReady() {
	e.T.Log("Waiting for control plane to be ready")
	err := retrier.New(5 * time.Minute).Retry(func() error {
		return e.KubectlClient.ValidateControlPlaneNodes(context.Background(), e.cluster(), e.ClusterName)
	})
	if err != nil {
		e.T.Fatal(err)
	}
}
func (e *ClusterE2ETest) WaitForWorkloadClusterControlPlaneReady(clusterName string) {
	ctx := context.Background()
	e.T.Logf("Waiting for control plane %s to be ready for cluster %s", clusterName, e.ClusterName)
	err := e.KubectlClient.WaitForControlPlaneReady(ctx, e.cluster(), "5m", clusterName)
	if err != nil {
		e.T.Fatal(err)
	}
}