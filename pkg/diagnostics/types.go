package diagnostics

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type supportBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec supportBundleSpec `json:"spec,omitempty"`
}

type supportBundleSpec struct {
	Collectors []*Collect `json:"collectors,omitempty"`
	Analyzers  []*Analyze `json:"analyzers,omitempty"`
}

type singleOutcome struct {
	When    string `json:"when,omitempty"`
	Message string `json:"message,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type outcome struct {
	Fail *singleOutcome `json:"fail,omitempty"`
	Warn *singleOutcome `json:"warn,omitempty"`
	Pass *singleOutcome `json:"pass,omitempty"`
}
