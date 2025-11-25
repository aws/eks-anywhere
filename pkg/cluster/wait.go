package cluster

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/util/conditions"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// WaitForCondition blocks until either the cluster has this condition as True
// or the retrier timeouts. If observedGeneration is not equal to generation,
// the condition is considered false regardless of the status value.
// total field is to check the total number of times the given condition is met for consistency.
func WaitForCondition(ctx context.Context, log logr.Logger, client kubernetes.Reader, cluster *anywherev1.Cluster, total int, retrier *retrier.Retrier, conditionType anywherev1.ConditionType) error {
	return WaitFor(ctx, log, client, cluster, total, retrier, func(c *anywherev1.Cluster) error {
		condition := conditions.Get(c, conditionType)
		if condition == nil {
			return fmt.Errorf("cluster doesn't yet have condition %s", conditionType)
		}

		if condition.Status != corev1.ConditionTrue {
			return fmt.Errorf("cluster condition %s is %s: %s", conditionType, condition.Status, condition.Message)
		}
		return nil
	})
}

// Matcher matches the given condition.
type Matcher func(*anywherev1.Cluster) error

// WaitFor gets the cluster object from the client
// checks for generation and observedGeneration condition
// matches condition and returns error if the condition is not met.
func WaitFor(ctx context.Context, log logr.Logger, client kubernetes.Reader, cluster *anywherev1.Cluster, total int, retrier *retrier.Retrier, matcher Matcher) error {
	count := 0
	return retrier.Retry(func() error {
		c := &anywherev1.Cluster{}

		namespace := cluster.Namespace
		if namespace == "" {
			namespace = constants.DefaultNamespace
		}

		if err := client.Get(ctx, cluster.Name, namespace, c); err != nil {
			return err
		}

		observedGeneration := c.Status.ObservedGeneration
		generation := c.Generation

		log.V(9).Info("Cluster generation and observedGeneration", "Generation", generation, "ObservedGeneration", observedGeneration)

		if observedGeneration != generation {
			return fmt.Errorf("cluster generation (%d) and observedGeneration (%d) differ", generation, observedGeneration)
		}

		if err := matcher(c); err != nil {
			count = 0
			return err
		}

		count++
		// total field is to check the total number of times the given condition is met.
		// for ex, the total is set to 5 and we want to check certain condition is met
		// we check the condition is met for 5 times to make sure the given behavior is consistent.
		if count < total {
			return fmt.Errorf("cluster has reached to expected condition in %d/%d times", count, total)
		}

		// when the count matches total number it returns without error
		return nil
	})
}
