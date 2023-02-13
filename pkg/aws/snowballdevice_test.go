package aws_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice"
	"github.com/aws/eks-anywhere/internal/aws-sdk-go-v2/service/snowballdevice/types"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/aws/mocks"
)

type snowballDeviceTest struct {
	*WithT
	ctx            context.Context
	client         *aws.Client
	snowballDevice *mocks.MockSnowballDeviceClient
}

func newSnowballDeviceTest(t *testing.T) *snowballDeviceTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	sbd := mocks.NewMockSnowballDeviceClient(ctrl)
	return &snowballDeviceTest{
		WithT:          NewWithT(t),
		ctx:            ctx,
		snowballDevice: sbd,
		client:         aws.NewClient(aws.WithSnowballDevice(sbd)),
	}
}

func TestIsSnowballDeviceUnlockedSuccess(t *testing.T) {
	g := newSnowballDeviceTest(t)
	out := &snowballdevice.DescribeDeviceOutput{
		UnlockStatus: &types.UnlockStatus{
			State: "UNLOCKED",
		},
	}
	g.snowballDevice.EXPECT().DescribeDevice(g.ctx, nil).Return(out, nil)
	got, err := g.client.IsSnowballDeviceUnlocked(g.ctx)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(true))
}

func TestIsSnowballDeviceUnlockedDescribeDeviceError(t *testing.T) {
	g := newSnowballDeviceTest(t)
	g.snowballDevice.EXPECT().DescribeDevice(g.ctx, nil).Return(nil, errors.New("error"))
	got, err := g.client.IsSnowballDeviceUnlocked(g.ctx)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestIsSnowballDeviceUnlockedDeviceLocked(t *testing.T) {
	g := newSnowballDeviceTest(t)
	out := &snowballdevice.DescribeDeviceOutput{
		UnlockStatus: &types.UnlockStatus{
			State: "LOCKED",
		},
	}
	g.snowballDevice.EXPECT().DescribeDevice(g.ctx, nil).Return(out, nil)
	got, err := g.client.IsSnowballDeviceUnlocked(g.ctx)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(false))
}

func TestSnowballDeviceSoftwareVersionSuccess(t *testing.T) {
	g := newSnowballDeviceTest(t)
	version := "100"
	out := &snowballdevice.DescribeDeviceSoftwareOutput{
		InstalledVersion: &version,
	}
	g.snowballDevice.EXPECT().DescribeDeviceSoftware(g.ctx, nil).Return(out, nil)
	got, err := g.client.SnowballDeviceSoftwareVersion(g.ctx)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(version))
}

func TestSnowballDeviceSoftwareVersionDescribeDeviceSoftwareError(t *testing.T) {
	g := newSnowballDeviceTest(t)
	g.snowballDevice.EXPECT().DescribeDeviceSoftware(g.ctx, nil).Return(nil, errors.New("error"))
	got, err := g.client.SnowballDeviceSoftwareVersion(g.ctx)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(Equal(""))
}
