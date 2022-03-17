package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateIdentityProviderNameIsUnique(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec) error {
	if len(spec.Cluster.Spec.IdentityProviderRefs) == 0 || spec.Cluster.IsSelfManaged() {
		logger.V(5).Info("skipping ValidateIdentityProviderNameIsUnique")
		return nil
	}

	var existingIR []string
	for _, ir := range spec.Cluster.Spec.IdentityProviderRefs {
		eIR, err := k.SearchIdentityProviderConfig(ctx, ir.Name, ir.Kind, cluster.KubeconfigFile, spec.Cluster.Namespace)
		if err != nil {
			return err
		}
		if len(eIR) > 0 {
			existingIR = append(existingIR, eIR[0].Name)
		}
	}

	if len(existingIR) > 0 {
		return fmt.Errorf("the following identityProviders already exists %s", existingIR)
	}
	return nil
}
