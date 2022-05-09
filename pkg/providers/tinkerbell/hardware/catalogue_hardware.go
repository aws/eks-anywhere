package hardware

import tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

// IndexHardware indexes Hardware instances on index by extracfting the key using fn.
func (c *Catalogue) IndexHardware(index string, fn KeyExtractorFunc) {
	c.hardwareIndex.IndexField(index, fn)
}

// InsertHardware inserts Hardware into the catalogue. If any indexes exist, the hardware is
// indexed.
func (c *Catalogue) InsertHardware(hardware *tinkv1alpha1.Hardware) error {
	if err := c.hardwareIndex.Insert(hardware); err != nil {
		return err
	}
	c.Hardware = append(c.Hardware, hardware)
	return nil
}

// AllHardware retrieves a copy of the catalogued Hardware instances.
func (c *Catalogue) AllHardware() []*tinkv1alpha1.Hardware {
	hardware := make([]*tinkv1alpha1.Hardware, len(c.Hardware))
	copy(hardware, c.Hardware)
	return hardware
}

// LookupHardware retrieves Hardware instances on index with a key of key. Multiple hardware _may_
// have the same key hence it can return multiple Hardware.
func (c *Catalogue) LookupHardware(index, key string) ([]*tinkv1alpha1.Hardware, error) {
	untyped, err := c.hardwareIndex.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	hardware := make([]*tinkv1alpha1.Hardware, len(untyped))
	for i, v := range untyped {
		hardware[i] = v.(*tinkv1alpha1.Hardware)
	}

	return hardware, nil
}

// TotalHardware returns the total hardware registered in the catalogue.
func (c *Catalogue) TotalHardware() int {
	return len(c.Hardware)
}

const HardwareIDIndex = ".Spec.ID"

// WithHardwareIDIndex creates a Hardware index using HardwareIDIndex on Hardware.Spec.ID values.
func WithHardwareIDIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareIDIndex, func(o interface{}) string {
			hardware := o.(*tinkv1alpha1.Hardware)
			return hardware.Spec.ID
		})
	}
}

const HardwareBMCRefIndex = ".Spec.BmcRef"

// WithHardwareBMCRefIndex creates a Hardware index using HardwareBMCRefIndex on Hardware.Spec.BmcRef.
func WithHardwareBMCRefIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareBMCRefIndex, func(o interface{}) string {
			hardware := o.(*tinkv1alpha1.Hardware)
			return hardware.Spec.BmcRef
		})
	}
}
