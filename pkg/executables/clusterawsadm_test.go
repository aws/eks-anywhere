package executables_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

func TestBootstrapIamSuccess(t *testing.T) {
	configFile := "testfile"

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "iam", "create-cloudformation-stack", "--config", configFile).Return(bytes.Buffer{}, nil)
	c := executables.NewClusterawsadm(executable)
	err := c.BootstrapIam(ctx, map[string]string{}, configFile)
	if err != nil {
		t.Fatalf("Clusterawsadm.BootstrapIam() error = %v, want = nil", err)
	}
}

func TestBootstrapIamError(t *testing.T) {
	configFile := "testfile"

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "iam", "create-cloudformation-stack", "--config", configFile).Return(bytes.Buffer{}, errors.New("error from execute with env"))
	c := executables.NewClusterawsadm(executable)
	err := c.BootstrapIam(ctx, map[string]string{}, configFile)
	if err == nil {
		t.Fatalf("Clusterawsadm.BootstrapIam() error = %v, want = not nil", err)
	}
}

func TestBootstrapCredsSuccess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "credentials", "encode-as-profile").Return(bytes.Buffer{}, nil)
	c := executables.NewClusterawsadm(executable)
	_, err := c.BootstrapCreds(ctx, map[string]string{})
	if err != nil {
		t.Fatalf("Clusterawsadm.BootstrapCreds() error = %v, want = nil", err)
	}
}

func TestBootstrapCredsError(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "credentials", "encode-as-profile").Return(bytes.Buffer{}, errors.New("error from execute with env"))
	c := executables.NewClusterawsadm(executable)
	_, err := c.BootstrapCreds(ctx, map[string]string{})
	if err == nil {
		t.Fatalf("Clusterawsadm.BootstrapCreds() error = %v, want = not nil", err)
	}
}

func TestListAccessKeysSuccess(t *testing.T) {
	userName := "user"
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "aws", "iam", "list-access-keys", "--user-name", userName).Return(bytes.Buffer{}, nil)
	c := executables.NewClusterawsadm(executable)
	_, err := c.ListAccessKeys(ctx, userName)
	if err != nil {
		t.Fatalf("Clusterawsadm.ListAccessKeys() error = %v, want nil", err)
	}
}

func TestListAccessKeysError(t *testing.T) {
	userName := "user"
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "aws", "iam", "list-access-keys", "--user-name", userName).Return(bytes.Buffer{}, errors.New("error from execute"))
	c := executables.NewClusterawsadm(executable)
	_, err := c.ListAccessKeys(ctx, userName)
	if err == nil {
		t.Fatalf("Clusterawsadm.ListAccessKeys() error = %v, want not nil", err)
	}
}

func TestDeleteCloudformationStackSuccess(t *testing.T) {
	fileName := "testfile"
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "iam", "delete-cloudformation-stack", "--config", fileName).Return(bytes.Buffer{}, nil)
	c := executables.NewClusterawsadm(executable)
	err := c.DeleteCloudformationStack(ctx, map[string]string{}, fileName)
	if err != nil {
		t.Fatalf("Clusterawsadm.DeleteCloudformationStack() error = %v, want nil", err)
	}
}

func TestDeleteCloudformationStackError(t *testing.T) {
	fileName := "testfile"
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().ExecuteWithEnv(ctx, map[string]string{}, "bootstrap", "iam", "delete-cloudformation-stack", "--config", fileName).Return(bytes.Buffer{}, errors.New("error from execute"))
	c := executables.NewClusterawsadm(executable)
	err := c.DeleteCloudformationStack(ctx, map[string]string{}, fileName)
	if err == nil {
		t.Fatalf("Clusterawsadm.DeleteCloudformationStack() error = %v, want not nil", err)
	}
}
