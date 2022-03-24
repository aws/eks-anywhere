package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/aws/eks-anywhere/pkg/features"
)

// log is for logging in this package.
var cloudstackmachineconfiglog = logf.Log.WithName("cloudstackmachineconfig-resource")

func (r *CloudStackMachineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-anywhere-eks-amazonaws-com-v1alpha1-cloudstackmachineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=anywhere.eks.amazonaws.com,resources=cloudstackmachineconfigs,verbs=create;update,versions=v1alpha1,name=validation.cloudstackmachineconfig.anywhere.amazonaws.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &CloudStackMachineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateCreate() error {
	cloudstackmachineconfiglog.Info("validate create", "name", r.Name)

	if !features.IsActive(features.CloudStackProvider()) {
		return apierrors.NewBadRequest("CloudStackProvider feature is not active, preventing CloudStackMachineConfig resource creation")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateUpdate(old runtime.Object) error {
	cloudstackmachineconfiglog.Info("validate update", "name", r.Name)

	oldCloudStackMachineConfig, ok := old.(*CloudStackMachineConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a CloudStackMachineConfig but got a %T", old))
	}

	if oldCloudStackMachineConfig.IsReconcilePaused() {
		cloudstackmachineconfiglog.Info("Reconciliation is paused")
		return nil
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CloudStackMachineConfig) ValidateDelete() error {
	cloudstackmachineconfiglog.Info("validate delete", "name", r.Name)

	return nil
}
