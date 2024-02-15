package upgradevalidations

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validation"
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
			return resultForRemediableValidation(
				"SSH Keys present",
				providers.ValidateSSHKeyPresentForUpgrade(ctx, u.Opts.Spec),
			)
		},
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
				Err:         ValidateServerVersionSkew(ctx, u.Opts.Spec.Cluster, u.Opts.WorkloadCluster, u.Opts.ManagementCluster, k),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "upgrade cluster worker node group kubernetes version increment",
				Remediation: "ensure that the cluster worker node group kubernetes version is incremented by one minor version exactly (e.g. 1.18 -> 1.19) and cluster level kubernetes version does not exceed worker node group version by two minor versions",
				Err:         ValidateWorkerServerVersionSkew(ctx, u.Opts.Spec.Cluster, u.Opts.WorkloadCluster, u.Opts.ManagementCluster, k),
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
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate cluster's eksaVersion matches EKS-Anywhere Version",
				Remediation: "ensure eksaVersion matches the EKS-Anywhere release or omit the value from the cluster config",
				Err:         validations.ValidateEksaVersion(ctx, u.Opts.CliVersion, u.Opts.Spec),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate kubernetes version 1.29 support",
				Remediation: fmt.Sprintf("ensure %v env variable is set", features.K8s129SupportEnvVar),
				Err:         validations.ValidateK8s129Support(u.Opts.Spec),
				Silent:      true,
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate eksa controller is not paused",
				Remediation: fmt.Sprintf("remove cluster controller reconciler pause annotation %s before upgrading the cluster %s", u.Opts.Spec.Cluster.PausedAnnotation(), targetCluster.Name),
				Err:         validations.ValidatePauseAnnotation(ctx, k, targetCluster, targetCluster.Name),
			}
		},
	}

	if u.Opts.Spec.Cluster.IsManaged() {
		upgradeValidations = append(
			upgradeValidations,
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate management cluster eksaVersion compatibility",
					Remediation: fmt.Sprintf("upgrade management cluster %s before upgrading workload cluster %s", u.Opts.Spec.Cluster.ManagedBy(), u.Opts.WorkloadCluster.Name),
					Err:         validations.ValidateManagementClusterEksaVersion(ctx, k, u.Opts.ManagementCluster, u.Opts.Spec),
				}
			},
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate eksa release components exist on management cluster",
					Remediation: fmt.Sprintf("ensure eksaVersion is in the correct format (vMajor.Minor.Patch) and matches one of the available releases on the management cluster: kubectl get eksareleases -n %s --kubeconfig %s", constants.EksaSystemNamespace, u.Opts.ManagementCluster.KubeconfigFile),
					Err:         validations.ValidateEksaReleaseExistOnManagement(ctx, u.Opts.KubeClient, u.Opts.Spec.Cluster),
				}
			},
		)
	}

	if !u.Opts.SkippedValidations[validations.PDB] {
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
	if !u.Opts.SkippedValidations[validations.EksaVersionSkew] {
		upgradeValidations = append(
			upgradeValidations,
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate eksaVersion skew is one minor version",
					Remediation: "ensure eksaVersion upgrades are sequential by minor version",
					Err:         validations.ValidateEksaVersionSkew(ctx, k, u.Opts.ManagementCluster, u.Opts.Spec),
				}
			})
	}
	return upgradeValidations
}

func resultForRemediableValidation(name string, err error) *validations.ValidationResult {
	r := &validations.ValidationResult{
		Name: name,
		Err:  err,
	}

	if r.Err == nil {
		return r
	}

	if validation.IsRemediable(r.Err) {
		r.Remediation = validation.Remediation(r.Err)
	}

	return r
}
