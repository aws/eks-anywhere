package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/machinehealthcheck"
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
// nolint:gocyclo
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	log.Info("Installing machine health checks for workload", "cluster", cluster.Name)

	// clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	// if err != nil {
	// 	return err
	// }

	if cluster.Spec.MachineHealthCheck == nil {
		mhc, err := machinehealthcheck.GetMachineHealthChecks(ctx, r.client, cluster)
		if err != nil {
			return err
		}
		if mhc != nil {
			if mhc.Spec.NodeStartupTimeout != nil {
				cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{
					NodeStartupTimeout: mhc.Spec.NodeStartupTimeout.Duration.String(),
				}
			}

			if len(mhc.Spec.UnhealthyConditions) > 0 {
				cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{
					UnhealthyMachineTimeout: mhc.Spec.UnhealthyConditions[0].Timeout.Duration.String(),
				}
			}
		}
	}

	cluster, err := r.defaulter.MachineHealthCheckDefault(ctx, cluster)
	if err != nil {
		return err
	}

	unhealthyMachineTimeout, err := time.ParseDuration(cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout)
	if err != nil {
		return err
	}

	nodeStartupTimeout, err := time.ParseDuration(cluster.Spec.MachineHealthCheck.NodeStartupTimeout)
	if err != nil {
		return err
	}

	mhc := clusterapi.MachineHealthCheckObjects(cluster, unhealthyMachineTimeout, nodeStartupTimeout)

	err = serverside.ReconcileObjects(ctx, r.client, clientutil.ObjectsToClientObjects(mhc))
	if err != nil {
		return fmt.Errorf("applying machine health checks: %v", err)
	}

	return nil
}
