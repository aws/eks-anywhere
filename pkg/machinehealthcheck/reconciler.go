package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
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
// nolint:gocyclo
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	log.Info("Installing machine health checks for workload", "cluster", cluster.Name)

	clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return err
	}

	if clusterSpec.Cluster.Spec.MachineHealthCheck == nil {
		mhc, err := r.getMachineHealthChecks(ctx, log, cluster)
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

	// machineHealthCheckDefaulter := anywhereCluster.NewMachineHealthCheckDefaulter(constants.DefaultNodeStartupTimeout.String(), constants.DefaultUnhealthyMachineTimeout.String())
	clusterSpec, err = r.defaulter.MachineHealthCheckDefault(ctx, clusterSpec)
	if err != nil {
		return err
	}

	unhealthyMachineTimeout, err := time.ParseDuration(clusterSpec.Cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout)
	if err != nil {
		return err
	}

	nodeStartupTimeout, err := time.ParseDuration(clusterSpec.Cluster.Spec.MachineHealthCheck.NodeStartupTimeout)
	if err != nil {
		return err
	}

	mhc := clusterapi.MachineHealthCheckObjects(clusterSpec.Cluster, unhealthyMachineTimeout, nodeStartupTimeout)

	err = serverside.ReconcileObjects(ctx, r.client, clientutil.ObjectsToClientObjects(mhc))
	if err != nil {
		return fmt.Errorf("applying machine health checks: %v", err)
	}

	return nil
}

// getMachineHealthChecks checks if machine health checks already exist on the cluster and returns it.
func (r *Reconciler) getMachineHealthChecks(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (*clusterv1.MachineHealthCheck, error) {
	mhc := &clusterv1.MachineHealthCheck{}
	// we get only kcp machine health check as it had the same timeouts as worker machine health checks
	kcpMHCName := clusterapi.ControlPlaneMachineHealthCheckName(cluster)
	err := r.client.Get(ctx, types.NamespacedName{Name: kcpMHCName, Namespace: constants.EksaSystemNamespace}, mhc)
	if apierrors.IsNotFound(err) {
		log.Info("Machine health checks don't exist for", "cluster", cluster.Name)
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fetching machine health check %s", kcpMHCName)
	}

	return mhc, nil
}
