package upgradevalidations

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (u *UpgradeValidations) PreflightValidations(ctx context.Context) (err error) {
	k := u.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           u.Opts.WorkloadCluster.Name,
		KubeconfigFile: u.Opts.ManagementCluster.KubeconfigFile,
	}
	upgradeValidations := []validations.ValidationResult{
		{
			Name:        "validate certificate for registry mirror",
			Remediation: fmt.Sprintf("provide a valid certificate for you registry endpoint using %s env var", anywherev1.RegistryMirrorCAKey),
			Err:         validations.ValidateCertForRegistryMirror(u.Opts.Spec, u.Opts.TlsValidator),
		},
		{
			Name:        "control plane ready",
			Remediation: fmt.Sprintf("ensure control plane nodes and pods for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateControlPlaneNodes(ctx, targetCluster, targetCluster.Name),
		},
		{
			Name:        "worker nodes ready",
			Remediation: fmt.Sprintf("ensure machine deployments for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateWorkerNodes(ctx, u.Opts.Spec.Cluster.Name, targetCluster.KubeconfigFile),
		},
		{
			Name:        "nodes ready",
			Remediation: fmt.Sprintf("check the Status of the control plane and worker nodes in cluster %s and verify they are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateNodes(ctx, u.Opts.WorkloadCluster.KubeconfigFile),
		},
		{
			Name:        "cluster CRDs ready",
			Remediation: "",
			Err:         k.ValidateClustersCRD(ctx, u.Opts.ManagementCluster),
		},
		{
			Name:        "cluster object present on workload cluster",
			Remediation: fmt.Sprintf("ensure that the CAPI cluster object %s representing cluster %s is present", clusterv1.GroupVersion, u.Opts.WorkloadCluster.Name),
			Err:         ValidateClusterObjectExists(ctx, k, u.Opts.ManagementCluster),
		},
		{
			Name:        "upgrade cluster kubernetes version increment",
			Remediation: "ensure that the cluster kubernetes version is incremented by one minor version exactly (e.g. 1.18 -> 1.19)",
			Err:         ValidateServerVersionSkew(ctx, u.Opts.Spec.Cluster.Spec.KubernetesVersion, u.Opts.WorkloadCluster, k),
		},
		{
			Name:        "validate immutable fields",
			Remediation: "",
			Err:         ValidateImmutableFields(ctx, k, targetCluster, u.Opts.Spec, u.Opts.Provider),
		},
	}

	return validations.RunPreflightValidations(upgradeValidations)
}
