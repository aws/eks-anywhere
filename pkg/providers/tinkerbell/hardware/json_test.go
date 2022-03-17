package hardware_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/mocks"
)

func TestTinkerbellHardwareJson(t *testing.T) {
	g := gomega.NewWithT(t)

	buffer := &WriteCloser{}
	writer := hardware.NewTinkerbellHardwareJson(buffer)

	expect := NewValidMachine()
	err := writer.Write(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	var hardware hardware.Hardware
	err = json.Unmarshal(buffer.Buffer[0], &hardware)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	AssertHardwareRepresentsMachine(g, hardware, expect)
}

func TestTinkerbellHardwareJsonMultiWriteErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	writer := hardware.NewTinkerbellHardwareJson(&WriteCloser{})

	err := writer.Write(NewValidMachine())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = writer.Write(NewValidMachine())
	g.Expect(err).To(gomega.Equal(hardware.ErrTinkebellHardwareJsonRepeatWrites))
}

func TestTinkerbellHardwareJsonWriter(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	buffer := &WriteCloser{}
	expect := NewValidMachine()
	path := filepath.Join("foo", expect.Hostname)

	factory := mocks.NewMockTinkerbellHardwareJsonFactory(ctrl)
	factory.EXPECT().
		Create(expect.Hostname).
		Return(hardware.NewTinkerbellHardwareJson(buffer), path, (error)(nil))

	writer := hardware.NewTinkerbellHardwareJsonWriter(factory)

	err := writer.Write(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	var hardware hardware.Hardware
	err = json.Unmarshal(buffer.Buffer[0], &hardware)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	AssertHardwareRepresentsMachine(g, hardware, expect)

	journal := writer.Journal()
	g.Expect(journal).To(gomega.ContainElement(path))
}

func TestTinkerbellHardwareJsonWriterCreateErrors(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	expect := errors.New("foo bar error")

	factory := mocks.NewMockTinkerbellHardwareJsonFactory(ctrl)
	factory.EXPECT().
		Create(gomock.Any()).
		Return(nil, "", expect)

	writer := hardware.NewTinkerbellHardwareJsonWriter(factory)

	err := writer.Write(NewValidMachine())
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func AssertHardwareRepresentsMachine(g *gomega.WithT, h hardware.Hardware, m hardware.Machine) {
	g.Expect(h.Id).To(gomega.Equal(m.Id))
	g.Expect(h.Metadata.Instance.Id).To(gomega.Equal(m.Id))
	g.Expect(h.Metadata.Instance.Hostname).To(gomega.Equal(m.Hostname))
	g.Expect(h.Network.Interfaces[0].Dhcp.Hostname).To(gomega.Equal(m.Hostname))
	g.Expect(h.Network.Interfaces[0].Dhcp.Mac).To(gomega.Equal(m.MacAddress))
	g.Expect(h.Network.Interfaces[0].Dhcp.NameServers).To(gomega.BeEquivalentTo(m.Nameservers))
	g.Expect(h.Network.Interfaces[0].Dhcp.Ip.Address).To(gomega.Equal(m.IpAddress))
	g.Expect(h.Network.Interfaces[0].Dhcp.Ip.Gateway).To(gomega.Equal(m.Gateway))
	g.Expect(h.Network.Interfaces[0].Dhcp.Ip.Netmask).To(gomega.Equal(m.Netmask))
}

type WriteCloser struct {
	Buffer [][]byte
}

func (w *WriteCloser) Write(b []byte) (int, error) {
	w.Buffer = append(w.Buffer, b)
	return len(b), nil
}

func (*WriteCloser) Close() error {
	return nil
}
