package clustermanager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestApplyKubeSpecFromBytes(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := types.Cluster{}

	client.EXPECT().ApplyKubeSpecFromBytes(context.Background(), &cluster, []byte{}).MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.ApplyKubeSpecFromBytes(context.Background(), &cluster, []byte{})
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestApply(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := v1alpha1.Cluster{}

	client.EXPECT().Apply(context.Background(), "", &cluster).MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.Apply(context.Background(), "", &cluster)
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestPauseCAPICluster(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().PauseCAPICluster(context.Background(), "cluster", "kubeconfig").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.PauseCAPICluster(context.Background(), "cluster", "kubeconfig")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestResumeCAPICluster(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().ResumeCAPICluster(context.Background(), "cluster", "kubeconfig").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.ResumeCAPICluster(context.Background(), "cluster", "kubeconfig")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestApplyKubeSpecFromBytesForce(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := types.Cluster{}

	client.EXPECT().ApplyKubeSpecFromBytesForce(context.Background(), &cluster, []byte{}).MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.ApplyKubeSpecFromBytesForce(context.Background(), &cluster, []byte{})
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestApplyKubeSpecFromBytesWithNamespace(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := types.Cluster{}

	client.EXPECT().ApplyKubeSpecFromBytesWithNamespace(context.Background(), &cluster, []byte{}, "test").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.ApplyKubeSpecFromBytesWithNamespace(context.Background(), &cluster, []byte{}, "test")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestUpdateAnnotationInNamespace(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := types.Cluster{}
	testMap := map[string]string{}

	client.EXPECT().UpdateAnnotationInNamespace(context.Background(), "test", "test", testMap, &cluster, "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.UpdateAnnotationInNamespace(context.Background(), "test", "test", testMap, &cluster, "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestRemoveAnnotationInNamespace(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)
	cluster := types.Cluster{}

	client.EXPECT().RemoveAnnotationInNamespace(context.Background(), "resourceType", "objectName", "key", &cluster, "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.RemoveAnnotationInNamespace(context.Background(), "resourceType", "objectName", "key", &cluster, "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestListObjects(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().ListObjects(context.Background(), "resourceType", "namespace", "kubeconfig", &v1alpha1.ClusterList{}).MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.ListObjects(context.Background(), "resourceType", "namespace", "kubeconfig", &v1alpha1.ClusterList{})
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestDeleteGitOpsConfig(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().DeleteGitOpsConfig(context.Background(), &types.Cluster{}, "name", "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.DeleteGitOpsConfig(context.Background(), &types.Cluster{}, "name", "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestDeleteEKSACluster(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().DeleteEKSACluster(context.Background(), &types.Cluster{}, "name", "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.DeleteEKSACluster(context.Background(), &types.Cluster{}, "name", "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestDeleteAWSIamConfig(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().DeleteAWSIamConfig(context.Background(), &types.Cluster{}, "name", "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.DeleteAWSIamConfig(context.Background(), &types.Cluster{}, "name", "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestDeleteOIDCConfig(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().DeleteOIDCConfig(context.Background(), &types.Cluster{}, "name", "namespace").MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.DeleteOIDCConfig(context.Background(), &types.Cluster{}, "name", "namespace")
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}

func TestDeleteCluster(t *testing.T) {
	tt := NewWithT(t)
	ctrl := gomock.NewController(t)
	client := mocks.NewMockKubernetesClient(ctrl)

	client.EXPECT().DeleteCluster(context.Background(), &types.Cluster{}, &types.Cluster{}).MinTimes(2).Return(errors.New("this is an error"))
	retrierClient := clustermanager.NewRetrierClient(client, retrier.NewWithMaxRetries(2, 0))

	err := retrierClient.DeleteCluster(context.Background(), &types.Cluster{}, &types.Cluster{})
	tt.Expect(err.Error()).To(ContainSubstring("this is an error"))
}
