package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestStore_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewIndexedStore(&T{})

	err := catalogue.Insert(&T{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.Size()).To(gomega.Equal(1))
}

func TestStore_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewIndexedStore(&T{})

	_, err := catalogue.Lookup(hardware.HardwareIDIndex, "ID")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestStore_Indexed(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewIndexedStore(&T{})
	catalogue.IndexField(TIndex, TIndexerFunc)

	expect := &T{S: "foo"}
	err := catalogue.Insert(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.Lookup(TIndex, expect.S)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestStore_AllHardwareReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewIndexedStore(&T{})

	const totalHardware = 1
	err := catalogue.Insert(&T{S: "foo"})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.All()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &T{S: "bar"}

	unchangedHardware := catalogue.All()
	g.Expect(unchangedHardware).To(gomega.Equal(changedHardware))
}

// T is a testing struct for easy indexing.
type T struct {
	S string
}

// TIndex is an index key
const TIndex = "IndexerFunc"

// IndexerFunc is an index fpr TIndex
func TIndexerFunc(o interface{}) []string {
	return []string{o.(*T).S}
}
