package clustermanager

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ManagementUpgrader contains cluster upgrade tasks with retry and timeout feature.
type ManagementUpgrader struct {
	upgradeClient           UpgradeClient
	retrier                 retrier.Retrier
	controlPlaneWaitTimeout time.Duration
}

// UpgradeClient builds Kubernetes clients.
type UpgradeClient interface {
	WaitForClusterCondition(ctx context.Context, cluster *types.Cluster, retryClient retrier.Retrier, condition string, conditionWaitTimeout string) error
}

// ManagementUpgraderOpt allows to customize a ManagementUpgrader
// on construction.
type ManagementUpgraderOpt func(*ManagementUpgrader)

// NewManagementUpgrader constructs a new ManagementUpgrader.
func NewManagementUpgrader(upgradeClient UpgradeClient, opts ...ManagementUpgraderOpt) *ManagementUpgrader {
	c := &ManagementUpgrader{
		upgradeClient:           upgradeClient,
		retrier:                 *retrier.NewWithMaxRetries(maxRetries, defaultBackOffPeriod),
		controlPlaneWaitTimeout: DefaultControlPlaneWait,
	}

	for _, o := range opts {
		fmt.Println(o)
		o(c)
	}

	return c
}

// WithUpgraderNoTimeouts disables the timeout for all the waits and retries in management upgrader.
func WithUpgraderNoTimeouts() ManagementUpgraderOpt {
	return func(c *ManagementUpgrader) {
		noTimeoutRetrier := *retrier.NewWithNoTimeout()
		maxTime := time.Duration(math.MaxInt64)
		c.retrier = noTimeoutRetrier
		c.controlPlaneWaitTimeout = maxTime
	}
}

// UpgradeManagementCluster incorporates EKS-A cluster status changes for cluster upgrade process.
func (c *ManagementUpgrader) UpgradeManagementCluster(ctx context.Context, managementCluster *types.Cluster) error {
	logger.V(3).Info("Waiting for control plane to be ready")
	err := c.upgradeClient.WaitForClusterCondition(ctx, managementCluster, c.retrier, string(v1alpha1.ControlPlaneReadyCondition), c.controlPlaneWaitTimeout.String())
	if err != nil {
		return fmt.Errorf("waiting for management cluster control plane to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for default CNI to be updated")
	err = c.upgradeClient.WaitForClusterCondition(ctx, managementCluster, c.retrier, string(v1alpha1.DefaultCNIConfiguredCondition), c.controlPlaneWaitTimeout.String())
	if err != nil {
		return fmt.Errorf("waiting for management cluster default CNI to be configured: %v", err)
	}

	logger.V(3).Info("Waiting for worker nodes to be ready after upgrade")
	err = c.upgradeClient.WaitForClusterCondition(ctx, managementCluster, c.retrier, string(v1alpha1.WorkersReadyCondition), c.controlPlaneWaitTimeout.String())
	if err != nil {
		return fmt.Errorf("waiting for management cluster control plane to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for cluster upgrade to be completed")
	err = c.upgradeClient.WaitForClusterCondition(ctx, managementCluster, c.retrier, string(v1alpha1.ReadyCondition), c.controlPlaneWaitTimeout.String())
	if err != nil {
		return fmt.Errorf("waiting for management cluster to be ready: %v", err)
	}

	return nil
}
