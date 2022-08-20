package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

type Reconciler struct {
	client client.Client
}

func New(client client.Client) *Reconciler {
	return &Reconciler{
		client: client,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	specWithBundles, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
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
			cluster.Status.FailureMessage = &failureMessage
			return controller.Result{}, nil
		}
	}

	return controller.Result{}, nil
}
