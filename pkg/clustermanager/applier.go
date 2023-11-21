package clustermanager

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	applyClusterSpecTimeout           = 2 * time.Minute
	waitForClusterReconcileTimeout    = time.Hour
	retryBackOff                      = time.Second
	waitForFailureMessageErrorTimeout = 10 * time.Minute
	defaultFieldManager               = "eks-a-cli"
	defaultConditionCheckTotalCount   = 20
)

// ApplierOpt allows to customize a Applier on construction.
type ApplierOpt func(*Applier)

// Applier applies the cluster spec to the management cluster and waits
// until the changes are fully reconciled.
type Applier struct {
	log                                                                 logr.Logger
	clientFactory                                                       ClientFactory
	applyClusterTimeout, waitForClusterReconcile, waitForFailureMessage time.Duration
	retryBackOff                                                        time.Duration
	conditionCheckoutTotalCount                                         int
}

// NewApplier builds an Applier.
func NewApplier(log logr.Logger, clientFactory ClientFactory, opts ...ApplierOpt) Applier {
	a := &Applier{
		log:                         log,
		clientFactory:               clientFactory,
		applyClusterTimeout:         applyClusterSpecTimeout,
		waitForClusterReconcile:     waitForClusterReconcileTimeout,
		waitForFailureMessage:       waitForFailureMessageErrorTimeout,
		retryBackOff:                retryBackOff,
		conditionCheckoutTotalCount: defaultConditionCheckTotalCount,
	}

	for _, opt := range opts {
		opt(a)
	}

	return *a
}

// WithApplierNoTimeouts disables the timeout for all the waits and retries in management upgrader.
func WithApplierNoTimeouts() ApplierOpt {
	return func(a *Applier) {
		maxTime := time.Duration(math.MaxInt64)
		a.applyClusterTimeout = maxTime
		a.waitForClusterReconcile = maxTime
		a.waitForFailureMessage = maxTime
	}
}

// WithApplierApplyClusterTimeout allows to configure how long the applier retries
// to apply the objects in case of failure.
// Generally only used in tests.
func WithApplierApplyClusterTimeout(timeout time.Duration) ApplierOpt {
	return func(a *Applier) {
		a.applyClusterTimeout = timeout
	}
}

// WithApplierWaitForClusterReconcile allows to configure how long the applier waits
// for the cluster to reach the Ready state after applying changes.
// Generally only used in tests.
func WithApplierWaitForClusterReconcile(timeout time.Duration) ApplierOpt {
	return func(a *Applier) {
		a.waitForClusterReconcile = timeout
	}
}

// WithApplierRetryBackOff allows to configure how long the applier waits between requests
// to update the cluster spec objects and check the status of the Cluster.
// Generally only used in tests.
func WithApplierRetryBackOff(backOff time.Duration) ApplierOpt {
	return func(a *Applier) {
		a.retryBackOff = backOff
	}
}

// WithApplierWaitForFailureMessage allows to configure how long the applier waits for failure message
// to be empty and check the status of the Cluster.
// Generally only used in tests.
func WithApplierWaitForFailureMessage(timeout time.Duration) ApplierOpt {
	return func(a *Applier) {
		a.waitForFailureMessage = timeout
		a.conditionCheckoutTotalCount = int(timeout)
	}
}

// Run applies the cluster's spec in the management cluster and waits
// until the changes are fully reconciled.
func (a Applier) Run(ctx context.Context, spec *cluster.Spec, managementCluster types.Cluster) error {
	var client kubernetes.Client
	a.log.V(3).Info("Applying cluster spec")
	err := retrier.New(
		a.applyClusterTimeout,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(a.retryBackOff)),
	).Retry(func() error {
		// We build the client inside the retrier to take advantage of the configurable timeout.
		// The only time when a client can fail to init is when there is a transient error contacting
		// the api-server (there might be some validations checks during init). These are the same
		// transient errors we want to protect the apply against, so it makes sense to group them together.
		// The drawback here is we rebuild the client on each retry if one of the objects fails to be
		// created/updated. It's a fair tradeoff given this is already handling an edge case.
		var err error
		client, err = a.clientFactory.BuildClientFromKubeconfig(managementCluster.KubeconfigFile)
		if err != nil {
			return errors.Wrap(err, "building client to apply cluster spec changes")
		}

		for _, obj := range spec.ClusterAndChildren() {
			if err := client.ApplyServerSide(ctx,
				defaultFieldManager,
				obj,
				kubernetes.ApplyServerSideOptions{ForceOwnership: true},
			); err != nil {
				return errors.Wrapf(err, "applying cluster spec")
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// We use this start time to compute the leftover time on each condition wait
	waitStartTime := time.Now()
	retry := a.retrierForWait(waitStartTime)

	if err := cluster.WaitFor(ctx, a.log, client, spec.Cluster, a.conditionCheckoutTotalCount, a.retrierForFailureMessage(), func(c *anywherev1.Cluster) error {
		if c.Status.FailureMessage != nil && *c.Status.FailureMessage != "" {
			return fmt.Errorf("cluster has an error: %s", *c.Status.FailureMessage)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cluster has a validation error that doesn't seem transient: %s", err)
	}

	a.log.V(3).Info("Waiting for control plane to be ready")
	if err := cluster.WaitForCondition(ctx, a.log, client, spec.Cluster, a.conditionCheckoutTotalCount, retry, anywherev1.ControlPlaneReadyCondition); err != nil {
		return errors.Wrapf(err, "waiting for cluster's control plane to be ready")
	}

	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.IsManaged() {
		a.log.V(3).Info("Waiting for default CNI to be updated")
		retry = a.retrierForWait(waitStartTime)
		if err := cluster.WaitForCondition(ctx, a.log, client, spec.Cluster, a.conditionCheckoutTotalCount, retry, anywherev1.DefaultCNIConfiguredCondition); err != nil {
			return errors.Wrapf(err, "waiting for cluster's CNI to be configured")
		}
	}

	a.log.V(3).Info("Waiting for worker nodes to be ready")
	retry = a.retrierForWait(waitStartTime)
	if err := cluster.WaitForCondition(ctx, a.log, client, spec.Cluster, a.conditionCheckoutTotalCount, retry, anywherev1.WorkersReadyCondition); err != nil {
		return errors.Wrapf(err, "waiting for cluster's workers to be ready")
	}

	a.log.V(3).Info("Waiting for cluster changes to be completed")
	retry = a.retrierForWait(waitStartTime)
	if err := cluster.WaitForCondition(ctx, a.log, client, spec.Cluster, a.conditionCheckoutTotalCount, retry, anywherev1.ReadyCondition); err != nil {
		return errors.Wrapf(err, "waiting for cluster to be ready")
	}

	return nil
}

func (a Applier) retrierForWait(waitStartTime time.Time) *retrier.Retrier {
	return retrier.New(
		a.waitForClusterReconcile-time.Since(waitStartTime),
		retrier.WithRetryPolicy(retrier.BackOffPolicy(a.retryBackOff)),
	)
}

func (a Applier) retrierForFailureMessage() *retrier.Retrier {
	return retrier.New(
		a.waitForFailureMessage,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(a.retryBackOff)),
	)
}
