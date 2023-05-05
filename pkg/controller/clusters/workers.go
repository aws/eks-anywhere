package clusters

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/collection"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
)

// Workers represents the CAPI spec for an eks-a cluster's workers.
type Workers struct {
	Groups []WorkerGroup

	// Other includes any other provider-specific objects that need to be reconciled
	// as part of the worker groups.
	Other []client.Object
}

// objects returns a list of API objects for a collection of worker groups.
func (w *Workers) objects() []client.Object {
	objs := make([]client.Object, 0, len(w.Groups)*3+len(w.Other))
	for _, g := range w.Groups {
		objs = append(objs, g.objects()...)
	}
	objs = append(objs, w.Other...)

	return objs
}

// WorkerGroup represents the CAPI spec for an eks-a worker group.
type WorkerGroup struct {
	KubeadmConfigTemplate   *kubeadmv1.KubeadmConfigTemplate
	MachineDeployment       *clusterv1.MachineDeployment
	ProviderMachineTemplate client.Object
}

func (g *WorkerGroup) objects() []client.Object {
	objs := []client.Object{g.KubeadmConfigTemplate, g.MachineDeployment}

	if !reflect.ValueOf(g.ProviderMachineTemplate).IsNil() {
		objs = append(objs, g.ProviderMachineTemplate)
	}

	return objs
}

// ToWorkers converts the generic clusterapi Workers definition to the concrete one defined
// here. It's just a helper for callers generating workers spec using the clusterapi package.
func ToWorkers[M clusterapi.Object[M]](capiWorkers *clusterapi.Workers[M]) *Workers {
	w := &Workers{
		Groups: make([]WorkerGroup, 0, len(capiWorkers.Groups)),
	}

	for _, g := range capiWorkers.Groups {
		w.Groups = append(w.Groups, WorkerGroup{
			MachineDeployment:       g.MachineDeployment,
			KubeadmConfigTemplate:   g.KubeadmConfigTemplate,
			ProviderMachineTemplate: g.ProviderMachineTemplate,
		})
	}

	return w
}

// ReconcileWorkersForEKSA orchestrates the worker node reconciliation logic for a particular EKS-A cluster.
// It takes care of applying all desired objects in the Workers spec and deleting the
// old MachineDeployments that are not in it.
func ReconcileWorkersForEKSA(ctx context.Context, log logr.Logger, c client.Client, cluster *anywherev1.Cluster, w *Workers) (controller.Result, error) {
	capiCluster, err := controller.GetCAPICluster(ctx, c, cluster)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "reconciling workers for EKS-A cluster")
	}

	if capiCluster == nil {
		// cluster doesn't exist, this might be transient, requeuing
		log.Info("CAPI cluster doesn't exist yet, this might be transient if the CP have just been created, requeueing")
		return controller.ResultWithRequeue(5 * time.Second), nil
	}

	return ReconcileWorkers(ctx, c, capiCluster, w)
}

// ReconcileWorkers orchestrates the worker node reconciliation logic.
// It takes care of applying all desired objects in the Workers spec and deleting the
// old MachineDeployments that are not in it.
func ReconcileWorkers(ctx context.Context, c client.Client, cluster *clusterv1.Cluster, w *Workers) (controller.Result, error) {
	if err := serverside.ReconcileObjects(ctx, c, w.objects()); err != nil {
		return controller.Result{}, errors.Wrap(err, "applying worker nodes CAPI objects")
	}

	machineDeployments := &clusterv1.MachineDeploymentList{}
	if err := c.List(ctx, machineDeployments,
		client.MatchingLabels{clusterv1.ClusterNameLabel: cluster.Name},
		client.InNamespace(cluster.Namespace)); err != nil {
		return controller.Result{}, errors.Wrap(err, "listing current machine deployments")
	}

	desiredMachineDeploymentNames := collection.MapSet(w.Groups, func(g WorkerGroup) string {
		return g.MachineDeployment.Name
	})

	var allErrs []error

	for _, m := range machineDeployments.Items {
		if !desiredMachineDeploymentNames.Contains(m.Name) {
			if err := c.Delete(ctx, &m); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	if len(allErrs) > 0 {
		aggregate := utilerrors.NewAggregate(allErrs)
		return controller.Result{}, errors.Wrap(aggregate, "deleting machine deployments during worker node reconciliation")
	}

	return controller.Result{}, nil
}
