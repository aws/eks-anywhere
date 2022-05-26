package hardware

import (
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
)

// IndexHardware indexes Hardware instances on index by extracfting the key using fn.
func (c *Catalogue) IndexHardware(index string, fn KeyExtractorFunc) {
	c.hardwareIndex.IndexField(index, fn)
}

// InsertHardware inserts Hardware into the catalogue. If any indexes exist, the hardware is
// indexed.
func (c *Catalogue) InsertHardware(hardware *v1alpha1.Hardware) error {
	if err := c.hardwareIndex.Insert(hardware); err != nil {
		return err
	}
	c.hardware = append(c.hardware, hardware)
	return nil
}

// AllHardware retrieves a copy of the catalogued Hardware instances.
func (c *Catalogue) AllHardware() []*v1alpha1.Hardware {
	hardware := make([]*v1alpha1.Hardware, len(c.hardware))
	copy(hardware, c.hardware)
	return hardware
}

// LookupHardware retrieves Hardware instances on index with a key of key. Multiple hardware _may_
// have the same key hence it can return multiple Hardware.
func (c *Catalogue) LookupHardware(index, key string) ([]*v1alpha1.Hardware, error) {
	untyped, err := c.hardwareIndex.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	hardware := make([]*v1alpha1.Hardware, len(untyped))
	for i, v := range untyped {
		hardware[i] = v.(*v1alpha1.Hardware)
	}

	return hardware, nil
}

// TotalHardware returns the total hardware registered in the catalogue.
func (c *Catalogue) TotalHardware() int {
	return len(c.hardware)
}

const HardwareIDIndex = ".Spec.Metadata.Instance.ID"

// WithHardwareIDIndex creates a Hardware index using HardwareIDIndex on .Spec.Metadata.Instance.ID
// values.
func WithHardwareIDIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareIDIndex, func(o interface{}) string {
			hardware := o.(*v1alpha1.Hardware)
			return hardware.Spec.Metadata.Instance.ID
		})
	}
}

const HardwareBMCRefIndex = ".Spec.BmcRef"

// WithHardwareBMCRefIndex creates a Hardware index using HardwareBMCRefIndex on .Spec.BmcRef.
func WithHardwareBMCRefIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareBMCRefIndex, func(o interface{}) string {
			hardware := o.(*v1alpha1.Hardware)
			return hardware.Spec.BMCRef.String()
		})
	}
}
