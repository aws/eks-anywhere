package executables_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

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

func TestDockerRunBasicSucess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "run", "-d", "-i", "--name", "basic_test", "basic_test:latest")

	if err := d.Run(ctx, "basic_test:latest", "basic_test", []string{}); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDockerRunWithCmdSucess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "run", "-d", "-i", "--name", "basic_test", "basic_test:latest", "foo", "bar")

	if err := d.Run(ctx, "basic_test:latest", "basic_test", []string{"foo", "bar"}); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDockerRunWithFlagsSucess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "run", "-d", "-i", "--flag1", "--flag2", "--name", "basic_test", "basic_test:latest")

	if err := d.Run(ctx, "basic_test:latest", "basic_test", []string{}, "--flag1", "--flag2"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDockerRunWithCmdAndFlagsSucess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "run", "-d", "-i", "--flag1", "--flag2", "--name", "basic_test", "basic_test:latest", "foo", "bar")

	if err := d.Run(ctx, "basic_test:latest", "basic_test", []string{"foo", "bar"}, "--flag1", "--flag2"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDockerRunFailure(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	name := "basic_test"
	image := "basic_test:latest"
	dockerRunError := "docker run error"
	expectedError := fmt.Sprintf("running docker container %s with image %s: %s", name, image, dockerRunError)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "run", "-d", "-i", "--name", name, image).Return(bytes.Buffer{}, errors.New(dockerRunError))

	err := d.Run(ctx, image, name, []string{})
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}

func TestDockerForceRemoveSuccess(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "rm", "-f", "basic_test")

	if err := d.ForceRemove(ctx, "basic_test"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDockerForceRemoveFailure(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	name := "basic_test"
	dockerForceRemoveError := "docker force remove error"
	expectedError := fmt.Sprintf("force removing docker container %s: %s", name, dockerForceRemoveError)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "rm", "-f", name).Return(bytes.Buffer{}, errors.New(dockerForceRemoveError))

	err := d.ForceRemove(ctx, name)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}

func TestDockerCheckContainerExistenceExists(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	name := "basic_test"

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "container", "inspect", name).Return(bytes.Buffer{}, nil)

	exists, err := d.CheckContainerExistence(ctx, name)
	assert.True(t, exists)
	assert.Nil(t, err)
}

func TestDockerCheckContainerExistenceDoesNotExists(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	name := "basic_test"

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "container", "inspect", name).Return(bytes.Buffer{}, fmt.Errorf("Error: No such container: %s", name))

	exists, err := d.CheckContainerExistence(ctx, name)
	assert.False(t, exists)
	assert.Nil(t, err)
}

func TestDockerCheckContainerExistenceOtherError(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	name := "basic_test"
	dockerError := "An unexpected error occured"
	expectedError := fmt.Sprintf("checking if a docker container with name %s exists: %s", name, dockerError)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	d := executables.NewDocker(executable)

	executable.EXPECT().Execute(ctx, "container", "inspect", name).Return(bytes.Buffer{}, errors.New(dockerError))

	exists, err := d.CheckContainerExistence(ctx, name)
	assert.False(t, exists)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}
