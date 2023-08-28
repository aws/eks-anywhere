package cluster

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/util/conditions"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// WaitForCondition blocks until either the cluster has this condition as True
// or the retrier timeouts. If observedGeneration is not equal to generation,
// the condition is considered false regardless of the status value.
func WaitForCondition(ctx context.Context, client kubernetes.Reader, cluster *anywherev1.Cluster, retrier *retrier.Retrier, conditionType anywherev1.ConditionType) error {
	return retrier.Retry(func() error {
		c := &anywherev1.Cluster{}
		if err := client.Get(ctx, cluster.Name, cluster.Namespace, c); err != nil {
			return err
		}

		observedGeneration := c.Status.ObservedGeneration
		generation := c.Generation
		if observedGeneration != generation {
			return errors.Errorf("cluster generation (%d) and observedGeneration (%d) differ", generation, observedGeneration)
		}

		condition := conditions.Get(c, conditionType)
		if condition == nil {
			return errors.Errorf("cluster doesn't yet have condition %s", conditionType)
		}

		if condition.Status != corev1.ConditionTrue {
			return errors.Errorf("cluster condition %s is %s: %s", conditionType, condition.Status, condition.Message)
		}

		return nil
	})
}
