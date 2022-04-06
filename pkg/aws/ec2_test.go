package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws/mocks"
)

type ec2Test struct {
	*WithT
	ctx    context.Context
	client *client
	ec2    *mocks.MockEC2Client
}

func newEC2Test(t *testing.T) *ec2Test {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	ec2 := mocks.NewMockEC2Client(ctrl)
	return &ec2Test{
		WithT: NewWithT(t),
		ctx:   ctx,
		ec2:   ec2,
		client: &client{
			ec2: ec2,
		},
	}
}

func TestImageExists(t *testing.T) {
	g := newEC2Test(t)
	image := "image-1"
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{image},
	}
	g.ec2.EXPECT().DescribeImages(g.ctx, params).Return(nil, nil)
	got, err := g.client.ImageExists(g.ctx, image)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(true))
}

func TestImageExistsError(t *testing.T) {
	g := newEC2Test(t)
	image := "image-1"
	params := &ec2.DescribeImagesInput{
		ImageIds: []string{image},
	}
	g.ec2.EXPECT().DescribeImages(g.ctx, params).Return(nil, errors.New("error"))
	got, err := g.client.ImageExists(g.ctx, image)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestKeyPairExists(t *testing.T) {
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
	got, err := g.client.KeyPairExists(g.ctx, key)
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
	got, err := g.client.KeyPairExists(g.ctx, key)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestKeyPairExistsError(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{key},
	}
	g.ec2.EXPECT().DescribeKeyPairs(g.ctx, params).Return(nil, errors.New("error"))
	got, err := g.client.KeyPairExists(g.ctx, key)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestCreateEC2KeyPairs(t *testing.T) {
	g := newEC2Test(t)
	key := "default"
	val := "pem"
	params := &ec2.CreateKeyPairInput{
		KeyName: &key,
	}
	out := &ec2.CreateKeyPairOutput{
		KeyMaterial: &val,
	}
	g.ec2.EXPECT().CreateKeyPair(g.ctx, params).Return(out, nil)
	got, err := g.client.CreateEC2KeyPairs(g.ctx, key)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(val))
}
