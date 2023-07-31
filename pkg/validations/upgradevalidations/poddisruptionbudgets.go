package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	policy "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// ValidatePodDisruptionBudgets returns an error if any pdbs are detected on a cluster.
func ValidatePodDisruptionBudgets(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster) error {
	podDisruptionBudgets := &policy.PodDisruptionBudgetList{}
	if err := k.List(ctx, cluster.KubeconfigFile, podDisruptionBudgets); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "listing cluster pod disruption budgets for upgrade")
		}
	}

	if len(podDisruptionBudgets.Items) != 0 {
		return fmt.Errorf("one or more pod disruption budgets were detected on the cluster. Use the --skip-validations=%s flag if you wish to skip the validations for pod disruption budgets and proceed with the upgrade operation", validations.PDB)
	}

	return nil
}
