package executables_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type dockerContainerTest struct {
	*WithT
	ctx context.Context
	c   *mocks.MockDockerClient
}

func newDockerContainerTest(t *testing.T) *dockerContainerTest {
	ctrl := gomock.NewController(t)
	c := mocks.NewMockDockerClient(ctrl)
	return &dockerContainerTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		c:     c,
	}
}

func TestDockerContainerInit(t *testing.T) {
	g := newDockerContainerTest(t)
	g.c.EXPECT().PullImage(g.ctx, "").Return(nil)
	g.c.EXPECT().Execute(g.ctx, gomock.Any()).Return(bytes.Buffer{}, nil)
	d := executables.NewDockerContainerCustomBinary(g.c)
	g.Expect(d.Init(context.Background())).To(Succeed())
}

func TestDockerContainerInitErrorPullImage(t *testing.T) {
	g := newDockerContainerTest(t)
	g.c.EXPECT().PullImage(g.ctx, "").Return(errors.New("error in pull")).Times(5)
	d := executables.NewDockerContainerCustomBinary(g.c)
	d.Retrier = retrier.NewWithMaxRetries(5, 0)
	g.Expect(d.Init(context.Background())).To(MatchError(ContainSubstring("error in pull")))
}
