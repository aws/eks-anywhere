package handlerutil

import (
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func ChildObjectToClusters(log logr.Logger) handler.MapFunc {
	return func(o client.Object) []reconcile.Request {
		requests := []reconcile.Request{}
		for _, owner := range o.GetOwnerReferences() {
			if owner.Kind == anywherev1.ClusterKind {
				requests = append(requests, reconcileRequestForOwnerRef(o, owner))
			}
		}

		if len(requests) == 0 {
			log.V(6).Info("Object doesn't contain references to a Cluster", "kind", o.GetObjectKind(), "name", o.GetName())
		}

		return requests
	}
}

func reconcileRequestForOwnerRef(o client.Object, owner metav1.OwnerReference) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      owner.Name,
			Namespace: o.GetNamespace(),
		},
	}
}
