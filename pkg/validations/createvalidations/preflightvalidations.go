package createvalidations

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
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
			Name:        "validate certificate for registry mirror",
			Remediation: fmt.Sprintf("provide a valid certificate for you registry endpoint using %s env var", anywherev1.RegistryMirrorCAKey),
			Err:         validations.ValidateCertForRegistryMirror(u.Opts.Spec, u.Opts.TlsValidator),
		},
		{
			Name:        "validate authentication for git provider",
			Remediation: fmt.Sprintf("ensure %s, %s env variable are set and valid", config.EksaGitPrivateKeyTokenEnv, config.EksaGitKnownHostsFileEnv),
			Err:         validations.ValidateAuthenticationForGitProvider(u.Opts.Spec, u.Opts.CliConfig),
		},
	}

	if u.Opts.Spec.Cluster.IsManaged() {
		createValidations = append(
			createValidations,
			validations.ValidationResult{
				Name:        "validate cluster name",
				Remediation: "",
				Err:         ValidateClusterNameIsUnique(ctx, k, targetCluster, u.Opts.Spec.Cluster.Name),
			},
			validations.ValidationResult{
				Name:        "validate gitops",
				Remediation: "",
				Err:         ValidateGitOps(ctx, k, u.Opts.ManagementCluster, u.Opts.Spec),
			},
			validations.ValidationResult{
				Name:        "validate identity providers' name",
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
