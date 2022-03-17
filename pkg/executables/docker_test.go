package executables_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

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
	executable.EXPECT().Execute(ctx, "version", "--format", "{{.Client.Version}}").Return(*bytes.NewBufferString(version), nil)
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

func TestDockerCgroupVersion(t *testing.T) {
	version := "'\"1\"'\n"
	wantVersion := 1

	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "info", "--format", "'{{json .CgroupVersion}}'").Return(*bytes.NewBufferString(version), nil)
	d := executables.NewDocker(executable)
	cgroupVersion, err := d.CgroupVersion(ctx)
	if err != nil {
		t.Fatalf("Docker.AllocatedMemory() error = %v, want %v", err, cgroupVersion)
	}
	if !reflect.DeepEqual(cgroupVersion, wantVersion) {
		t.Fatalf("Docker.Version() version = %v, want %v", cgroupVersion, wantVersion)
	}
}

func TestDockerLoadFromFile(t *testing.T) {
	file := "file"

	g := NewWithT(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "load", "-i", file).Return(bytes.Buffer{}, nil)
	d := executables.NewDocker(executable)

	g.Expect(d.LoadFromFile(ctx, file)).To(Succeed())
}

func TestDockerSaveToFileMultipleImages(t *testing.T) {
	file := "file"
	image1 := "image1:tag1"
	image2 := "image2:tag2"
	image3 := "image3:tag3"

	g := NewWithT(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "save", "-o", file, image1, image2, image3).Return(bytes.Buffer{}, nil)
	d := executables.NewDocker(executable)

	g.Expect(d.SaveToFile(ctx, file, image1, image2, image3)).To(Succeed())
}

func TestDockerSaveToFileNoImages(t *testing.T) {
	file := "file"

	g := NewWithT(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	executable.EXPECT().Execute(ctx, "save", "-o", file).Return(bytes.Buffer{}, nil)
	d := executables.NewDocker(executable)

	g.Expect(d.SaveToFile(ctx, file)).To(Succeed())
}
