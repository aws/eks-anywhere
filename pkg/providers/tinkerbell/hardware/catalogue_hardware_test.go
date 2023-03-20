package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCatalogue_Hardware_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardwares_Remove(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw1",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw2",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.RemoveHardwares([]v1alpha1.Hardware{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw2",
				Namespace: "namespace",
			},
		},
	})).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardwares_RemoveDuplicates(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw1",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw2",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw2",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.RemoveHardwares([]v1alpha1.Hardware{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw2",
				Namespace: "namespace",
			},
		},
	})).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardwares_RemoveExtraHw(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw1",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw2",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw3",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.RemoveHardwares([]v1alpha1.Hardware{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw2",
				Namespace: "namespace",
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw3",
				Namespace: "namespace",
			},
		},
	})).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardwares_RemoveNothing(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw1",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.RemoveHardwares([]v1alpha1.Hardware{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw2",
				Namespace: "namespace",
			},
		},
	})).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardwares_RemoveEverything(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hw1",
			Namespace: "namespace",
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.RemoveHardwares([]v1alpha1.Hardware{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hw1",
				Namespace: "namespace",
			},
		},
	})).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(0))
}

func TestCatalogue_Hardware_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	_, err := catalogue.LookupHardware(hardware.HardwareIDIndex, "ID")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestCatalogue_Hardware_IDIndex(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const id = "hello"
	expect := &v1alpha1.Hardware{
		Spec: v1alpha1.HardwareSpec{
			Metadata: &v1alpha1.HardwareMetadata{
				Instance: &v1alpha1.MetadataInstance{
					ID: id,
				},
			},
		},
	}
	err := catalogue.InsertHardware(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupHardware(hardware.HardwareIDIndex, id)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_Hardware_BmcRefIndex(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareBMCRefIndex())

	group := "foobar"
	ref := &corev1.TypedLocalObjectReference{
		APIGroup: &group,
		Kind:     "bazqux",
		Name:     "secret",
	}
	expect := &v1alpha1.Hardware{Spec: v1alpha1.HardwareSpec{BMCRef: ref}}
	err := catalogue.InsertHardware(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupHardware(hardware.HardwareBMCRefIndex, ref.String())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_Hardware_AllHardwareReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const totalHardware = 1
	hw := &v1alpha1.Hardware{
		Spec: v1alpha1.HardwareSpec{
			Metadata: &v1alpha1.HardwareMetadata{
				Instance: &v1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	}
	err := catalogue.InsertHardware(hw)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllHardware()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &v1alpha1.Hardware{
		Spec: v1alpha1.HardwareSpec{
			Metadata: &v1alpha1.HardwareMetadata{
				Instance: &v1alpha1.MetadataInstance{
					ID: "qux",
				},
			},
		},
	}

	unchangedHardware := catalogue.AllHardware()
	g.Expect(unchangedHardware).ToNot(gomega.Equal(changedHardware))
}

func TestHardwareCatalogueWriter_Write(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	writer := hardware.NewHardwareCatalogueWriter(catalogue)
	machine := NewValidMachine()

	err := writer.Write(machine)
	g.Expect(err).To(gomega.Succeed())

	hardware := catalogue.AllHardware()
	g.Expect(hardware).To(gomega.HaveLen(1))
	g.Expect(hardware[0].Name).To(gomega.Equal(machine.Hostname))
}
