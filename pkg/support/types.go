package supportbundle

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type supportBundle struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec supportBundleSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type supportBundleSpec struct {
	Collectors []*Collect `json:"collectors,omitempty" yaml:"collectors,omitempty"`
	Analyzers  []*Analyze `json:"analyzers,omitempty" yaml:"analyzers,omitempty"`
}

type singleOutcome struct {
	When    string `json:"when,omitempty" yaml:"when,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	URI     string `json:"uri,omitempty" yaml:"uri,omitempty"`
}

type outcome struct {
	Fail *singleOutcome `json:"fail,omitempty" yaml:"fail,omitempty"`
	Warn *singleOutcome `json:"warn,omitempty" yaml:"warn,omitempty"`
	Pass *singleOutcome `json:"pass,omitempty" yaml:"pass,omitempty"`
}
