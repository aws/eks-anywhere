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

func TestCreateAccessKeySuccess(t *testing.T) {
	var userName string
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "iam", "create-access-key", "--user-name", userName).Return(bytes.Buffer{}, nil)
	c := executables.NewAwsCli(executable)
	_, err := c.CreateAccessKey(ctx, userName)
	if err != nil {
		t.Fatalf("Awscli.CreateAccessKey() error = %v, want nil", err)
	}
}

func TestCreateAccessKeyError(t *testing.T) {
	var userName string
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "iam", "create-access-key", "--user-name", userName).Return(bytes.Buffer{}, errors.New("error from execute"))
	c := executables.NewAwsCli(executable)
	_, err := c.CreateAccessKey(ctx, userName)
	if err == nil {
		t.Fatalf("Awscli.CreateAccessKey() error = %v, want not nil", err)
	}
}
