package hardware

import pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"

// IndexBMCs indexes BMC instances on index by extracfting the key using fn.
func (c *Catalogue) IndexBMCs(index string, fn KeyExtractorFunc) {
	c.bmcIndex.IndexField(index, fn)
}

// InsertBMC inserts BMCs into the catalogue. If any indexes exist, the BMC is indexed.
func (c *Catalogue) InsertBMC(bmc *pbnjv1alpha1.BMC) error {
	if err := c.bmcIndex.Insert(bmc); err != nil {
		return err
	}
	c.BMCs = append(c.BMCs, bmc)
	return nil
}

// AllBMCs retrieves a copy of the catalogued BMC instances.
func (c *Catalogue) AllBMCs() []*pbnjv1alpha1.BMC {
	bmcs := make([]*pbnjv1alpha1.BMC, len(c.BMCs))
	copy(bmcs, c.BMCs)
	return bmcs
}

// LookupBMC retrieves BMC instances on index with a key of key. Multiple BMCs _may_
// have the same key hence it can return multiple BMCs.
func (c *Catalogue) LookupBMC(index, key string) ([]*pbnjv1alpha1.BMC, error) {
	untyped, err := c.bmcIndex.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	bmcs := make([]*pbnjv1alpha1.BMC, len(untyped))
	for i, v := range untyped {
		bmcs[i] = v.(*pbnjv1alpha1.BMC)
	}

	return bmcs, nil
}

// TotalBMCs returns the total BMCs registered in the catalogue.
func (c *Catalogue) TotalBMCs() int {
	return len(c.BMCs)
}

const BMCNameIndex = ".ObjectMeta.Name"

// WithBMCNameIndex creates a BMC index using BMCNameIndex on BMC.ObjectMeta.Name.
func WithBMCNameIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexBMCs(BMCNameIndex, func(o interface{}) string {
			bmc := o.(*pbnjv1alpha1.BMC)
			return bmc.ObjectMeta.Name
		})
	}
}
