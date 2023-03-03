package v1alpha1

import (
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

//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-nutanixdatacenterconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=nutanixdatacenterconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &NutanixDatacenterConfig{}
var _ webhook.Defaulter = &NutanixDatacenterConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateCreate() error {
	nutanixdatacenterconfiglog.Info("validate create", "name", r.Name)

	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateUpdate(old runtime.Object) error {
	nutanixdatacenterconfiglog.Info("validate update", "name", r.Name)

	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) ValidateDelete() error {
	nutanixdatacenterconfiglog.Info("validate delete", "name", r.Name)

	return nil
}

//+kubebuilder:webhook:path=/mutate-anywhere-eks-amazonaws-com-v1alpha1-nutanixdatacenterconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=nutanixdatacenterconfigs,verbs=create;update,versions=v1alpha1,name=mutation.nutanixdatacenterconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *NutanixDatacenterConfig) Default() {
	nutanixdatacenterconfiglog.Info("default", "name", r.Name)
	r.SetDefaults()
}
