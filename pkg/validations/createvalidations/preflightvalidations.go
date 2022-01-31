package createvalidations

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (u *CreateValidations) PreflightValidations(ctx context.Context) (err error) {
	k := u.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           u.Opts.WorkloadCluster.Name,
		KubeconfigFile: u.Opts.ManagementCluster.KubeconfigFile,
	}

	createValidations := []validations.ValidationResult{
		{
			Name:        "validate taints support",
			Remediation: "ensure TAINTS_SUPPORT env variable is set",
			Err:         validations.ValidateTaintsSupport(u.Opts.Spec),
			Silent:      true,
		},
		{
			Name:        "validate node labels support",
			Remediation: "ensure NODE_LABELS_SUPPORT env variable is set",
			Err:         validations.ValidateNodeLabelsSupport(u.Opts.Spec),
			Silent:      true,
		},
	}

	if u.Opts.Spec.IsManaged() {
		createValidations = append(
			createValidations,
			validations.ValidationResult{
				Name:        "validate cluster name",
				Remediation: "",
				Err:         ValidateClusterNameIsUnique(ctx, k, targetCluster, u.Opts.Spec.Name),
			},
			validations.ValidationResult{
				Name:        "validate gitops",
				Remediation: "",
				Err:         ValidateGitOps(ctx, k, u.Opts.ManagementCluster, u.Opts.Spec),
			},
			validations.ValidationResult{
				Name:        "validate identityproviders name",
				Remediation: "",
				Err:         ValidateIdentityProviderNameIsUnique(ctx, k, targetCluster, u.Opts.Spec),
			},
			validations.ValidationResult{
				Name:        "validate management cluster has eksa crds",
				Remediation: "",
				Err:         ValidateManagementCluster(ctx, k, targetCluster),
			},
		)
	}

	return validations.RunPreflightValidations(createValidations)
}
