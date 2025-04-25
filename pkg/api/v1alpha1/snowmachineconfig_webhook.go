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
var snowmachineconfiglog = logf.Log.WithName("snowmachineconfig-resource")

func (r *SnowMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-snowmachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowmachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.snowmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &SnowMachineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (r *SnowMachineConfig) Default(_ context.Context, obj runtime.Object) error {
	snowConfig, ok := obj.(*SnowMachineConfig)
	if !ok {
		return fmt.Errorf("expected a SnowMachineConfig but got %T", obj)
	}

	snowmachineconfiglog.Info("Setting up Snow Machine Config defaults for", "name", snowConfig.Name)
	snowConfig.SetDefaults()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-snowmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.snowmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &SnowMachineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowMachineConfig but got %T", obj)
	}

	snowmachineconfiglog.Info("validate create", "name", snowConfig.Name)

	if err := snowConfig.ValidateHasSSHKeyName(); err != nil {
		return nil, err
	}

	return nil, snowConfig.Validate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateUpdate(_ context.Context, obj, _ runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowMachineConfig but got %T", obj)
	}

	snowmachineconfiglog.Info("validate update", "name", snowConfig.Name)

	if err := snowConfig.ValidateHasSSHKeyName(); err != nil {
		return nil, err
	}

	return nil, snowConfig.Validate()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	snowConfig, ok := obj.(*SnowMachineConfig)
	if !ok {
		return nil, fmt.Errorf("expected a SnowMachineConfig but got %T", obj)
	}

	snowmachineconfiglog.Info("validate delete", "name", snowConfig.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
