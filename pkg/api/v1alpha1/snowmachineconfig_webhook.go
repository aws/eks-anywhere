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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var snowmachineconfiglog = logf.Log.WithName("snowmachineconfig-resource")

func (r *SnowMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-snowmachineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowmachineconfigs,verbs=create;update,versions=v1alpha1,name=mutation.snowmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &SnowMachineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *SnowMachineConfig) Default() {
	snowmachineconfiglog.Info("Setting up Snow Machine Config defaults for", "name", r.Name)
	r.SetDefaults()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-snowmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=snowmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.snowmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &SnowMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateCreate() error {
	snowmachineconfiglog.Info("validate create", "name", r.Name)

	if err := r.ValidateHasSSHKeyName(); err != nil {
		return err
	}

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateUpdate(old runtime.Object) error {
	snowmachineconfiglog.Info("validate update", "name", r.Name)

	if err := r.ValidateHasSSHKeyName(); err != nil {
		return err
	}

	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *SnowMachineConfig) ValidateDelete() error {
	snowmachineconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
