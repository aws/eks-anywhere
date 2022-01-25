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
			Err:         ValidateTaintsSupport(u.Opts.Spec),
			FeatureFlag: true,
		},
		{
			Name:        "validate node labels support",
			Remediation: "ensure NODE_LABELS_SUPPORT env variable is set",
			Err:         ValidateNodeLabelsSupport(u.Opts.Spec),
			FeatureFlag: true,
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

	var errs []string
	for _, validation := range createValidations {
		if validation.Err != nil {
			errs = append(errs, validation.Err.Error())
		} else if !validation.FeatureFlag {
			validation.LogPass()
		}
	}

	if len(errs) > 0 {
		return &validations.ValidationError{Errs: errs}
	}
	return nil
}
