package hardware_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/mocks"
)

func TestNormalizer(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockMachineReader(ctrl)

	normalizer := hardware.NewNormalizer(reader)

	expect := NewValidMachine()
	expect.MACAddress = "AA:BB:CC:DD:EE:FF"
	reader.EXPECT().Read().Return(expect, (error)(nil))

	machine, err := normalizer.Read()

	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Re-use the expect machine instance and lower-case the MAC.
	expect.MACAddress = "aa:bb:cc:dd:ee:ff"

	g.Expect(machine).To(gomega.Equal(expect))
}

func TestRawNormalizer(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockMachineReader(ctrl)

	normalizer := hardware.NewNormalizer(reader)

	expect := NewValidMachine()
	expect.MACAddress = "AA:BB:CC:DD:EE:FF"
	reader.EXPECT().Read().Return(expect, (error)(nil))

	machine, err := normalizer.Read()

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.Equal(machine))
}

func TestRawNormalizerReadError(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockMachineReader(ctrl)

	normalizer := hardware.NewNormalizer(reader)

	expect := errors.New("foo bar")
	reader.EXPECT().Read().Return(hardware.Machine{}, expect)

	_, err := normalizer.Read()

	g.Expect(err).To(gomega.HaveOccurred())
}
