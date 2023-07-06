package clustermanager_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type managementUpgraderMocks struct {
	ctx     context.Context
	retrier retrier.Retrier
	cluster *types.Cluster
}

func newManagementUpgraderMocks() *managementUpgraderMocks {
	tt := &managementUpgraderMocks{}
	tt.ctx = context.Background()
	tt.cluster = &types.Cluster{
		Name: "mgmt-cluster",
	}
	tt.retrier = *retrier.NewWithMaxRetries(30, 5*time.Second)
	return tt
}

func TestUpgradeManagementClusterSuccess(t *testing.T) {
	tt := newManagementUpgraderMocks()

	ctrl := gomock.NewController(t)
	upgradeClient := mocks.NewMockUpgradeClient(ctrl)

	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ControlPlaneReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.DefaultCNIConfiguredCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.WorkersReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ReadyCondition), "1h0m0s")

	upgradeCluster := clustermanager.NewManagementUpgrader(upgradeClient)

	if err := upgradeCluster.UpgradeManagementCluster(tt.ctx, tt.cluster); err != nil {
		t.Errorf("ClusterManager.UpgradeManagementCluster() error = %v, wantErr nil", err)
	}
}

func TestUpgradeManagementClusterControlPlaneError(t *testing.T) {
	expectedError := "Error waiting for CP to be available"
	tt := newManagementUpgraderMocks()

	ctrl := gomock.NewController(t)
	upgradeClient := mocks.NewMockUpgradeClient(ctrl)

	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ControlPlaneReadyCondition), "1h0m0s").Return(errors.New(expectedError))

	upgradeCluster := clustermanager.NewManagementUpgrader(upgradeClient)

	if err := upgradeCluster.UpgradeManagementCluster(tt.ctx, tt.cluster); err == nil {
		t.Errorf("ClusterManager.UpgradeManagementCluster() error = %v, wantErr %s", err, expectedError)
	}
}

func TestUpgradeManagementClusterCNIError(t *testing.T) {
	expectedError := "Error waiting for default CNI to be available"
	tt := newManagementUpgraderMocks()

	ctrl := gomock.NewController(t)
	upgradeClient := mocks.NewMockUpgradeClient(ctrl)

	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ControlPlaneReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.DefaultCNIConfiguredCondition), "1h0m0s").Return(errors.New(expectedError))

	upgradeCluster := clustermanager.NewManagementUpgrader(upgradeClient)
	if err := upgradeCluster.UpgradeManagementCluster(tt.ctx, tt.cluster); err == nil {
		t.Errorf("ClusterManager.UpgradeManagementCluster() error = %v, wantErr %s", err, expectedError)
	}
}

func TestUpgradeManagementClusterWorkerReadyError(t *testing.T) {
	expectedError := "Error waiting for worker nodes to be ready"
	tt := newManagementUpgraderMocks()

	ctrl := gomock.NewController(t)
	upgradeClient := mocks.NewMockUpgradeClient(ctrl)

	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ControlPlaneReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.DefaultCNIConfiguredCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.WorkersReadyCondition), "1h0m0s").Return(errors.New(expectedError))

	upgradeCluster := clustermanager.NewManagementUpgrader(upgradeClient)
	if err := upgradeCluster.UpgradeManagementCluster(tt.ctx, tt.cluster); err == nil {
		t.Errorf("ClusterManager.UpgradeManagementCluster() error = %v, wantErr %s", err, expectedError)
	}
}

func TestUpgradeManagementClusterClusterReadyError(t *testing.T) {
	expectedError := "Error waiting for cluster to be ready"
	tt := newManagementUpgraderMocks()

	ctrl := gomock.NewController(t)
	upgradeClient := mocks.NewMockUpgradeClient(ctrl)

	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ControlPlaneReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.DefaultCNIConfiguredCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.WorkersReadyCondition), "1h0m0s")
	upgradeClient.EXPECT().WaitForClusterCondition(tt.ctx, tt.cluster, gomock.Any(), string(v1alpha1.ReadyCondition), "1h0m0s").Return(errors.New(expectedError))

	upgradeCluster := clustermanager.NewManagementUpgrader(upgradeClient)
	if err := upgradeCluster.UpgradeManagementCluster(tt.ctx, tt.cluster); err == nil {
		t.Errorf("ClusterManager.UpgradeManagementCluster() error = %v, wantErr %s", err, expectedError)
	}
}
