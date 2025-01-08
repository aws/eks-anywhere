package clustermanager

import (
	"context"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// MoverOpt allows to customize a Mover on construction.
type MoverOpt func(*Mover)

// Mover applies the cluster spec to the management cluster and waits
// until the changes are fully reconciled.
type Mover struct {
	log                logr.Logger
	clientFactory      ClientFactory
	moveClusterTimeout time.Duration
	retryBackOff       time.Duration
}

// NewMover builds an Mover.
func NewMover(log logr.Logger, clientFactory ClientFactory, opts ...MoverOpt) *Mover {
	m := &Mover{
		log:                log,
		clientFactory:      clientFactory,
		moveClusterTimeout: applyClusterSpecTimeout,
		retryBackOff:       retryBackOff,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithMoverNoTimeouts disables the timeout for all the waits and retries in management upgrader.
func WithMoverNoTimeouts() MoverOpt {
	return func(a *Mover) {
		maxTime := time.Duration(math.MaxInt64)
		a.moveClusterTimeout = maxTime
	}
}

// WithMoverApplyClusterTimeout allows to configure how long the mover retries
// to apply the objects in case of failure.
// Generally only used in tests.
func WithMoverApplyClusterTimeout(timeout time.Duration) MoverOpt {
	return func(m *Mover) {
		m.moveClusterTimeout = timeout
	}
}

// WithMoverRetryBackOff allows to configure how long the mover waits between requests
// to update the cluster spec objects and check the status of the Cluster.
// Generally only used in tests.
func WithMoverRetryBackOff(backOff time.Duration) MoverOpt {
	return func(m *Mover) {
		m.retryBackOff = backOff
	}
}

// Move applies the cluster's namespace and spec without checking for reconcile conditions.
func (m *Mover) Move(ctx context.Context, spec *cluster.Spec, fromClient, toClient kubernetes.Client) error {
	m.log.V(3).Info("Moving the cluster object")
	err := retrier.New(
		m.moveClusterTimeout,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(m.retryBackOff)),
	).Retry(func() error {
		// read the cluster from bootstrap
		cluster := &v1alpha1.Cluster{}
		if err := fromClient.Get(ctx, spec.Cluster.Name, spec.Cluster.Namespace, cluster); err != nil {
			return errors.Wrapf(err, "reading cluster from source")
		}

		// pause cluster on bootstrap
		cluster.PauseReconcile()

		// For baremetal provider we need to clear this annotation once we move to management cluster
		// at this point all the hardware is provisioned and tink stack is up on the management cluster.
		// We don't need the Bootstrap IP anymore.
		// ideally this logic should be outside of move but the current code structure needs refactor
		// to be able to do that.
		cluster.ClearTinkerbellIPAnnotation()
		if err := fromClient.Update(ctx, cluster); err != nil {
			return errors.Wrapf(err, "updating cluster on source")
		}

		if err := moveClusterResource(ctx, cluster, toClient); err != nil {
			return err
		}

		if err := moveChildObjects(ctx, spec, fromClient, toClient); err != nil {
			return err
		}

		return nil
	})

	return err
}

func moveClusterResource(ctx context.Context, cluster *v1alpha1.Cluster, client kubernetes.Client) error {
	cluster.ResourceVersion = ""
	cluster.UID = ""

	// move eksa cluster
	if err := client.Create(ctx, cluster); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "moving cluster %s", cluster.Name)
	}

	return nil
}

func moveChildObjects(ctx context.Context, spec *cluster.Spec, fromClient, toClient kubernetes.Client) error {
	// read and move child objects
	for _, child := range spec.ChildObjects() {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(child.GetObjectKind().GroupVersionKind())
		if err := fromClient.Get(ctx, child.GetName(), child.GetNamespace(), obj); err != nil {
			return errors.Wrapf(err, "reading child object %s %s", child.GetObjectKind().GroupVersionKind().Kind, child.GetName())
		}

		obj.SetResourceVersion("")
		obj.SetUID("")
		obj.SetOwnerReferences(nil)

		if err := toClient.Create(ctx, obj); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "moving child object %s %s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		}
	}

	return nil
}
