/*
Copyright 2022 The Tinkerbell Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	osUbuntu             = "ubuntu"
	defaultUbuntuVersion = "20.04"
)

var _ admission.Validator = &TinkerbellCluster{}

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (c *TinkerbellCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(c).Complete() //nolint:wrapcheck
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-tinkerbellcluster,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=tinkerbellclusters,versions=v1beta1,name=validation.tinkerbellcluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-tinkerbellcluster,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=tinkerbellclusters,versions=v1beta1,name=default.tinkerbellcluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateCreate() (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func defaultVersionForOSDistro(distro string) string {
	if strings.ToLower(distro) == osUbuntu {
		return defaultUbuntuVersion
	}

	return ""
}

// Default implements webhookutil.defaulter so a webhook will be registered for the type.
func (c *TinkerbellCluster) Default() {
	if c.Spec.ImageLookupFormat == "" {
		c.Spec.ImageLookupFormat = "{{.BaseRegistry}}/{{.OSDistro}}-{{.OSVersion}}:{{.KubernetesVersion}}.gz"
	}

	if c.Spec.ImageLookupOSVersion == "" {
		c.Spec.ImageLookupOSVersion = defaultVersionForOSDistro(c.Spec.ImageLookupOSDistro)
	}
}
