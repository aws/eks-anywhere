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
	tests := []struct {
		name                   string
		validatingWebhooks     []admissionregistrationv1.ValidatingWebhookConfiguration
		mutatingWebhooks       []admissionregistrationv1.MutatingWebhookConfiguration
		expectError            bool
		expectedErrorSubstring string
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
			name: "Custom validating webhook",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
				},
			},
			mutatingWebhooks:       []admissionregistrationv1.MutatingWebhookConfiguration{},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
		},
		{
			name:               "Custom mutating webhook",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-mutating-webhook",
					},
				},
			},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
		},
		{
			name: "Both custom validating and mutating webhooks",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
				},
			},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-mutating-webhook",
					},
				},
			},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
		},
		{
			name: "Mix of system and custom webhooks",
			validatingWebhooks: []admissionregistrationv1.ValidatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "validating-webhook-configuration",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom-validating-webhook",
					},
				},
			},
			mutatingWebhooks: []admissionregistrationv1.MutatingWebhookConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mutating-webhook-configuration",
					},
				},
			},
			expectError:            true,
			expectedErrorSubstring: "one or more custom webhooks were detected",
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
