package validations

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/signature"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ValidateExtendedK8sVersionSupport validates all the validations needed for the support of extended kubernetes support.
func ValidateExtendedK8sVersionSupport(_ context.Context, _ *anywherev1.Cluster, bundle *v1alpha1.Bundles, _ kubernetes.Client) error {
	// Validate EKS-A bundle has not been modified by verifying the signature in the bundle annotation
	if err := validateBundleSignature(bundle); err != nil {
		return fmt.Errorf("validating bundle signature: %w", err)
	}
	return nil
}

// validateBundleSignature validates bundles signature with the KMS public key.
func validateBundleSignature(bundle *v1alpha1.Bundles) error {
	valid, err := signature.ValidateSignature(bundle, constants.KMSPublicKey)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("signature on the bundle is invalid, error: %w", err)
	}
	return nil
}
