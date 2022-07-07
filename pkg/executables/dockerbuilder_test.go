package executables_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
)

func TestDockerExecutableBuilderInit(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	c := mocks.NewMockDockerContainer(ctrl)
	c.EXPECT().Init(ctx)

	d := executables.NewDockerExecutableBuilder(c)

	closer, err := d.Init(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(closer).NotTo(BeNil())
}

func TestDockerExecutableBuilderBuild(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	c := mocks.NewMockDockerContainer(ctrl)
	c.EXPECT().ContainerName()

	d := executables.NewDockerExecutableBuilder(c)

	executable := d.Build("my-binary")
	g.Expect(executable).NotTo(BeNil())
}
