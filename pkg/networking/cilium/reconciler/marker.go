package reconciler

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

// EKSACiliumInstalledAnnotation indicates a cluster has previously been observed to have
// EKS-A Cilium installed irrespective of whether its still installed.
const EKSACiliumInstalledAnnotation = "anywhere.eks.amazonaws.com/eksa-cilium"

// ciliumWasInstalled checks cluster for the EKSACiliumInstalledAnnotation.
func ciliumWasInstalled(ctx context.Context, cluster *v1alpha1.Cluster) bool {
	if cluster.Annotations == nil {
		return false
	}
	_, ok := cluster.Annotations[EKSACiliumInstalledAnnotation]
	return ok
}

// markCiliumInstalled populates the EKSACiliumInstalledAnnotation on cluster. It may trigger
// anothe reconciliation event.
func markCiliumInstalled(ctx context.Context, cluster *v1alpha1.Cluster) {
	clientutil.AddAnnotation(cluster, EKSACiliumInstalledAnnotation, "")
}
