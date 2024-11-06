package v1alpha1

const (
	// TemplateIDAnnotation is used by the controller to store the
	// ID assigned to the template by Tinkerbell for migrated templates.
	TemplateIDAnnotation = "template.tinkerbell.org/id"
)

// TinkID returns the Tinkerbell ID associated with this Template.
func (t *Template) TinkID() string {
	return t.Annotations[TemplateIDAnnotation]
}

// SetTinkID sets the Tinkerbell ID associated with this Template.
func (t *Template) SetTinkID(id string) {
	if t.Annotations == nil {
		t.Annotations = make(map[string]string)
	}
	t.Annotations[TemplateIDAnnotation] = id
}
