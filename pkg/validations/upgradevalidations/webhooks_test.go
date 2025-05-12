package upgradevalidations_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidateCustomWebhooks(t *testing.T) {
	failurePolicy := admissionregistrationv1.Fail
	ignorePolicy := admissionregistrationv1.Ignore
	shortTimeout := int32(5)
	normalTimeout := int32(15)
	port := int32(443)

	tests := []struct {
		name                   string
		validatingWebhooks     []admissionregistrationv1.ValidatingWebhookConfiguration
		mutatingWebhooks       []admissionregistrationv1.MutatingWebhookConfiguration
		expectError            bool
		expectedErrorSubstring string
		expectHighSeverity     bool
		expectMediumSeverity   bool
	}{
		{
			name:               "No webhooks",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{},
			mutatingWebhooks:   []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:        false,
		},
		{
			name: "Only system validating webhooks",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "validating-webhook-configuration",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capi-validating-webhook-configuration",
					},
				},
			},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:      false,
		},
		{
			name:               "Only system mutating webhooks",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mutating-webhook-configuration",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capi-mutating-webhook-configuration",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Custom validating webhook with no issues",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &ignorePolicy,
							TimeoutSeconds: &normalTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{"apps"},
										APIVersions: []string{"v1"},
										Resources:   []string{"replicasets"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE"},
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
			expectHighSeverity:     false,
			expectMediumSeverity:   false,
		},
		{
			name: "Custom validating webhook with failure policy issue",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &failurePolicy,
							TimeoutSeconds: &normalTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{"apps"},
										APIVersions: []string{"v1"},
										Resources:   []string{"replicasets"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE"},
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "HIGH SEVERITY ISSUES",
			expectHighSeverity:     true,
			expectMediumSeverity:   false,
		},
		{
			name: "Custom validating webhook with timeout issue",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &ignorePolicy,
							TimeoutSeconds: &shortTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{"apps"},
										APIVersions: []string{"v1"},
										Resources:   []string{"replicasets"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE"},
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "MEDIUM SEVERITY ISSUES",
			expectHighSeverity:     false,
			expectMediumSeverity:   true,
		},
		{
			name: "Custom validating webhook with critical resources issue",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &ignorePolicy,
							TimeoutSeconds: &normalTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{"apps"},
										APIVersions: []string{"v1"},
										Resources:   []string{"deployments"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "HIGH SEVERITY ISSUES",
			expectHighSeverity:     true,
			expectMediumSeverity:   false,
		},
		{
			name: "Custom validating webhook with multiple issues",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &failurePolicy,
							TimeoutSeconds: &shortTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{"apps"},
										APIVersions: []string{"v1"},
										Resources:   []string{"deployments"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "HIGH SEVERITY ISSUES",
			expectHighSeverity:     true,
			expectMediumSeverity:   true,
		},
		{
			name:               "Custom mutating webhook with issues",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-mutating-webhook",
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &failurePolicy,
							TimeoutSeconds: &shortTimeout,
							Rules: []admissionregistrationv1.RuleWithOperations{
								{
									Rule: admissionregistrationv1.Rule{
										APIGroups:   []string{""},
										APIVersions: []string{"v1"},
										Resources:   []string{"pods"},
									},
									Operations: []admissionregistrationv1.OperationType{"CREATE"},
								},
							},
						},
					},
				},
			},
			expectError:            true,
			expectedErrorSubstring: "HIGH SEVERITY ISSUES",
			expectHighSeverity:     true,
			expectMediumSeverity:   true,
		},
		{
			name: "Custom validating webhook with service port issue",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &ignorePolicy,
							TimeoutSeconds: &normalTimeout,
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "webhook-service",
									Namespace: "default",
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "MEDIUM SEVERITY ISSUES",
			expectHighSeverity:     false,
			expectMediumSeverity:   true,
		},
		{
			name: "Custom validating webhook with service port specified",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							Name:           "webhook1.example.com",
							FailurePolicy:  &ignorePolicy,
							TimeoutSeconds: &normalTimeout,
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "webhook-service",
									Namespace: "default",
									Port:      &port,
								},
							},
						},
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
			expectHighSeverity:     false,
			expectMediumSeverity:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			k := mocks.NewMockKubectlClient(ctrl)
			cluster := &types.Cluster{
				Name:           "test-cluster",
				KubeconfigFile: "kubeconfig",
			}

			// Setup mock for validating webhooks
			k.EXPECT().
				List(ctx, cluster.KubeconfigFile, gomock.AssignableToTypeOf(&admissionregistrationv1.ValidatingWebhookConfigurationList{})).
				DoAndReturn(func(_ context.Context, _ string, obj runtime.Object) error {
					webhookList := obj.(*admissionregistrationv1.ValidatingWebhookConfigurationList)
					webhookList.Items = tt.validatingWebhooks
					return nil
				})

			// Setup mock for mutating webhooks
			k.EXPECT().
				List(ctx, cluster.KubeconfigFile, gomock.AssignableToTypeOf(&admissionregistrationv1.MutatingWebhookConfigurationList{})).
				DoAndReturn(func(_ context.Context, _ string, obj runtime.Object) error {
					webhookList := obj.(*admissionregistrationv1.MutatingWebhookConfigurationList)
					webhookList.Items = tt.mutatingWebhooks
					return nil
				})

			err := upgradevalidations.ValidateCustomWebhooks(ctx, k, cluster)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorSubstring)
				assert.Contains(t, err.Error(), validations.CustomWebhook)

				if tt.expectHighSeverity {
					assert.Contains(t, err.Error(), "HIGH SEVERITY ISSUES")
				}

				if tt.expectMediumSeverity {
					assert.Contains(t, err.Error(), "MEDIUM SEVERITY ISSUES")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
