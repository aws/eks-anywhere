package aws_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/aws/mocks"
)

type ec2Test struct {
	*WithT
	ctx    context.Context
	client *aws.Client
	ec2    *mocks.MockEC2Client
	imds   *mocks.MockIMDSClient
}

func newEC2Test(t *testing.T) *ec2Test {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	ec2 := mocks.NewMockEC2Client(ctrl)
	imds := mocks.NewMockIMDSClient(ctrl)
	ec2Client := &aws.EC2{
		EC2Client:  ec2,
		IMDSClient: imds,
	}
	return &ec2Test{
		WithT:  NewWithT(t),
		ctx:    ctx,
		ec2:    ec2,
		imds:   imds,
		client: aws.NewClient(aws.WithEC2(ec2Client)),
	}
}

func TestEC2ImageExists(t *testing.T) {
	g := newEC2Test(t)
	image := "image-1"
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{image},
	}
	g.ec2.EXPECT().DescribeImages(g.ctx, params).Return(nil, nil)
	got, err := g.client.EC2ImageExists(g.ctx, image)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(true))
}

func TestEC2ImageExistsError(t *testing.T) {
	g := newEC2Test(t)
	image := "image-1"
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{image},
	}
	g.ec2.EXPECT().DescribeImages(g.ctx, params).Return(nil, errors.New("error"))
	got, err := g.client.EC2ImageExists(g.ctx, image)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestEC2KeyNameExists(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{key},
	}
	out := &ec2.DescribeKeyPairsOutput{
		KeyPairs: []types.KeyPairInfo{
			{
				KeyName: &key,
			},
		},
	}
	g.ec2.EXPECT().DescribeKeyPairs(g.ctx, params).Return(out, nil)
	got, err := g.client.EC2KeyNameExists(g.ctx, key)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(true))
}

func TestKeyPairNotExists(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{key},
	}
	out := &ec2.DescribeKeyPairsOutput{
		KeyPairs: []types.KeyPairInfo{},
	}
	g.ec2.EXPECT().DescribeKeyPairs(g.ctx, params).Return(out, nil)
	got, err := g.client.EC2KeyNameExists(g.ctx, key)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestEC2KeyNameExistsError(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{key},
	}
	g.ec2.EXPECT().DescribeKeyPairs(g.ctx, params).Return(nil, errors.New("error"))
	got, err := g.client.EC2KeyNameExists(g.ctx, key)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestEC2ImportKeyPair(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	val := []byte("pem")
	params := &ec2.ImportKeyPairInput{
		KeyName:           &key,
		PublicKeyMaterial: []byte(val),
	}
	out := &ec2.ImportKeyPairOutput{}
	g.ec2.EXPECT().ImportKeyPair(g.ctx, params).Return(out, nil)
	err := g.client.EC2ImportKeyPair(g.ctx, key, val)
	g.Expect(err).To(Succeed())
}

func TestEC2InstanceIP(t *testing.T) {
	g := newEC2Test(t)
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
	g := newEC2Test(t)
	params := &imds.GetMetadataInput{
		Path: "public-ipv4",
	}
	g.imds.EXPECT().GetMetadata(g.ctx, params).Return(nil, errors.New("error"))
	_, err := g.client.EC2InstanceIP(g.ctx)
	g.Expect(err).NotTo(Succeed())
}
