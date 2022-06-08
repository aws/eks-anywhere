package hardware

import (
	"fmt"
	"time"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// DiskExtractor represents a hardware labels and map it with the appropriate disk given in the hardware csv file
type DiskExtractor struct {
	Selector map[string]eksav1alpha1.HardwareSelector
	Disks    map[string]string
}

func NewDiskExtractor() *DiskExtractor {
	return &DiskExtractor{
		Selector: make(map[string]eksav1alpha1.HardwareSelector),
		Disks:    make(map[string]string),
	}
}

func (d *DiskExtractor) Write(m Machine) error {
	labelKey := m.Labels.String()
	if _, ok := d.Selector[labelKey]; ok {
		d.Disks[labelKey] = m.Disk
	}
	return nil
}

// RegisterHardwareSelector adds the key into DiskExtractor for each machine config type
func (d *DiskExtractor) RegisterHardwareSelector(hardwareSelector eksav1alpha1.HardwareSelector) {
	key := serializedHardwareSelector(hardwareSelector)
	if _, ok := d.Selector[key]; !ok {
		d.Selector[key] = hardwareSelector
	}
}

// serializedHardwareSelector returns a string by serializing all the hardware selectors given per machine config
func serializedHardwareSelector(hardwareSelector eksav1alpha1.HardwareSelector) string {
	selectorKey := make(Labels)
	for key, val := range hardwareSelector {
		selectorKey[key] = val
	}
	return selectorKey.String()
}

// GetDisk returns a disk associated with hardwareSelector of the machine config
func (d *DiskExtractor) GetDisk(hardwareSelector eksav1alpha1.HardwareSelector) (string, error) {
	key := serializedHardwareSelector(hardwareSelector)
	if _, ok := d.Disks[key]; !ok {
		return "", fmt.Errorf("hardware selector does not match with any given hardware machines: %v", hardwareSelector)
	}
	return d.Disks[key], nil
}

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
	c.hardware = append(c.hardware, hardware)
	return nil
}

// AllHardware retrieves a copy of the catalogued Hardware instances.
func (c *Catalogue) AllHardware() []*tinkv1alpha1.Hardware {
	hardware := make([]*tinkv1alpha1.Hardware, len(c.hardware))
	copy(hardware, c.hardware)
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
	return len(c.hardware)
}

const HardwareIDIndex = ".Spec.Metadata.Instance.ID"

// WithHardwareIDIndex creates a Hardware index using HardwareIDIndex on .Spec.Metadata.Instance.ID
// values.
func WithHardwareIDIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareIDIndex, func(o interface{}) string {
			hardware := o.(*tinkv1alpha1.Hardware)
			return hardware.Spec.Metadata.Instance.ID
		})
	}
}

const HardwareBMCRefIndex = ".Spec.BmcRef"

// WithHardwareBMCRefIndex creates a Hardware index using HardwareBMCRefIndex on .Spec.BmcRef.
func WithHardwareBMCRefIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexHardware(HardwareBMCRefIndex, func(o interface{}) string {
			hardware := o.(*tinkv1alpha1.Hardware)
			return hardware.Spec.BMCRef.String()
		})
	}
}

// HardwareCatalogueWriter converts Machine instances to Tinkerbell Hardware and inserts them
// in a catalogue.
type HardwareCatalogueWriter struct {
	catalogue *Catalogue
}

var _ MachineWriter = &HardwareCatalogueWriter{}

// NewHardwareCatalogueWriter creates a new HardwareCatalogueWriter instance.
func NewHardwareCatalogueWriter(catalogue *Catalogue) *HardwareCatalogueWriter {
	return &HardwareCatalogueWriter{catalogue: catalogue}
}

// Write converts m to a Tinkerbell Hardware and inserts it into w's Catalogue.
func (w *HardwareCatalogueWriter) Write(m Machine) error {
	return w.catalogue.InsertHardware(hardwareFromMachine(m))
}

func hardwareFromMachine(m Machine) *tinkv1alpha1.Hardware {
	// allow is necessary to allocate memory so we can get a bool pointer required by
	// the hardware.
	allow := true

	// TODO(chrisdoherty4) Set the namespace to the CAPT namespace.
	return &tinkv1alpha1.Hardware{
		TypeMeta: newHardwareTypeMeta(),
		ObjectMeta: v1.ObjectMeta{
			Name:      m.Hostname,
			Namespace: constants.EksaSystemNamespace,
			Labels:    m.Labels,
		},
		Spec: tinkv1alpha1.HardwareSpec{
			BMCRef: newBMCRefFromMachine(m),
			Disks:  []tinkv1alpha1.Disk{{Device: m.Disk}},
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Facility: &tinkv1alpha1.MetadataFacility{
					FacilityCode: "onprem",
					PlanSlug:     "c2.medium.x86",
				},
				Instance: &tinkv1alpha1.MetadataInstance{
					ID:       m.MACAddress,
					Hostname: m.Hostname,
					Ips: []*tinkv1alpha1.MetadataInstanceIP{
						{
							Address: m.IPAddress,
							Netmask: m.Netmask,
							Gateway: m.Gateway,
							Family:  4,
							Public:  true,
						},
					},
					// TODO(chrisdoherty4) Fix upstream. The OperatingSystem is used in boots to
					// detect what iPXE scripts should be served. The Kubernetes back-end nilifies
					// its response to retrieving the OS data and the handling code doesn't check
					// for nil resulting in a segfault.
					//
					// Upstream needs patching but this will suffice for now.
					OperatingSystem: &tinkv1alpha1.MetadataInstanceOperatingSystem{},
					AllowPxe:        true,
					AlwaysPxe:       true,
				},
				State: "provisioning",
			},
			Interfaces: []tinkv1alpha1.Interface{
				{
					Netboot: &tinkv1alpha1.Netboot{
						AllowPXE:      &allow,
						AllowWorkflow: &allow,
					},
					DHCP: &tinkv1alpha1.DHCP{
						Arch: "x86_64",
						MAC:  m.MACAddress,
						IP: &tinkv1alpha1.IP{
							Address: m.IPAddress,
							Netmask: m.Netmask,
							Gateway: m.Gateway,
							Family:  4,
						},
						// set LeaseTime to 1 month while we figure out how to set static IPs
						LeaseTime:   int64(time.Hour * 24 * 30 / time.Second),
						Hostname:    m.Hostname,
						NameServers: m.Nameservers,
						UEFI:        true,
					},
				},
			},
		},
	}
}

// newBMCRefFromMachine returns a BMCRef pointer for Hardware.
func newBMCRefFromMachine(m Machine) *corev1.TypedLocalObjectReference {
	if m.HasBMC() {
		return &corev1.TypedLocalObjectReference{
			Name: formatBMCRef(m),
			Kind: tinkerbellBMCKind,
		}
	}

	return nil
}
