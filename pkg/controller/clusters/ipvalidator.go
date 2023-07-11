package clusters

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// IPUniquenessValidator defines an interface for the methods to validate the control plane IP.
type IPUniquenessValidator interface {
	ValidateControlPlaneIPUniqueness(cluster *anywherev1.Cluster) error
}

// IPValidator validates control plane IP.
type IPValidator struct {
	ipUniquenessValidator IPUniquenessValidator
	client                client.Client
}

// NewIPValidator returns a new NewIPValidator.
func NewIPValidator(ipUniquenessValidator IPUniquenessValidator, client client.Client) *IPValidator {
	return &IPValidator{
		ipUniquenessValidator: ipUniquenessValidator,
		client:                client,
	}
}

// ValidateControlPlaneIP only validates IP on cluster creation.
func (i *IPValidator) ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	capiCluster, err := controller.GetCAPICluster(ctx, i.client, spec.Cluster)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "validating control plane IP")
	}
	if capiCluster != nil {
		// If CAPI cluster exists, the control plane IP has already been validated,
		// and it's possibly already in use so no need to validate it again
		log.Info("CAPI cluster already exists, skipping control plane IP validation")
		return controller.Result{}, nil
	}
	if err := i.ipUniquenessValidator.ValidateControlPlaneIPUniqueness(spec.Cluster); err != nil {
		spec.Cluster.SetFailure(anywherev1.UnavailableControlPlaneIPReason, err.Error())
		log.Error(err, "Unavailable control plane IP")
		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}
