package v1alpha1

const (
	// HardwareIDAnnotation is used by the controller to store the
	// ID assigned to the hardware by Tinkerbell for migrated hardware.
	HardwareIDAnnotation = "hardware.tinkerbell.org/id"
)

// TinkID returns the Tinkerbell ID associated with this Hardware.
func (h *Hardware) TinkID() string {
	return h.Annotations[HardwareIDAnnotation]
}

// SetTinkID sets the Tinkerbell ID associated with this Hardware.
func (h *Hardware) SetTinkID(id string) {
	if h.Annotations == nil {
		h.Annotations = make(map[string]string)
	}
	h.Annotations[HardwareIDAnnotation] = id
}
