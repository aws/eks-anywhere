package validations

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"

	"github.com/aws/eks-anywhere/pkg/constants"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

// ValidateAvailabilityZones checks each availability zones defined cloudstackdatacenterconfig in the cluster.Spec
// have corresponding cloudstackfailuredomains objects within the cluster.
func ValidateAvailabilityZones(ctx context.Context, vc clusterf.StateValidationConfig) error {
	c := vc.ManagementClusterClient
	for _, az := range vc.ClusterSpec.CloudStackDatacenter.Spec.AvailabilityZones {
		fdName := cloudstackv1.FailureDomainHashedMetaName(az.Name, vc.ClusterSpec.Cluster.Name)
		key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: fdName}

		failureDomain := &cloudstackv1.CloudStackFailureDomain{}
		if err := c.Get(context.Background(), key, failureDomain); err != nil {
			return fmt.Errorf("failed to find failure domain %s corresponding to availability zone %s: %v", az.Name, fdName, err)
		}

	}

	return nil
}
