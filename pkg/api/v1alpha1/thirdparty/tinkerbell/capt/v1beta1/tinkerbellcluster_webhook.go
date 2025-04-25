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
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	osUbuntu             = "ubuntu"
	defaultUbuntuVersion = "20.04"
)

var (
	_ webhook.CustomValidator = &TinkerbellCluster{}
	_ webhook.CustomDefaulter = &TinkerbellCluster{}
)

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (c *TinkerbellCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(c).WithDefaulter(c).
		WithValidator(c).Complete() //nolint:wrapcheck
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateCreate(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateUpdate(_ context.Context, _, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (c *TinkerbellCluster) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func defaultVersionForOSDistro(distro string) string {
	if strings.ToLower(distro) == osUbuntu {
		return defaultUbuntuVersion
	}

	return ""
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (c *TinkerbellCluster) Default(_ context.Context, obj runtime.Object) error {
	cluster, ok := obj.(*TinkerbellCluster)
	if !ok {
		return fmt.Errorf("expected a TinkerbellCluster but got %T", obj)
	}

	if cluster.Spec.ImageLookupFormat == "" {
		cluster.Spec.ImageLookupFormat = "{{.BaseRegistry}}/{{.OSDistro}}-{{.OSVersion}}:{{.KubernetesVersion}}.gz"
	}

	if cluster.Spec.ImageLookupOSVersion == "" {
		cluster.Spec.ImageLookupOSVersion = defaultVersionForOSDistro(cluster.Spec.ImageLookupOSDistro)
	}

	return nil
}
