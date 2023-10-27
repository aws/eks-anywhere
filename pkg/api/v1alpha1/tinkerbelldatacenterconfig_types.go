package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// TinkerbellDatacenterConfigSpec defines the desired state of TinkerbellDatacenterConfig.
type TinkerbellDatacenterConfigSpec struct {
	// TinkerbellIP is used to configure a VIP for hosting the Tinkerbell services.
	TinkerbellIP string `json:"tinkerbellIP"`
	// OSImageURL can be used to override the default OS image path to pull from a local server.
	// OSImageURL is a URL to the OS image used during provisioning. To perform modular upgrades
	// the OSImageURL must be specified on the TinkerbellMachineConfig objects. You cannot specify
	// an OSImageURL on the TinkerbellDatacenterConfig and TinkerbellMachineConfigs simultaneously.
	// It must include the Kubernetes version(s). For example, a URL used for Kubernetes 1.27 could
	// be http://localhost:8080/ubuntu-2204-1.27.tgz
	//+optional
	OSImageURL string `json:"osImageURL,omitempty"`
	// HookImagesURLPath can be used to override the default Hook images path to pull from a local server.
	HookImagesURLPath string `json:"hookImagesURLPath,omitempty"`
	// SkipLoadBalancerDeployment when set to "true" can be used to skip deploying a load balancer to expose Tinkerbell stack.
	// Users will need to deploy and configure a load balancer manually after the cluster is created.
	SkipLoadBalancerDeployment bool `json:"skipLoadBalancerDeployment,omitempty"`
}

// TinkerbellDatacenterConfigStatus defines the observed state of TinkerbellDatacenterConfig
//
// Important: Run "make generate" to regenerate code after modifying this file.
type TinkerbellDatacenterConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TinkerbellDatacenterConfig is the Schema for the TinkerbellDatacenterConfigs API.
type TinkerbellDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinkerbellDatacenterConfigSpec   `json:"spec,omitempty"`
	Status TinkerbellDatacenterConfigStatus `json:"status,omitempty"`
}

func (t *TinkerbellDatacenterConfig) Kind() string {
	return t.TypeMeta.Kind
}

func (v *TinkerbellDatacenterConfig) ExpectedKind() string {
	return TinkerbellDatacenterKind
}

func (t *TinkerbellDatacenterConfig) PauseReconcile() {
	if t.Annotations == nil {
		t.Annotations = map[string]string{}
	}
	t.Annotations[pausedAnnotation] = "true"
}

func (t *TinkerbellDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := t.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (t *TinkerbellDatacenterConfig) ClearPauseAnnotation() {
	if t.Annotations != nil {
		delete(t.Annotations, pausedAnnotation)
	}
}

// Validate validates the Tinkerbell datacenter config.
func (t *TinkerbellDatacenterConfig) Validate() error {
	return validateDatacenterConfig(t)
}

func (t *TinkerbellDatacenterConfig) ConvertConfigToConfigGenerateStruct() *TinkerbellDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if t.Namespace != "" {
		namespace = t.Namespace
	}
	config := &TinkerbellDatacenterConfigGenerate{
		TypeMeta: t.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        t.Name,
			Annotations: t.Annotations,
			Namespace:   namespace,
		},
		Spec: t.Spec,
	}

	return config
}

func (t *TinkerbellDatacenterConfig) Marshallable() Marshallable {
	return t.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as TinkerbellDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig.
type TinkerbellDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec TinkerbellDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// TinkerbellDatacenterConfigList contains a list of TinkerbellDatacenterConfig.
type TinkerbellDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TinkerbellDatacenterConfig{}, &TinkerbellDatacenterConfigList{})
}
