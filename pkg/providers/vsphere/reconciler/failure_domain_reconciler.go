package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileFailureDomains applies the Vsphere FailureDomain objects to the cluster.
func ReconcileFailureDomains(ctx context.Context, log logr.Logger, client client.Client, spec *c.Spec) (controller.Result, error) {
	fd, err := vsphere.FailureDomainsSpec(log, spec)
	if err != nil {
		return controller.Result{}, err
	}
	if err := serverside.ReconcileObjects(ctx, client, fd.Objects()); err != nil {
		return controller.Result{}, errors.Wrap(err, "applying Vsphere Failure Domain objects")
	}
	return controller.Result{}, nil
}
