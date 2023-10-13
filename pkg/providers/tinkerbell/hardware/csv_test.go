package hardware_test

import (
	"bytes"
	stdcsv "encoding/csv"
	"errors"
	"fmt"
	"strings"
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

	reader, err := hardware.NewCSVReader(buf.Buffer, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
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

	reader, err := hardware.NewCSVReader(buf.Buffer, nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestCSVReaderFromFile(t *testing.T) {
	g := gomega.NewWithT(t)

	reader, err := hardware.NewNormalizedCSVReaderFromFile("./testdata/hardware.csv", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.Equal(
		hardware.Machine{
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
		},
	))
}

func TestNewCSVReaderWithIOReaderError(t *testing.T) {
	g := gomega.NewWithT(t)

	expect := errors.New("read err")

	_, err := hardware.NewCSVReader(iotest.ErrReader(expect), nil)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestCSVReaderWithoutBMCHeaders(t *testing.T) {
	g := gomega.NewWithT(t)

	reader, err := hardware.NewNormalizedCSVReaderFromFile("./testdata/hardware_no_bmc_headers.csv", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(machine).To(gomega.Equal(
		hardware.Machine{
			Labels:       map[string]string{"type": "cp"},
			Nameservers:  []string{"1.1.1.1"},
			Gateway:      "10.10.10.1",
			Netmask:      "255.255.255.0",
			IPAddress:    "10.10.10.10",
			MACAddress:   "00:00:00:00:00:01",
			Hostname:     "worker1",
			Disk:         "/dev/sda",
			BMCIPAddress: "",
			BMCUsername:  "",
			BMCPassword:  "",
		},
	))
}

func TestCSVReaderWithMissingRequiredColumns(t *testing.T) {
	allHeaders := []string{
		"hostname",
		"ip_address",
		"netmask",
		"gateway",
		"nameservers",
		"mac",
		"disk",
		"labels",
	}

	for i, missing := range allHeaders {
		t.Run(fmt.Sprintf("Missing_%v", missing), func(t *testing.T) {
			// Create the set of included headers based on the current iteration.
			included := make([]string, len(allHeaders))
			copy(included, allHeaders)
			included = append(included[0:i], included[i+1:]...)

			// Create a buffer containing the included headers so the CSV reader can pull them.
			buf := bytes.NewBufferString(fmt.Sprintf("%v", strings.Join(included, ",")))

			g := gomega.NewWithT(t)
			_, err := hardware.NewCSVReader(buf, nil)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring(missing))
		})
	}
}

func TestCSVBuildHardwareYamlFromCSV(t *testing.T) {
	g := gomega.NewWithT(t)

	hardwareYaml, err := hardware.BuildHardwareYAML("./testdata/hardware.csv", nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(hardwareYaml).To(gomega.Equal([]byte(`apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  labels:
    type: cp
  name: worker1
  namespace: eksa-system
spec:
  bmcRef:
    kind: Machine
    name: bmc-worker1
  disks:
  - device: /dev/sda
  interfaces:
  - dhcp:
      arch: x86_64
      hostname: worker1
      ip:
        address: 10.10.10.10
        family: 4
        gateway: 10.10.10.1
        netmask: 255.255.255.0
      lease_time: 4294967294
      mac: "00:00:00:00:00:01"
      name_servers:
      - 1.1.1.1
      uefi: true
    netboot:
      allowPXE: true
      allowWorkflow: true
  metadata:
    facility:
      facility_code: onprem
      plan_slug: c2.medium.x86
    instance:
      allow_pxe: true
      always_pxe: true
      hostname: worker1
      id: "00:00:00:00:00:01"
      ips:
      - address: 10.10.10.10
        family: 4
        gateway: 10.10.10.1
        netmask: 255.255.255.0
        public: true
      operating_system: {}
status: {}
---
apiVersion: bmc.tinkerbell.org/v1alpha1
kind: Machine
metadata:
  name: bmc-worker1
  namespace: eksa-system
spec:
  connection:
    authSecretRef:
      name: bmc-worker1-auth
      namespace: eksa-system
    host: 192.168.0.10
    insecureTLS: true
    port: 0
status: {}
---
apiVersion: v1
data:
  password: YWRtaW4=
  username: QWRtaW4=
kind: Secret
metadata:
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth`)))
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
