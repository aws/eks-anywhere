package vsphere

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FailureDomainMover struct {
	client  client.Client
	log     logr.Logger
	cluster *anywherev1.Cluster
}

func NewFailureDomainMover(
	log logr.Logger,
	client client.Client,
	cluster *anywherev1.Cluster,
) *FailureDomainMover {
	return &FailureDomainMover{
		client:  client,
		log:     log,
		cluster: cluster,
	}
}

func (m *FailureDomainMover) ApplyFailureDomains(ctx context.Context) error {
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(m.client), m.cluster)
	if err != nil {
		return err
	}

	fd, err := FailureDomainsSpec(m.log, clusterSpec)
	if err != nil {
		return err
	}

	if err := serverside.ReconcileObjects(ctx, m.client, fd.Objects()); err != nil {
		return err
	}

	return nil
}
