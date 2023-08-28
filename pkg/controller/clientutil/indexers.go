package clientutil

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterNameIndexer is an indexer for controller to list Cluster objects based on name.
func ClusterNameIndexer(obj client.Object) []string {
	cluster, ok := obj.(*anywherev1.Cluster)
	if !ok {
		panic(fmt.Errorf("indexer function for type %T's metadata.name field received"+
			" object of type %T, this should never happen", anywherev1.Cluster{}, obj))
	}
	return []string{cluster.Name}
}
