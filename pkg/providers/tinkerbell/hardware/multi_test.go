package hardware_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/mocks"
)

func TestTeeWriterWritesToAllWriters(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	writer1 := mocks.NewMockMachineWriter(ctrl)
	writer2 := mocks.NewMockMachineWriter(ctrl)

	expect := hardware.Machine{
		Id: "quxer",
	}

	var machine1, machine2 hardware.Machine

	writer1.EXPECT().
		Write(expect).
		DoAndReturn(func(m hardware.Machine) error {
			machine1 = m
			return nil
		})
	writer2.EXPECT().
		Write(expect).
		Do(func(m hardware.Machine) error {
			machine2 = m
			return nil
		}).
		Return((error)(nil))

	tee := hardware.MultiMachineWriter(writer1, writer2)

	err := tee.Write(expect)

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine1).To(gomega.BeEquivalentTo(expect))
	g.Expect(machine2).To(gomega.BeEquivalentTo(expect))
}

func TestTeeWriterFirstWriterErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	writer1 := mocks.NewMockMachineWriter(ctrl)
	writer2 := mocks.NewMockMachineWriter(ctrl)

	machine := hardware.Machine{Id: "qux-foo"}

	expect := errors.New("first writer error")

	writer1.EXPECT().
		Write(machine).
		Return(expect)

	tee := hardware.MultiMachineWriter(writer1, writer2)

	err := tee.Write(machine)

	g.Expect(err).To(gomega.BeEquivalentTo(expect))
}

func TestTeeWriterSecondWriterErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	writer1 := mocks.NewMockMachineWriter(ctrl)
	writer2 := mocks.NewMockMachineWriter(ctrl)

	machine := hardware.Machine{Id: "qux-foo"}

	expect := errors.New("first writer error")

	writer1.EXPECT().
		Write(machine).
		Return((error)(nil))
	writer2.EXPECT().
		Write(machine).
		Return(expect)

	tee := hardware.MultiMachineWriter(writer1, writer2)

	err := tee.Write(machine)

	g.Expect(err).To(gomega.BeEquivalentTo(expect))
}
