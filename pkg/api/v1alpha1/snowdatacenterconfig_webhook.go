// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var snowdatacenterconfiglog = logf.Log.WithName("snowdatacenterconfig-resource")

func (r *SnowDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-snowdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=snowdatacenterconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &SnowDatacenterConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowDatacenterConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowDatacenterConfig but got %T", obj)
	}

	snowdatacenterconfiglog.Info("validate create", "name", snowConfig.Name)

	return nil, snowConfig.Validate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowDatacenterConfig) ValidateUpdate(_ context.Context, obj, _ runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowDatacenterConfig but got %T", obj)
	}

	snowdatacenterconfiglog.Info("validate update", "name", snowConfig.Name)

	return nil, snowConfig.Validate()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowDatacenterConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowDatacenterConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowDatacenterConfig but got %T", obj)
	}

	snowdatacenterconfiglog.Info("validate delete", "name", snowConfig.Name)

	return nil, nil
}
