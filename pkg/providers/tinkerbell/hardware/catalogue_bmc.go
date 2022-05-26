package hardware

import (
	"github.com/tinkerbell/rufio/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// IndexBMCs indexes BMC instances on index by extracfting the key using fn.
func (c *Catalogue) IndexBMCs(index string, fn KeyExtractorFunc) {
	c.bmcIndex.IndexField(index, fn)
}

// InsertBMC inserts BMCs into the catalogue. If any indexes exist, the BMC is indexed.
func (c *Catalogue) InsertBMC(bmc *v1alpha1.BaseboardManagement) error {
	if err := c.bmcIndex.Insert(bmc); err != nil {
		return err
	}
	c.bmcs = append(c.bmcs, bmc)
	return nil
}

// AllBMCs retrieves a copy of the catalogued BMC instances.
func (c *Catalogue) AllBMCs() []*v1alpha1.BaseboardManagement {
	bmcs := make([]*v1alpha1.BaseboardManagement, len(c.bmcs))
	copy(bmcs, c.bmcs)
	return bmcs
}

// LookupBMC retrieves BMC instances on index with a key of key. Multiple BMCs _may_
// have the same key hence it can return multiple BMCs.
func (c *Catalogue) LookupBMC(index, key string) ([]*v1alpha1.BaseboardManagement, error) {
	untyped, err := c.bmcIndex.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	bmcs := make([]*v1alpha1.BaseboardManagement, len(untyped))
	for i, v := range untyped {
		bmcs[i] = v.(*v1alpha1.BaseboardManagement)
	}

	return bmcs, nil
}

// TotalBMCs returns the total BMCs registered in the catalogue.
func (c *Catalogue) TotalBMCs() int {
	return len(c.bmcs)
}

const BMCNameIndex = ".ObjectMeta.Name"

// WithBMCNameIndex creates a BMC index using BMCNameIndex on .ObjectMeta.Name.
func WithBMCNameIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexBMCs(BMCNameIndex, func(o interface{}) string {
			bmc := o.(*v1alpha1.BaseboardManagement)
			return bmc.ObjectMeta.Name
		})
	}
}

// BMCCatalogueWriter converts Machine instances to Tinkerbell BaseboardManagement and inserts them
// in a catalogue.
type BMCCatalogueWriter struct {
	catalogue *Catalogue
}

var _ MachineWriter = &BMCCatalogueWriter{}

// NewBMCCatalogueWriter creates a new BMCCatalogueWriter instance.
func NewBMCCatalogueWriter(catalogue *Catalogue) *BMCCatalogueWriter {
	return &BMCCatalogueWriter{catalogue: catalogue}
}

// Write converts m to a Tinkerbell BaseboardManagement and inserts it into w's Catalogue.
func (w *BMCCatalogueWriter) Write(m Machine) error {
	if m.HasBMC() {
		return w.catalogue.InsertBMC(baseboardManagementComputerFromMachine(m))
	}
	return nil
}

func baseboardManagementComputerFromMachine(m Machine) *v1alpha1.BaseboardManagement {
	// TODO(chrisdoherty4)
	// 	- Set the namespace to the CAPT namespace.
	// 	- Patch through insecure TLS.
	return &v1alpha1.BaseboardManagement{
		TypeMeta: newBaseboardManagementTypeMeta(),
		ObjectMeta: v1.ObjectMeta{
			Name:      formatBMCRef(m),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: v1alpha1.BaseboardManagementSpec{
			Connection: v1alpha1.Connection{
				Host: m.BMCIPAddress,
				AuthSecretRef: corev1.SecretReference{
					Name: formatBMCSecretRef(m),
				},
				InsecureTLS: true,
			},
		},
	}
}
