package upgradevalidations

import (
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/pkg/validations"
)

func (u *UpgradeValidations) PreflightValidations(ctx context.Context) (err error) {
	kubeconfig := u.Opts.WorkloadCluster.KubeconfigFile
	clusterName := u.Opts.WorkloadCluster.Name
	k := u.Opts.Kubectl

	var upgradeValidations []validations.ValidationResult
	upgradeValidations = append(
		upgradeValidations,
		validations.ValidationResult{
			Name:        "control plane ready",
			Remediation: fmt.Sprintf("ensure control plane nodes and pods for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateControlPlaneNodes(ctx, clusterName, kubeconfig),
		},
		validations.ValidationResult{
			Name:        "worker nodes ready",
			Remediation: fmt.Sprintf("ensure machine deployments for cluster %s are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateWorkerNodes(ctx, clusterName, kubeconfig),
		},
		validations.ValidationResult{
			Name:        "nodes ready",
			Remediation: fmt.Sprintf("check the Status of the control plane and worker nodes in cluster %s and verify they are Ready", u.Opts.WorkloadCluster.Name),
			Err:         k.ValidateNodes(ctx, kubeconfig),
		},
		validations.ValidationResult{
			Name:        "cluster CRDs ready",
			Remediation: "",
			Err:         k.ValidateClustersCRD(ctx, u.Opts.WorkloadCluster),
		},
		validations.ValidationResult{
			Name:        "cluster object present on workload cluster",
			Remediation: fmt.Sprintf("ensure that the CAPI cluster object %s representing cluster %s is present", v1alpha3.GroupVersion, u.Opts.WorkloadCluster.Name),
			Err:         ValidateClusterObjectExists(ctx, k, u.Opts.WorkloadCluster),
		},
		validations.ValidationResult{
			Name:        "upgrade cluster kubernetes version increment",
			Remediation: "ensure that the cluster kubernetes version is incremented by one minor version exactly (e.g. 1.18 -> 1.19)",
			Err:         ValidateServerVersionSkew(ctx, u.Opts.Spec.Spec.KubernetesVersion, u.Opts.WorkloadCluster, k),
		},
		validations.ValidationResult{
			Name:        "validate immutable fields",
			Remediation: "",
			Err:         ValidateImmutableFields(ctx, k, u.Opts.WorkloadCluster, u.Opts.Spec, u.Opts.Provider),
		},
	)

	var errs []string
	for _, validation := range upgradeValidations {
		if validation.Err != nil {
			errs = append(errs, validation.Err.Error())
		} else {
			validation.LogPass()
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return nil
}

type ValidationError struct {
	Errs []string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d errors: %s", len(v.Errs), strings.Join(v.Errs[:], ","))
}
