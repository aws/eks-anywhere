package awsiamauth_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/utils/pointer"

	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	filewritermock "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	kubeconfigmocks "github.com/aws/eks-anywhere/pkg/kubeconfig/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestGenerateManagementAWSIAMKubeconfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	k8s := NewMockKubernetesClient(ctrl)
	secretValue := []byte("kubeconfig")
	k8s.EXPECT().GetAWSIAMKubeconfigSecretValue(gomock.Any(), gomock.Any(), gomock.Any()).Return(secretValue, nil)

	cluster := &types.Cluster{
		Name: "cluster-name",
	}
	writer := filewritermock.NewMockFileWriter(ctrl)
	fileName := fmt.Sprintf("%s-aws.kubeconfig", cluster.Name)
	path := "testpath"
	fileWriter := os.NewFile(uintptr(*pointer.Uint(0)), "test")

	writer.EXPECT().Create(fileName, gomock.AssignableToTypeOf([]filewriter.FileOptionsFunc{})).Return(fileWriter, path, nil)
	kwriter := kubeconfigmocks.NewMockWriter(ctrl)
	installer := awsiamauth.NewInstaller(k8s, writer, kwriter)
	kwriter.EXPECT().WriteKubeconfigContent(ctx, cluster.Name, secretValue, fileWriter)

	err := installer.GenerateManagementKubeconfig(context.Background(), cluster)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateManagementAWSIAMKubeconfigError(t *testing.T) {
	ctrl := gomock.NewController(t)

	k8s := NewMockKubernetesClient(ctrl)
	k8s.EXPECT().GetAWSIAMKubeconfigSecretValue(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test"))

	cluster := &types.Cluster{
		Name: "cluster-name",
	}
	writer := filewritermock.NewMockFileWriter(ctrl)
	fileName := fmt.Sprintf("%s-aws.kubeconfig", cluster.Name)
	path := "testpath"
	fileWriter := os.NewFile(uintptr(*pointer.Uint(0)), "test")

	writer.EXPECT().Create(fileName, gomock.AssignableToTypeOf([]filewriter.FileOptionsFunc{})).Return(fileWriter, path, nil)
	kwriter := kubeconfigmocks.NewMockWriter(ctrl)
	installer := awsiamauth.NewInstaller(k8s, writer, kwriter)

	err := installer.GenerateManagementKubeconfig(context.Background(), cluster)
	if err == nil {
		t.Fatal(err)
	}
}

func TestGenerateAWSIAMKubeconfigError(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	k8s := NewMockKubernetesClient(ctrl)
	secretValue := []byte("kubeconfig")
	k8s.EXPECT().GetAWSIAMKubeconfigSecretValue(gomock.Any(), gomock.Any(), gomock.Any()).Return(secretValue, nil)

	cluster := &types.Cluster{
		Name: "cluster-name",
	}
	writer := filewritermock.NewMockFileWriter(ctrl)
	fileName := fmt.Sprintf("%s-aws.kubeconfig", cluster.Name)
	path := "testpath"
	fileWriter := os.NewFile(uintptr(*pointer.Uint(0)), "test")

	writer.EXPECT().Create(fileName, gomock.AssignableToTypeOf([]filewriter.FileOptionsFunc{})).Return(fileWriter, path, nil)
	kwriter := kubeconfigmocks.NewMockWriter(ctrl)
	installer := awsiamauth.NewInstaller(k8s, writer, kwriter)
	kwriter.EXPECT().WriteKubeconfigContent(ctx, cluster.Name, secretValue, fileWriter).Return(errors.New("test"))

	err := installer.GenerateManagementKubeconfig(context.Background(), cluster)
	if err == nil {
		t.Fatal(err)
	}
}

func TestCleanupKubeconfigSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)

	k8s := NewMockKubernetesClient(ctrl)
	writer := filewritermock.NewMockFileWriter(ctrl)
	kwriter := kubeconfigmocks.NewMockWriter(ctrl)

	clusterName := "test-cluster"
	fileName := fmt.Sprintf("%s-aws.kubeconfig", clusterName)

	// Mock Delete to return success (no error)
	writer.EXPECT().Delete(fileName).Return(nil)

	installer := awsiamauth.NewInstaller(k8s, writer, kwriter)

	err := installer.CleanupKubeconfig(clusterName)
	if err != nil {
		t.Fatalf("CleanupKubeconfig failed: %v", err)
	}
}

func TestCleanupKubeconfigRemoveError(t *testing.T) {
	ctrl := gomock.NewController(t)

	k8s := NewMockKubernetesClient(ctrl)
	writer := filewritermock.NewMockFileWriter(ctrl)
	kwriter := kubeconfigmocks.NewMockWriter(ctrl)

	clusterName := "test-cluster"
	fileName := fmt.Sprintf("%s-aws.kubeconfig", clusterName)

	// Mock Delete to return an error
	writer.EXPECT().Delete(fileName).Return(errors.New("permission denied"))

	installer := awsiamauth.NewInstaller(k8s, writer, kwriter)

	err := installer.CleanupKubeconfig(clusterName)
	if err == nil {
		t.Fatal("CleanupKubeconfig should fail when file cannot be removed")
	}

	if !strings.Contains(err.Error(), "removing aws-iam-authenticator kubeconfig") {
		t.Fatalf("Expected error message to contain 'removing aws-iam-authenticator kubeconfig', got: %v", err)
	}
}
