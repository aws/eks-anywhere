package clustermanager

import (
	"context"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// DeleterOpt allows to customize a Deleter on construction.
type DeleterOpt func(*Deleter)

// Deleter deletes the cluster from the management cluster and waits
// until the deletions are fully reconciled.
type Deleter struct {
	log                         logr.Logger
	clientFactory               ClientFactory
	deleteClusterTimeout        time.Duration
	retryBackOff                time.Duration
	conditionCheckoutTotalCount int
}

const deleteClusterSpecTimeout = 5 * time.Minute

// NewDeleter builds an Deleter.
func NewDeleter(log logr.Logger, clientFactory ClientFactory, opts ...DeleterOpt) Deleter {
	a := &Deleter{
		log:                         log,
		clientFactory:               clientFactory,
		deleteClusterTimeout:        deleteClusterSpecTimeout,
		retryBackOff:                retryBackOff,
		conditionCheckoutTotalCount: defaultConditionCheckTotalCount,
	}

	for _, opt := range opts {
		opt(a)
	}

	return *a
}

// WithDeleterNoTimeouts disables the timeout for all the waits and retries in management upgrader.
func WithDeleterNoTimeouts() DeleterOpt {
	return func(a *Deleter) {
		maxTime := time.Duration(math.MaxInt64)
		a.deleteClusterTimeout = maxTime
	}
}

// WithDeleterApplyClusterTimeout allows to configure how long the deleter retries
// to delete the objects in case of failure.
// Generally only used in tests.
func WithDeleterApplyClusterTimeout(timeout time.Duration) DeleterOpt {
	return func(a *Deleter) {
		a.deleteClusterTimeout = timeout
	}
}

// WithDeleterRetryBackOff allows to configure how long the deleter waits between requests
// to update the cluster spec objects and check the status of the Cluster.
// Generally only used in tests.
func WithDeleterRetryBackOff(backOff time.Duration) DeleterOpt {
	return func(a *Deleter) {
		a.retryBackOff = backOff
	}
}

// Run deletes the cluster's spec in the management cluster and waits
// until the changes are fully reconciled.
func (a Deleter) Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error {
	var client kubernetes.Client
	a.log.V(3).Info("Deleting cluster spec")
	err := retrier.New(
		a.deleteClusterTimeout,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(a.retryBackOff)),
	).Retry(func() error {
		var err error
		client, err = a.clientFactory.BuildClientFromKubeconfig(managementCluster.KubeconfigFile)
		if err != nil {
			return errors.Wrap(err, "building client to delete cluster")
		}

		if err := client.Delete(ctx, spec.Cluster); err != nil {
			return errors.Wrapf(err, "deleting cluster")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
