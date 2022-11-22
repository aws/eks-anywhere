package clusters

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// CleanupStatusAfterValidate removes errors from the cluster status. Intended to be used as a reconciler phase
// after all validation phases have been executed.
func CleanupStatusAfterValidate(_ context.Context, _ logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	spec.Cluster.Status.FailureMessage = nil
	return controller.Result{}, nil
}
