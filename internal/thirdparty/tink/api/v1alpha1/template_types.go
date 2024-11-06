package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemplateState represents the template state.
type TemplateState string

const (
	// TemplateError represents a template that is in an error state.
	TemplateError = TemplateState("Error")

	// TemplateReady represents a template that is in a ready state.
	TemplateReady = TemplateState("Ready")
)

// TemplateSpec defines the desired state of Template.
type TemplateSpec struct {
	// +optional
	Data *string `json:"data,omitempty"`
}

// TemplateStatus defines the observed state of Template.
type TemplateStatus struct {
	State TemplateState `json:"state,omitempty"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=templates,scope=Namespaced,categories=tinkerbell,shortName=tpl,singular=template
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// Template is the Schema for the Templates API.
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Templates.
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
