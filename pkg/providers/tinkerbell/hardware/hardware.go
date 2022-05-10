package hardware

import (
	"fmt"

	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	validationutil "k8s.io/apimachinery/pkg/util/validation"
)

const HardwareProvisioningState = "provisioning"

// HardwareCatalogue stores v1alpha1.Hardware objects. It provides efficient lookup of objects
// based on registered indexes.
type HardwareCatalogue struct {
	store IndexedStore
}

// NewHardwareCatalogue constructs a HardwareCatalogue object for the v1alpha1.Hardware type.
func NewHardwareCatalogue() *HardwareCatalogue {
	catalogue := &HardwareCatalogue{
		store: NewIndexedStore(&v1alpha1.Hardware{}),
	}

	return catalogue
}

// IndexHardware indexes Hardware instances on index by extracfting the key using fn. Objects passed
// to fn will be the same type passed to NewHardwareCatalogue().
func (c *HardwareCatalogue) IndexHardware(index string, fn KeyExtractorFunc) {
	c.store.IndexField(index, fn)
}

// InsertHardware inserts Hardware into the catalogue. If any indexes exist, the hardware is
// indexed.
func (c *HardwareCatalogue) InsertHardware(hardware *v1alpha1.Hardware) error {
	return c.store.Insert(hardware)
}

// AllHardware retrieves a copy of the catalogued Hardware instances.
func (c *HardwareCatalogue) AllHardware() []*v1alpha1.Hardware {
	hardware := make([]*v1alpha1.Hardware, c.store.Size())
	for i, v := range c.store.All() {
		hardware[i] = v.(*v1alpha1.Hardware)
	}
	return hardware
}

// LookupHardware retrieves Hardware instances on index with a key of key. Multiple hardware _may_
// have the same key hence it can return multiple Hardware.
func (c *HardwareCatalogue) LookupHardware(index, key string) ([]*v1alpha1.Hardware, error) {
	untyped, err := c.store.Lookup(index, key)
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
func (c *HardwareCatalogue) TotalHardware() int {
	return c.store.Size()
}

// validateHardware performs static validations on a Tinkerbell Hardware custom resource.
func ValidateHardware(hardware *v1alpha1.Hardware) error {
	if hardware.Name == "" {
		return fmt.Errorf("hardware name is required")
	}

	if errs := validationutil.IsDNS1123Subdomain(hardware.Name); len(errs) > 0 {
		return fmt.Errorf("hardware name not dns1123 compliant: %v: %v", hardware.Name, errs)
	}

	return nil
}

// HardwareAssertion defines a condition the injected Hardware must satisfy.
type HardwareAssertion func(*v1alpha1.Hardware) error

// HardwareValidator executes registered HardwareAssertions against an instance of Tinkberll
// Hardware.
type HardwareValidator []HardwareAssertion

// NewHardwareValidator returns a HardwareValidator instance with default assertions.
func NewHardwareValidator() HardwareValidator {
	var v HardwareValidator
	v.Register(ValidateHardware)
	return v
}

// Register registers assertions with v.
func (v *HardwareValidator) Register(assertions ...HardwareAssertion) {
	*v = append(*v, assertions...)
}

// Validate runs all HardwareAssertions registered with v against hardware.
func (v *HardwareValidator) Validate(hardware *v1alpha1.Hardware) error {
	for _, a := range *v {
		if err := a(hardware); err != nil {
			return err
		}
	}
	return nil
}

type WorkflowIndex interface {
	LookupWorkflow(index string, key string) ([]*v1alpha1.Workflow, error)
}

func WithNoWorkflowsForHardwareAssertion(index WorkflowIndex) HardwareAssertion {
	return func(h *v1alpha1.Hardware) error {
		for _, iface := range h.Spec.Interfaces {
			if iface.DHCP != nil {
				workflowsForMac, err := index.LookupWorkflow(WorkflowHardwareMACIndex, iface.DHCP.MAC)
				if err != nil {
					return err
				}

				if len(workflowsForMac) > 0 {
					return fmt.Errorf(
						"found pre-registered workflows for hardware: hardware=%v; mac=%v",
						h.Name,
						iface.DHCP.MAC,
					)
				}
			}
		}

		return nil
	}
}

func WithHardwareStateProvisioningAssertion() HardwareAssertion {
	return func(hardware *v1alpha1.Hardware) error {
		if hardware.Spec.Metadata == nil {
			return fmt.Errorf("missing hardware metadata: %v", hardware.Name)
		}

		if hardware.Spec.Metadata.State != "provisioning" {
			return fmt.Errorf("hardware metadata state not set to 'provisioning': name=%v", hardware.Name)
		}

		return nil
	}
}

func ValidateCataloguedHardware(validator HardwareValidator, catalogue *HardwareCatalogue) error {
	for _, hardware := range catalogue.AllHardware() {
		if err := validator.Validate(hardware); err != nil {
			return err
		}
	}
	return nil
}
