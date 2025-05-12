package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// ValidateCustomWebhooks returns an error if any custom webhooks are detected on a cluster
// that might interfere with the upgrade process.
func ValidateCustomWebhooks(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster) error {
	// Check for ValidatingWebhookConfigurations
	validatingWebhooks := &admissionregistrationv1.ValidatingWebhookConfigurationList{}
	if err := k.List(ctx, cluster.KubeconfigFile, validatingWebhooks); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "listing cluster validating webhook configurations for upgrade")
		}
	}

	// Check for MutatingWebhookConfigurations
	mutatingWebhooks := &admissionregistrationv1.MutatingWebhookConfigurationList{}
	if err := k.List(ctx, cluster.KubeconfigFile, mutatingWebhooks); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "listing cluster mutating webhook configurations for upgrade")
		}
	}

	// Filter out system webhooks
	customValidatingWebhooks := filterSystemValidatingWebhooks(validatingWebhooks.Items)
	customMutatingWebhooks := filterSystemMutatingWebhooks(mutatingWebhooks.Items)

	// If any custom webhooks are found, return an error
	if len(customValidatingWebhooks) > 0 || len(customMutatingWebhooks) > 0 {
		return fmt.Errorf("one or more custom webhooks were detected on the cluster that might interfere with the upgrade process. Use the --skip-validations=%s flag if you wish to skip the validations for custom webhooks and proceed with the upgrade operation", validations.CustomWebhook)
	}

	return nil
}

// filterSystemValidatingWebhooks filters out known system validating webhooks.
func filterSystemValidatingWebhooks(webhooks []admissionregistrationv1.ValidatingWebhookConfiguration) []admissionregistrationv1.ValidatingWebhookConfiguration {
	systemWebhookPrefixes := []string{
		"validating-webhook-configuration", // EKS-A webhook
		"cert-manager-webhook",             // cert-manager
		"eks-anywhere-webhook",             // EKS-A webhook
		"capi-",                            // Cluster API webhooks
		"capv-",                            // Cluster API vSphere webhooks
		"capt-",                            // Cluster API Tinkerbell webhooks
		"capcs-",                           // Cluster API CloudStack webhooks
		"capn-",                            // Cluster API Nutanix webhooks
		"caps-",                            // Cluster API Snow webhooks
	}

	var customWebhooks []admissionregistrationv1.ValidatingWebhookConfiguration
	for _, webhook := range webhooks {
		if !hasPrefix(webhook.Name, systemWebhookPrefixes) {
			customWebhooks = append(customWebhooks, webhook)
		}
	}
	return customWebhooks
}

// filterSystemMutatingWebhooks filters out known system mutating webhooks.
func filterSystemMutatingWebhooks(webhooks []admissionregistrationv1.MutatingWebhookConfiguration) []admissionregistrationv1.MutatingWebhookConfiguration {
	systemWebhookPrefixes := []string{
		"mutating-webhook-configuration", // EKS-A webhook
		"cert-manager-webhook",           // cert-manager
		"eks-anywhere-webhook",           // EKS-A webhook
		"capi-",                          // Cluster API webhooks
		"capv-",                          // Cluster API vSphere webhooks
		"capt-",                          // Cluster API Tinkerbell webhooks
		"capcs-",                         // Cluster API CloudStack webhooks
		"capn-",                          // Cluster API Nutanix webhooks
		"caps-",                          // Cluster API Snow webhooks
	}

	var customWebhooks []admissionregistrationv1.MutatingWebhookConfiguration
	for _, webhook := range webhooks {
		if !hasPrefix(webhook.Name, systemWebhookPrefixes) {
			customWebhooks = append(customWebhooks, webhook)
		}
	}
	return customWebhooks
}

// hasPrefix checks if a string has any of the given prefixes.
func hasPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
