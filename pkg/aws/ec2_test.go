package aws_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/aws/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

type ec2Test struct {
	*WithT
	ctx    context.Context
	client *aws.Client
	ec2    *mocks.MockEC2Client
}

func newEC2Test(t *testing.T) *ec2Test {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	ec2 := mocks.NewMockEC2Client(ctrl)
	return &ec2Test{
		WithT:  NewWithT(t),
		ctx:    ctx,
		ec2:    ec2,
		client: aws.NewClient(aws.WithEC2(ec2)),
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

func TestEC2InstanceTypes(t *testing.T) {
	g := newEC2Test(t)
	out := &ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []types.InstanceTypeInfo{
			{
				InstanceType: types.InstanceTypeC1Medium,
				VCpuInfo: &types.VCpuInfo{
					DefaultVCpus: ptr.Int32(8),
				},
			},
			{
				InstanceType: types.InstanceTypeA1Large,
				VCpuInfo: &types.VCpuInfo{
					DefaultVCpus: ptr.Int32(2),
				},
			},
		},
	}
	want := []aws.EC2InstanceType{
		{
			Name:        "c1.medium",
			DefaultVCPU: ptr.Int32(8),
		},
		{
			Name:        "a1.large",
			DefaultVCPU: ptr.Int32(2),
		},
	}
	g.ec2.EXPECT().DescribeInstanceTypes(g.ctx, &ec2.DescribeInstanceTypesInput{}).Return(out, nil)
	got, err := g.client.EC2InstanceTypes(g.ctx)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(want))
}

func TestEC2InstanceTypesError(t *testing.T) {
	g := newEC2Test(t)
	g.ec2.EXPECT().DescribeInstanceTypes(g.ctx, &ec2.DescribeInstanceTypesInput{}).Return(nil, errors.New("describe instance type error"))
	_, err := g.client.EC2InstanceTypes(g.ctx)
	g.Expect(err).To(MatchError(ContainSubstring("describing ec2 instance type in device")))
}
