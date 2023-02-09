package aws_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/aws/mocks"
)

type imdsTest struct {
	*WithT
	ctx    context.Context
	client *aws.Client
	imds   *mocks.MockIMDSClient
}

func newIMDSTest(t *testing.T) *imdsTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	imds := mocks.NewMockIMDSClient(ctrl)
	return &imdsTest{
		WithT:  NewWithT(t),
		ctx:    ctx,
		imds:   imds,
		client: aws.NewClient(aws.WithIMDS(imds)),
	}
}

func TestNewIMDSClient(t *testing.T) {
	_ = aws.NewIMDSClient(awsv2.Config{})
}

func TestBuildIMDS(t *testing.T) {
	g := newIMDSTest(t)
	err := g.client.BuildIMDS(g.ctx)
	g.Expect(err).To(Succeed())
}

func TestEC2InstanceIPIMDSNotInit(t *testing.T) {
	g := newIMDSTest(t)
	g.client = aws.NewClient()

	_, err := g.client.EC2InstanceIP(g.ctx)
	g.Expect(err).To(MatchError(ContainSubstring("imds client is not initialized")))
}

func TestEC2InstanceIP(t *testing.T) {
	g := newIMDSTest(t)
	params := &imds.GetMetadataInput{
		Path: "public-ipv4",
	}
	want := "1.2.3.4"
	out := &imds.GetMetadataOutput{
		Content: io.NopCloser(strings.NewReader(want)),
	}
	g.imds.EXPECT().GetMetadata(g.ctx, params).Return(out, nil)
	got, err := g.client.EC2InstanceIP(g.ctx)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(want))
}

func TestEC2InstanceIPGetMetadataError(t *testing.T) {
	g := newIMDSTest(t)
	params := &imds.GetMetadataInput{
		Path: "public-ipv4",
	}
	g.imds.EXPECT().GetMetadata(g.ctx, params).Return(nil, errors.New("error"))
	_, err := g.client.EC2InstanceIP(g.ctx)
	g.Expect(err).NotTo(Succeed())
}
