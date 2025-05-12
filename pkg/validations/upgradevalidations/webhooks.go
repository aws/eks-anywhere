package upgradevalidations

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// WebhookIssue represents a potential issue with a webhook that could interfere with upgrades
type WebhookIssue struct {
	Name        string
	Description string
	Severity    string // "High", "Medium", "Low"
}

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

	// If no custom webhooks are found, return nil
	if len(customValidatingWebhooks) == 0 && len(customMutatingWebhooks) == 0 {
		return nil
	}

	// Check for potential issues with custom webhooks
	var issues []WebhookIssue

	// Check validating webhooks
	for _, webhook := range customValidatingWebhooks {
		webhookIssues := checkValidatingWebhookForIssues(webhook.Name, webhook.Webhooks)
		issues = append(issues, webhookIssues...)
	}

	// Check mutating webhooks
	for _, webhook := range customMutatingWebhooks {
		webhookIssues := checkMutatingWebhookForIssues(webhook.Name, webhook.Webhooks)
		issues = append(issues, webhookIssues...)
	}

	// If any issues are found, return an error with details
	if len(issues) > 0 {
		return formatWebhookIssuesError(issues)
	}

	// If no specific issues are found but custom webhooks exist, return a general warning
	return fmt.Errorf("one or more custom webhooks were detected on the cluster that might interfere with the upgrade process. Use the --skip-validations=%s flag if you wish to skip the validations for custom webhooks and proceed with the upgrade operation", validations.CustomWebhook)
}

// checkValidatingWebhookForIssues checks a validating webhook for common issues that could interfere with upgrades
func checkValidatingWebhookForIssues(webhookName string, webhooks []admissionregistrationv1.ValidatingWebhook) []WebhookIssue {
	var issues []WebhookIssue

	for i, webhook := range webhooks {
		// Check for failure policy set to Fail
		if webhook.FailurePolicy != nil && *webhook.FailurePolicy == admissionregistrationv1.Fail {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: "Webhook has failurePolicy set to Fail, which could block API operations during upgrade if the webhook service is unavailable",
				Severity:    "High",
			})
		}

		// Check for short timeout
		if webhook.TimeoutSeconds != nil && *webhook.TimeoutSeconds < 10 {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: fmt.Sprintf("Webhook has a short timeout (%d seconds), which might not be sufficient during cluster load in an upgrade", *webhook.TimeoutSeconds),
				Severity:    "Medium",
			})
		}

		// Check for critical resource interception
		criticalResources := checkForCriticalResources(webhook.Rules)
		if len(criticalResources) > 0 {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: fmt.Sprintf("Webhook intercepts critical resources that are modified during upgrades: %s", strings.Join(criticalResources, ", ")),
				Severity:    "High",
			})
		}

		// Check for service reference without selector
		if webhook.ClientConfig.Service != nil && webhook.ClientConfig.Service.Name != "" && webhook.ClientConfig.Service.Namespace != "" {
			// We can't check if the service has selectors here since we don't have access to the service object
			// This would require additional API calls to check the service
			// For now, we'll just note that the webhook uses a service reference
			if webhook.ClientConfig.Service.Port == nil || *webhook.ClientConfig.Service.Port == 0 {
				issues = append(issues, WebhookIssue{
					Name:        fmt.Sprintf("%s[%d]", webhookName, i),
					Description: "Webhook service reference doesn't specify a port, which might cause connection issues",
					Severity:    "Medium",
				})
			}
		}
	}

	return issues
}

// checkMutatingWebhookForIssues checks a mutating webhook for common issues that could interfere with upgrades
func checkMutatingWebhookForIssues(webhookName string, webhooks []admissionregistrationv1.MutatingWebhook) []WebhookIssue {
	var issues []WebhookIssue

	for i, webhook := range webhooks {
		// Check for failure policy set to Fail
		if webhook.FailurePolicy != nil && *webhook.FailurePolicy == admissionregistrationv1.Fail {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: "Webhook has failurePolicy set to Fail, which could block API operations during upgrade if the webhook service is unavailable",
				Severity:    "High",
			})
		}

		// Check for short timeout
		if webhook.TimeoutSeconds != nil && *webhook.TimeoutSeconds < 10 {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: fmt.Sprintf("Webhook has a short timeout (%d seconds), which might not be sufficient during cluster load in an upgrade", *webhook.TimeoutSeconds),
				Severity:    "Medium",
			})
		}

		// Check for critical resource interception
		criticalResources := checkForCriticalResources(webhook.Rules)
		if len(criticalResources) > 0 {
			issues = append(issues, WebhookIssue{
				Name:        fmt.Sprintf("%s[%d]", webhookName, i),
				Description: fmt.Sprintf("Webhook intercepts critical resources that are modified during upgrades: %s", strings.Join(criticalResources, ", ")),
				Severity:    "High",
			})
		}

		// Check for service reference without selector
		if webhook.ClientConfig.Service != nil && webhook.ClientConfig.Service.Name != "" && webhook.ClientConfig.Service.Namespace != "" {
			// We can't check if the service has selectors here since we don't have access to the service object
			// This would require additional API calls to check the service
			// For now, we'll just note that the webhook uses a service reference
			if webhook.ClientConfig.Service.Port == nil || *webhook.ClientConfig.Service.Port == 0 {
				issues = append(issues, WebhookIssue{
					Name:        fmt.Sprintf("%s[%d]", webhookName, i),
					Description: "Webhook service reference doesn't specify a port, which might cause connection issues",
					Severity:    "Medium",
				})
			}
		}
	}

	return issues
}

// checkForCriticalResources checks if the webhook intercepts critical resources that are modified during upgrades
func checkForCriticalResources(rules []admissionregistrationv1.RuleWithOperations) []string {
	criticalResources := make(map[string]bool)

	criticalAPIGroups := []string{
		"apps",
		"", // core API group
		"extensions",
		"apiextensions.k8s.io",
		"admissionregistration.k8s.io",
		"certificates.k8s.io",
	}

	criticalResourceNames := []string{
		"deployments",
		"daemonsets",
		"statefulsets",
		"pods",
		"nodes",
		"services",
		"customresourcedefinitions",
		"validatingwebhookconfigurations",
		"mutatingwebhookconfigurations",
		"certificatesigningrequests",
	}

	for _, rule := range rules {
		// Check if the rule applies to any critical API groups
		for _, apiGroup := range rule.Rule.APIGroups {
			if apiGroup == "*" {
				// Webhook applies to all API groups, which includes critical ones
				for _, resource := range criticalResourceNames {
					criticalResources[resource] = true
				}
				continue
			}

			isCriticalGroup := false
			for _, criticalGroup := range criticalAPIGroups {
				if apiGroup == criticalGroup {
					isCriticalGroup = true
					break
				}
			}

			if !isCriticalGroup {
				continue
			}

			// Check if the rule applies to any critical resources
			for _, resource := range rule.Rule.Resources {
				if resource == "*" {
					// Webhook applies to all resources in this API group
					for _, criticalResource := range criticalResourceNames {
						criticalResources[criticalResource] = true
					}
					continue
				}

				for _, criticalResource := range criticalResourceNames {
					if resource == criticalResource {
						criticalResources[criticalResource] = true
						break
					}
				}
			}
		}
	}

	// Convert map keys to slice
	var result []string
	for resource := range criticalResources {
		result = append(result, resource)
	}

	return result
}

// formatWebhookIssuesError formats the webhook issues into a detailed error message
func formatWebhookIssuesError(issues []WebhookIssue) error {
	var highSeverityIssues, mediumSeverityIssues, lowSeverityIssues []string

	for _, issue := range issues {
		issueStr := fmt.Sprintf("- %s: %s", issue.Name, issue.Description)

		switch issue.Severity {
		case "High":
			highSeverityIssues = append(highSeverityIssues, issueStr)
		case "Medium":
			mediumSeverityIssues = append(mediumSeverityIssues, issueStr)
		case "Low":
			lowSeverityIssues = append(lowSeverityIssues, issueStr)
		}
	}

	var message strings.Builder
	message.WriteString("Custom webhooks were detected that might interfere with the upgrade process:\n\n")

	if len(highSeverityIssues) > 0 {
		message.WriteString("HIGH SEVERITY ISSUES:\n")
		message.WriteString(strings.Join(highSeverityIssues, "\n"))
		message.WriteString("\n\n")
	}

	if len(mediumSeverityIssues) > 0 {
		message.WriteString("MEDIUM SEVERITY ISSUES:\n")
		message.WriteString(strings.Join(mediumSeverityIssues, "\n"))
		message.WriteString("\n\n")
	}

	if len(lowSeverityIssues) > 0 {
		message.WriteString("LOW SEVERITY ISSUES:\n")
		message.WriteString(strings.Join(lowSeverityIssues, "\n"))
		message.WriteString("\n\n")
	}

	message.WriteString("Consider temporarily removing these webhooks during the upgrade process or use the --skip-validations=")
	message.WriteString(validations.CustomWebhook)
	message.WriteString(" flag to proceed with the upgrade operation.")

	return fmt.Errorf(message.String())
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
		"gatekeeper",                       // OPA Gatekeeper
		"istio-",                           // Istio service mesh
		"vault-",                           // HashiCorp Vault
		"kyverno-",                         // Kyverno policy engine
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
		"gatekeeper",                     // OPA Gatekeeper
		"istio-",                         // Istio service mesh
		"vault-",                         // HashiCorp Vault
		"kyverno-",                       // Kyverno policy engine
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
