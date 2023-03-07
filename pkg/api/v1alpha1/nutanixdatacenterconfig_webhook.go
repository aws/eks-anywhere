package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// nutanixdatacenterconfiglog is for logging in this package.
var nutanixdatacenterconfiglog = logf.Log.WithName("nutanixdatacenterconfig-resource")

// SetupWebhookWithManager sets up the webhook with the manager.
func (r *NutanixDatacenterConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=validation.nutanixdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &NutanixDatacenterConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateCreate() error {
	nutanixdatacenterconfiglog.Info("validate create", "name", r.Name)
	if r.Spec.CredentialRef == nil {
		return fmt.Errorf("credentialRef is required to be set to create a new NutanixDatacenterConfig")
	}

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateUpdate(old runtime.Object) error {
	nutanixdatacenterconfiglog.Info("validate update", "name", r.Name)
	if r.Spec.CredentialRef == nil {
		// check if the old object has a credentialRef set
		oldNutanixDatacenterConfig, ok := old.(*NutanixDatacenterConfig)
		if !ok {
			return fmt.Errorf("old object is not a NutanixDatacenterConfig")
		}
		if oldNutanixDatacenterConfig.Spec.CredentialRef != nil {
			return fmt.Errorf("credentialRef cannot be removed from an existing NutanixDatacenterConfig")
		}
	}

	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateDelete() error {
	nutanixdatacenterconfiglog.Info("validate delete", "name", r.Name)

	return nil
}
