package hardware_test

import (
	"bytes"
	stdcsv "encoding/csv"
	"errors"
	"testing"
	"testing/iotest"

	csv "github.com/gocarina/gocsv"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCSVReaderReads(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()

	expect := NewValidMachine()

	err := csv.MarshalCSV([]hardware.Machine{expect}, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	reader, err := hardware.NewCSVReader(buf.Buffer)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestCSVReaderReadsWithNoIDSpecified(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()

	expect := NewValidMachine()
	expect.ID = ""

	err := csv.MarshalCSV([]hardware.Machine{expect}, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	const uuid = "unique-id"
	reader, err := hardware.NewCSVReaderWithUUIDGenerator(buf.Buffer, func() string { return uuid })
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	expect.ID = uuid // patch the expected machine with the expected uuid
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestCSVReaderWithMultipleLabels(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()

	expect := NewValidMachine()
	expect.Labels["foo"] = "bar"
	expect.Labels["qux"] = "baz"

	err := csv.MarshalCSV([]hardware.Machine{expect}, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	const uuid = "unique-id"
	reader, err := hardware.NewCSVReaderWithUUIDGenerator(buf.Buffer, func() string { return uuid })
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestCSVReaderFromFile(t *testing.T) {
	g := gomega.NewWithT(t)

	reader, err := hardware.NewCSVReaderFromFile("./testdata/hardware.csv")
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.Equal(
		hardware.Machine{
			ID:           "worker1",
			Labels:       map[string]string{"type": "cp"},
			Nameservers:  []string{"1.1.1.1"},
			Gateway:      "10.10.10.1",
			Netmask:      "255.255.255.0",
			IPAddress:    "10.10.10.10",
			MACAddress:   "00:00:00:00:00:01",
			Hostname:     "worker1",
			Disk:         "/dev/sda",
			BMCIPAddress: "192.168.0.10",
			BMCUsername:  "Admin",
			BMCPassword:  "admin",
			BMCVendor:    "HP",
		},
	))
}

func TestNewCSVReaderWithIOReaderError(t *testing.T) {
	g := gomega.NewWithT(t)

	expect := errors.New("read err")

	_, err := hardware.NewCSVReader(iotest.ErrReader(expect))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

// BufferedCSV is an in-memory CSV that satisfies io.Reader and io.Writer.
type BufferedCSV struct {
	*bytes.Buffer
	*stdcsv.Writer
	*stdcsv.Reader
}

func NewBufferedCSV() *BufferedCSV {
	buf := &BufferedCSV{Buffer: &bytes.Buffer{}}
	buf.Writer = stdcsv.NewWriter(buf.Buffer)
	buf.Reader = stdcsv.NewReader(buf.Buffer)
	return buf
}

// Write writes record to b using the underlying csv.Writer but immediately flushes. This
// ensures the in-memory buffer is always up-to-date.
func (b *BufferedCSV) Write(record []string) error {
	if err := b.Writer.Write(record); err != nil {
		return err
	}
	b.Flush()
	return nil
}
