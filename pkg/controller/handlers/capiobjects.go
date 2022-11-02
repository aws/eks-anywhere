package handlers

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

// CAPIObjectToCluster returns a request handler that enqueues an EKS-A Cluster
// reconcile request for CAPI objects that contain the cluster name and namespace labels.
func CAPIObjectToCluster(log logr.Logger) handler.MapFunc {
	return func(o client.Object) []reconcile.Request {
		labels := o.GetLabels()
		clusterName, ok := labels[clusterapi.EKSAClusterLabelName]
		if !ok {
			// Object not managed by an eks-a Cluster, don't enqueue
			log.V(6).Info("Object not managed by an eks-a Cluster, ignoring", "type", fmt.Sprintf("%T", o), "name", o.GetName())
			return nil
		}

		clusterNamespace := labels[clusterapi.EKSAClusterLabelNamespace]
		if clusterNamespace == "" {
			log.Info("Object managed by an eks-a Cluster but missing cluster namespace", "type", fmt.Sprintf("%T", o), "name", o.GetName())
			return nil
		}

		log.Info("Enqueuing Cluster request coming from CAPI object", "type", fmt.Sprintf("%T", o), "name", o.GetName(), "cluster", clusterName)
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{
				Namespace: clusterNamespace,
				Name:      clusterName,
			},
		}}
	}
}
