package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (u *CreateValidations) PreflightValidations(ctx context.Context) (err error) {
	k := u.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           u.Opts.WorkloadCluster.Name,
		KubeconfigFile: u.Opts.ManagementCluster.KubeconfigFile,
	}
	var createValidations []validations.ValidationResult

	fmt.Printf("create validations u.Opts.Spec.IsManaged() %v\n", u.Opts.Spec.IsManaged())

	if u.Opts.Spec.IsManaged() {
		createValidations = append(
			createValidations,
			validations.ValidationResult{
				Name:        "validate cluster name",
				Remediation: "",
				Err:         ValidateClusterObjectExists(ctx, k, targetCluster, u.Opts.Spec.Name),
			},
		)
	}

	var errs []string
	for _, validation := range createValidations {
		if validation.Err != nil {
			errs = append(errs, validation.Err.Error())
		} else {
			validation.LogPass()
		}
	}

	if len(errs) > 0 {
		return &validations.ValidationError{Errs: errs}
	}
	return nil
}
