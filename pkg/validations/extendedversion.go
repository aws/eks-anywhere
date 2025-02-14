package validations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/signature"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// LicensePublicKey is the public key for verifying license token signature.
// this is injected at build time.
var LicensePublicKey string

// ValidateExtendedK8sVersionSupport validates all the validations needed for the support of extended kubernetes support.
func ValidateExtendedK8sVersionSupport(ctx context.Context, clusterSpec anywherev1.Cluster, bundle *v1alpha1.Bundles, k kubernetes.Client) error {
	// Validate EKS-A bundle has not been modified by verifying the signature in the bundle annotation
	if err := validateBundleSignature(bundle); err != nil {
		return fmt.Errorf("validating bundle signature: %w", err)
	}

	// Check whether the kubernetes version for the cluster is currently under extended support by comparing the endOfStandardSupport date from the bundle with the current date.
	isExtended, err := isExtendedSupport(clusterSpec.Spec.KubernetesVersion, bundle)
	if err != nil {
		return err
	}
	if isExtended {
		token, err := getLicense(clusterSpec.Spec.LicenseToken)
		if err != nil {
			return fmt.Errorf("getting licenseToken: %w", err)
		}
		if err = validateLicense(token); err != nil {
			return fmt.Errorf("validating licenseToken: %w", err)
		}
		if clusterSpec.IsManaged() {
			if err := validateLicenseKeyIsUnique(ctx, clusterSpec.Name, clusterSpec.Spec.LicenseToken, k); err != nil {
				return fmt.Errorf("validating licenseToken is unique for cluster %s: %w", clusterSpec.Name, err)
			}
		}
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
		return errors.New("signature on the bundle is invalid")
	}
	return nil
}

func isExtendedSupport(kubernetesVersion anywherev1.KubernetesVersion, bundle *v1alpha1.Bundles) (bool, error) {
	versionsBundle, err := cluster.GetVersionsBundle(kubernetesVersion, bundle)
	if err != nil {
		return false, fmt.Errorf("getting versions bundle for %s kubernetes version: %w", kubernetesVersion, err)
	}

	endOfStandardSupport, err := time.Parse("2006-01-02", versionsBundle.EndOfStandardSupport)
	if err != nil {
		return false, fmt.Errorf("parsing EndOfStandardSupport field format: %w", err)
	}

	return isPastDateThanToday(endOfStandardSupport), nil
}

func getLicense(licenseToken string) (*jwt.Token, error) {
	if licenseToken == "" {
		return nil, errors.New("licenseToken is required for extended kubernetes support")
	}
	token, err := signature.ParseLicense(licenseToken, LicensePublicKey)
	if err != nil {
		return nil, fmt.Errorf("parsing licenseToken: %w", err)
	}

	return token, nil
}

func validateLicense(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("could not parse the licenseToken claims")
	}
	endValidity, ok := claims["endValidity"].(string)
	if !ok {
		return errors.New("license validity field missing from the licenseToken, not a valid license")
	}

	validity, err := time.Parse(time.RFC3339, endValidity)
	if err != nil {
		return fmt.Errorf("parsing endValidity field from licenseToken: %w", err)
	}

	if isPastDateThanToday(validity) {
		return errors.New("license is expired, please renew the license")
	}

	return nil
}

func isPastDateThanToday(dateToCompare time.Time) bool {
	today := time.Now().Truncate(24 * time.Hour)
	return dateToCompare.Before(today)
}

func validateLicenseKeyIsUnique(ctx context.Context, clusterName string, licenseToken string, k kubernetes.Client) error {
	eksaClusters := &anywherev1.ClusterList{}
	err := k.List(ctx, eksaClusters, kubernetes.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing all EKS-A clusters: %w", err)
	}
	for _, eksaCluster := range eksaClusters.Items {
		if eksaCluster.Name != clusterName && eksaCluster.Spec.LicenseToken == licenseToken {
			return fmt.Errorf("license key %s is already in use by cluster %s", licenseToken, eksaCluster.Name)
		}
	}
	return nil
}
