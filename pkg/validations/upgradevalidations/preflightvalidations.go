package upgradevalidations

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// PreflightValidations returns the validations required before upgrading a cluster.
func (u *UpgradeValidations) PreflightValidations(ctx context.Context) []validations.Validation {
	k := u.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           u.Opts.WorkloadCluster.Name,
		KubeconfigFile: u.Opts.ManagementCluster.KubeconfigFile,
	}
	upgradeValidations := []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate OS is compatible with registry mirror configuration",
				Remediation: "please use a valid OS for your registry mirror configuration",
				Err:         validations.ValidateOSForRegistryMirror(u.Opts.Spec, u.Opts.Provider),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate certificate for registry mirror",
				Remediation: fmt.Sprintf("provide a valid certificate for you registry endpoint using %s env var", anywherev1.RegistryMirrorCAKey),
				Err:         validations.ValidateCertForRegistryMirror(u.Opts.Spec, u.Opts.TLSValidator),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "control plane ready",
				Remediation: fmt.Sprintf("ensure control plane nodes and pods for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
				Err:         k.ValidateControlPlaneNodes(ctx, targetCluster, targetCluster.Name),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "worker nodes ready",
				Remediation: fmt.Sprintf("ensure machine deployments for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
				Err:         k.ValidateWorkerNodes(ctx, u.Opts.Spec.Cluster.Name, targetCluster.KubeconfigFile),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "nodes ready",
				Remediation: fmt.Sprintf("check the Status of the control plane and worker nodes in cluster %s and verify they are Ready", u.Opts.WorkloadCluster.Name),
				Err:         k.ValidateNodes(ctx, u.Opts.WorkloadCluster.KubeconfigFile),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "cluster CRDs ready",
				Remediation: "",
				Err:         k.ValidateClustersCRD(ctx, u.Opts.ManagementCluster),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "cluster object present on workload cluster",
				Remediation: fmt.Sprintf("ensure that the CAPI cluster object %s representing cluster %s is present", clusterv1.GroupVersion, u.Opts.WorkloadCluster.Name),
				Err:         ValidateClusterObjectExists(ctx, k, u.Opts.ManagementCluster),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "upgrade cluster kubernetes version increment",
				Remediation: "ensure that the cluster kubernetes version is incremented by one minor version exactly (e.g. 1.18 -> 1.19)",
				Err:         ValidateServerVersionSkew(ctx, u.Opts.Spec.Cluster.Spec.KubernetesVersion, u.Opts.WorkloadCluster, k),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate authentication for git provider",
				Remediation: fmt.Sprintf("ensure %s, %s env variable are set and valid", config.EksaGitPrivateKeyTokenEnv, config.EksaGitKnownHostsFileEnv),
				Err:         validations.ValidateAuthenticationForGitProvider(u.Opts.Spec, u.Opts.CliConfig),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate immutable fields",
				Remediation: "",
				Err:         ValidateImmutableFields(ctx, k, targetCluster, u.Opts.Spec, u.Opts.Provider),
			}
		},
	}

	if u.Opts.Spec.Cluster.IsManaged() {
		upgradeValidations = append(
			upgradeValidations,
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate management cluster bundle version compatibility",
					Remediation: fmt.Sprintf("upgrade management cluster %s before upgrading workload cluster %s", u.Opts.Spec.Cluster.ManagedBy(), u.Opts.WorkloadCluster.Name),
					Err:         validations.ValidateManagementClusterBundlesVersion(ctx, k, u.Opts.ManagementCluster, u.Opts.Spec),
				}
			})
	}

	if !u.Opts.SkippedValidations[PDB] {
		upgradeValidations = append(
			upgradeValidations,
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate pod disruption budgets",
					Remediation: "",
					Err:         ValidatePodDisruptionBudgets(ctx, k, u.Opts.WorkloadCluster),
				}
			})
	}
	return upgradeValidations
}
