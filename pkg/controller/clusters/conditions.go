package clusters

import (
	"context"
	"sync"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConditionUpdater calculates a condition from the Cluster state and updates the Cluster status.
type ConditionUpdater func(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) error

// ConditionChecker is composed of a set of ConditionUpdater.
type ConditionChecker []ConditionUpdater

// Register registers updaters with c.
func (c *ConditionChecker) Register(updaters ...ConditionUpdater) {
	*c = append(*c, updaters...)
}

// Runs the registered ConditionUpdater against the Cluster.
func (c *ConditionChecker) Run(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) error {
	updaters := *c
	wg := &sync.WaitGroup{}
	wg.Add(len(updaters))
	errList := make([]error, 0)

	for _, updater := range updaters {
		go func(updater ConditionUpdater) {
			defer wg.Done()
			err := updater(ctx, client, clusterSpec)
			if err != nil {
				errList = append(errList, err)
			}
		}(updater)
	}

	wg.Wait()
	return errors.NewAggregate(errList)
}

// NewConditionChecker creates a ConditionChecker and any updaters passed will be registered.
func NewConditionChecker(updaters ...ConditionUpdater) *ConditionChecker {
	var v ConditionChecker
	v.Register(updaters...)
	return &v
}
