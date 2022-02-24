package upgradevalidations

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (u *UpgradeValidations) PreflightValidations(ctx context.Context) (err error) {
	k := u.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           u.Opts.WorkloadCluster.Name,
		KubeconfigFile: u.Opts.ManagementCluster.KubeconfigFile,
	}
	var upgradeValidations []validations.ValidationResult
	upgradeValidations = append(
		upgradeValidations,
		validations.ValidationResult{
			Name:        "validate kubernetes version 1.22 support",
			Remediation: fmt.Sprintf("ensure %v env variable is set", features.K8s122SupportEnvVar),
			Err:         validations.ValidateK8s122Support(u.Opts.Spec),
			Silent:      true,
		},
		validations.ValidationResult{
			Name:        "control plane ready",
			Remediation: fmt.Sprintf("ensure control plane nodes and pods for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateControlPlaneNodes(ctx, targetCluster, targetCluster.Name),
		},
		validations.ValidationResult{
			Name:        "worker nodes ready",
			Remediation: fmt.Sprintf("ensure machine deployments for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateWorkerNodes(ctx, u.Opts.Spec.Name, targetCluster.KubeconfigFile),
		},
		validations.ValidationResult{
			Name:        "nodes ready",
			Remediation: fmt.Sprintf("check the Status of the control plane and worker nodes in cluster %s and verify they are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateNodes(ctx, u.Opts.WorkloadCluster.KubeconfigFile),
		},
		validations.ValidationResult{
			Name:        "cluster CRDs ready",
			Remediation: "",
			Err:         k.ValidateClustersCRD(ctx, u.Opts.ManagementCluster),
		},
		validations.ValidationResult{
			Name:        "cluster object present on workload cluster",
			Remediation: fmt.Sprintf("ensure that the CAPI cluster object %s representing cluster %s is present", clusterv1.GroupVersion, u.Opts.WorkloadCluster.Name),
			Err:         ValidateClusterObjectExists(ctx, k, u.Opts.ManagementCluster),
		},
		validations.ValidationResult{
			Name:        "upgrade cluster kubernetes version increment",
			Remediation: "ensure that the cluster kubernetes version is incremented by one minor version exactly (e.g. 1.18 -> 1.19)",
			Err:         ValidateServerVersionSkew(ctx, u.Opts.Spec.Spec.KubernetesVersion, u.Opts.WorkloadCluster, k),
		},
		validations.ValidationResult{
			Name:        "validate immutable fields",
			Remediation: "",
			Err:         ValidateImmutableFields(ctx, k, targetCluster, u.Opts.Spec, u.Opts.Provider),
		},
	)

	return validations.RunPreflightValidations(upgradeValidations)
}
