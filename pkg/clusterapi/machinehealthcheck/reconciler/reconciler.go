package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/machinehealthcheck"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
)

// Reconciler allows to reconcile machine health checks.
type Reconciler struct {
	client    client.Client
	defaulter anywhereCluster.MachineHealthCheckDefaulter
}

// New returns a new Reconciler.
func New(client client.Client, defaulter anywhereCluster.MachineHealthCheckDefaulter) *Reconciler {
	return &Reconciler{
		client:    client,
		defaulter: defaulter,
	}
}

// Reconcile installs machine health checks for a given cluster.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	log.Info("Configuring machine health checks for workload", "cluster", cluster.Name)
	clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return err
	}

	if clusterSpec.Cluster.Spec.MachineHealthCheck == nil {
		clusterSpec.Cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{}
		mhc, err := machinehealthcheck.GetControlPlaneMachineHealthCheck(ctx, r.client, clusterSpec.Cluster)
		if err != nil {
			return err
		}
		if mhc != nil {
			copyNodeStartupTimeoutToCluster(clusterSpec.Cluster, mhc)
			copyUnhealthyMachineTimeoutToCluster(clusterSpec.Cluster, mhc)
		}
	}

	clusterSpec, err = r.defaulter.MachineHealthCheckDefault(ctx, clusterSpec)
	if err != nil {
		return err
	}

	mhcObject := clusterapi.MachineHealthCheckObjects(clusterSpec.Cluster)

	err = serverside.ReconcileObjects(ctx, r.client, clientutil.ObjectsToClientObjects(mhcObject))
	if err != nil {
		return fmt.Errorf("applying machine health checks: %v", err)
	}

	return nil
}

func copyNodeStartupTimeoutToCluster(cluster *anywherev1.Cluster, capiMHC *clusterv1.MachineHealthCheck) {
	if capiMHC.Spec.NodeStartupTimeout != nil {
		cluster.Spec.MachineHealthCheck.NodeStartupTimeout = capiMHC.Spec.NodeStartupTimeout
	}
}

func copyUnhealthyMachineTimeoutToCluster(cluster *anywherev1.Cluster, capiMHC *clusterv1.MachineHealthCheck) {
	if len(capiMHC.Spec.UnhealthyConditions) > 0 {
		cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout = &capiMHC.Spec.UnhealthyConditions[0].Timeout
	}
}
