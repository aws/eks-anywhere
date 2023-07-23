package machinehealthcheck

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// GetControlPlaneMachineHealthCheck checks if machine health checks already exist on the cluster and returns it.
func GetControlPlaneMachineHealthCheck(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) (*clusterv1.MachineHealthCheck, error) {
	mhc := &clusterv1.MachineHealthCheck{}
	// we get only kcp machine health check as it had the same timeouts as worker machine health checks
	kcpMHCName := clusterapi.ControlPlaneMachineHealthCheckName(cluster)
	err := client.Get(ctx, types.NamespacedName{Name: kcpMHCName, Namespace: constants.EksaSystemNamespace}, mhc)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fetching machine health check %s", kcpMHCName)
	}

	return mhc, nil
}
