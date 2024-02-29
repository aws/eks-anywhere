package clusters

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterValidator runs cluster level validations.
type ClusterValidator struct {
	client client.Client
}

// NewClusterValidator returns a validator that will run cluster level validations.
func NewClusterValidator(client client.Client) *ClusterValidator {
	return &ClusterValidator{
		client: client,
	}
}

// ValidateManagementClusterName checks if the management cluster specified in the workload cluster spec is valid.
func (v *ClusterValidator) ValidateManagementClusterName(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	mgmtCluster, err := FetchManagementEksaCluster(ctx, v.client, cluster)
	if err != nil {
		return err
	}
	if mgmtCluster.IsManaged() {
		err := fmt.Errorf("%s is not a valid management cluster", mgmtCluster.Name)
		return err
	}

	return nil
}
