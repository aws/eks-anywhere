package hardware

import (
	"fmt"
	"math"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// DiskExtractor represents a hardware labels and map it with the appropriate disk given in the hardware csv file
type DiskExtractor struct {
	selector                 map[string]eksav1alpha1.HardwareSelector
	disks                    map[string]string
	disksProvisionedHardware map[string]string
}

// NewDiskExtractor creates a DiskExtractor instance.
func NewDiskExtractor() *DiskExtractor {
	return &DiskExtractor{
		selector:                 make(map[string]eksav1alpha1.HardwareSelector),
		disks:                    make(map[string]string),
		disksProvisionedHardware: make(map[string]string),
	}
}

// Write matches m to registered hardware selectors and caches the disk for a given selector.
// If subsequent Machine matches are made for a given selector, the disk from the first Machine
// match is used.
func (d *DiskExtractor) Write(m Machine) error {
	for key, selector := range d.selector {
		if _, ok := d.disks[key]; !ok && LabelsMatchSelector(selector, m.Labels) {
			d.disks[key] = m.Disk
		}
	}

	return nil
}

// Register registers selector with d such that a disk can be cached when
// machines are written to Write().
func (d *DiskExtractor) Register(selector eksav1alpha1.HardwareSelector) error {
	key, err := serializeHardwareSelector(selector)
	if err != nil {
		return err
	}

	if _, ok := d.selector[key]; !ok {
		d.selector[key] = selector
	}

	return nil
}

func (d *DiskExtractor) InsertDisks(hardware *tinkv1alpha1.Hardware) error {
	for key, selector := range d.selector {
		if _, ok := d.disks[key]; !ok && LabelsMatchSelector(selector, hardware.Labels) {
			d.disks[key] = hardware.Spec.Disks[0].Device
		}
	}

	return nil
}

func (d *DiskExtractor) InsertProvisionedHardwareDisks(hardware *tinkv1alpha1.Hardware) error {
	for key, selector := range d.selector {
		if _, ok := d.disksProvisionedHardware[key]; !ok && LabelsMatchSelector(selector, hardware.Labels) {
			d.disksProvisionedHardware[key] = hardware.Spec.Disks[0].Device
		}
	}

	return nil
}

// GetDisk returns the disk cached for selector. If selector has no disk cached ErrDiskNotFound
// is returned.
func (d *DiskExtractor) GetDisk(selector eksav1alpha1.HardwareSelector) (string, error) {
	key, err := serializeHardwareSelector(selector)
	if err != nil {
		return "", err
	}

	if _, ok := d.disks[key]; !ok {
		return "", ErrDiskNotFound{key}
	}

	return d.disks[key], nil
}

// GetDiskProvisionedHardware returns the disk cached for selector. If selector has no disk cached ErrDiskNotFound
// is returned.
func (d *DiskExtractor) GetDiskProvisionedHardware(selector eksav1alpha1.HardwareSelector) (string, error) {
	key, err := serializeHardwareSelector(selector)
	if err != nil {
		return "", err
	}

	if _, ok := d.disksProvisionedHardware[key]; !ok {
		return "", ErrDiskNotFound{key}
	}

	return d.disksProvisionedHardware[key], nil
}

// serializeHardwareSelector returns a key for use in a map unique selector.
func serializeHardwareSelector(selector eksav1alpha1.HardwareSelector) (string, error) {
	return selector.ToString()
}

// ErrDiskNotFound indicates a disk was not found for a given selector.
type ErrDiskNotFound struct {
	// A unique identifier for the selector, preferrably something useful to an end-user.
	SelectorID string
}

func (e ErrDiskNotFound) Error() string {
	return fmt.Sprintf("no disk found for hardware selector %v", e.SelectorID)
}

func (ErrDiskNotFound) Is(t error) bool {
	_, ok := t.(ErrDiskNotFound)
	return ok
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

// RemoveHardware removes a specific hardware at a given index from the catalogue.
func (c *Catalogue) RemoveHardware(hardware *tinkv1alpha1.Hardware, index int) error {
	if err := c.hardwareIndex.Remove(hardware); err != nil {
		return err
	}

	if index >= len(c.hardware) {
		return fmt.Errorf("index out of range: %d", index)
	}
	c.hardware[index] = c.hardware[len(c.hardware)-1]
	c.hardware[len(c.hardware)-1] = nil
	c.hardware = c.hardware[:len(c.hardware)-1]

	return nil
}

// RemoveHardwares removes a slice of hardwares from the catalogue.
func (c *Catalogue) RemoveHardwares(hardware []tinkv1alpha1.Hardware) error {
	m := make(map[string]int, len(c.hardware))
	for i, hw := range c.hardware {
		m[hw.Name+":"+hw.Namespace] = i
	}

	for _, hw := range hardware {
		key := hw.Name + ":" + hw.Namespace
		if _, ok := m[key]; ok {
			if err := c.RemoveHardware(c.hardware[m[key]], m[key]); err != nil {
				return err
			}
			delete(m, key)
		}
	}
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
						// set LeaseTime to the max value so it effectively hands out max duration leases (~136 years)
						// This value gets ignored for Ubuntu because we set static IPs for it
						// It's only temporarily needed for Bottlerocket until Bottlerocket supports static IPs
						LeaseTime:   int64(math.Pow(2, 32) - 2),
						Hostname:    m.Hostname,
						NameServers: m.Nameservers,
						UEFI:        true,
						VLANID:      m.VLANID,
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
