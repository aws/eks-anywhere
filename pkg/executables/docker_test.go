package executables_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

func TestGetDockerLBPort(t *testing.T) {
	clusterName := "clusterName"
	wantPort := "test:port"
	clusterLBName := fmt.Sprintf("%s-lb", clusterName)

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, []string{"port", clusterLBName, "6443/tcp"}).Return(*bytes.NewBufferString(wantPort), nil)
	d := executables.NewDocker(executable)
	_, err := d.GetDockerLBPort(ctx, clusterName)
	if err != nil {
		t.Fatalf("Docker.GetDockerLBPort() error = %v, want nil", err)
	}
}

func TestDockerPullImage(t *testing.T) {
	image := "test_image"

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "pull", image).Return(bytes.Buffer{}, nil)
	d := executables.NewDocker(executable)
	err := d.PullImage(ctx, image)
	if err != nil {
		t.Fatalf("Docker.PullImage() error = %v, want nil", err)
	}
}

func TestDockerVersion(t *testing.T) {
	version := "1.234"
	wantVersion := 1

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "version", "--format", "{{.client.Version}}").Return(*bytes.NewBufferString(version), nil)
	d := executables.NewDocker(executable)
	v, err := d.Version(ctx)
	if err != nil {
		t.Fatalf("Docker.Version() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(v, wantVersion) {
		t.Fatalf("Docker.Version() version = %v, want %v", v, wantVersion)
	}
}

func TestDockerAllocatedMemory(t *testing.T) {
	memory := "12345"

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "info", "--format", "'{{json .MemTotal}}'").Return(*bytes.NewBufferString(memory), nil)
	d := executables.NewDocker(executable)
	mem, err := d.AllocatedMemory(ctx)
	if err != nil {
		t.Fatalf("Docker.AllocatedMemory() error = %v, want %v", err, mem)
	}
}
