package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

type Reconciler struct {
	client        client.Client
	cniReconciler CNIReconciler
}

func New(client client.Client, cniReconciler CNIReconciler) *Reconciler {
	return &Reconciler{
		client:        client,
		cniReconciler: cniReconciler,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, c *anywherev1.Cluster) (controller.Result, error) {
	specWithBundles, err := cluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), c)
	if err != nil {
		return controller.Result{}, err
	}

	for _, machineConfig := range specWithBundles.SnowMachineConfigs {
		if !machineConfig.Status.SpecValid {
			failureMessage := fmt.Sprintf("SnowMachineConfig %s is invalid", machineConfig.Name)
			if machineConfig.Status.FailureMessage != nil {
				failureMessage += ": " + *machineConfig.Status.FailureMessage
			}

			log.Error(nil, failureMessage)
			c.Status.FailureMessage = &failureMessage
			return controller.Result{}, nil
		}
	}

	return controller.Result{}, nil
}
