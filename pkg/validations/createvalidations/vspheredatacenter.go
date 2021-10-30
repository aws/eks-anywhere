package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateDatacenterNameIsUnique(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec) error {
	if spec.IsSelfManaged() {
		logger.V(6).Info("skipping ValidateDatacenterNameIsUnique")
		return nil
	}

	existingDatacenter, err := k.SearchVsphereDatacenterConfig(ctx, spec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, spec.Namespace)
	if err != nil {
		return err
	}
	if len(existingDatacenter) > 0 {
		return fmt.Errorf("VSphereDatacenter %s already exists", spec.Spec.DatacenterRef.Name)
	}
	return nil
}
