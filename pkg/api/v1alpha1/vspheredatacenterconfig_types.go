package v1alpha1

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VSphereDatacenterConfigSpec defines the desired state of VSphereDatacenterConfig.
type VSphereDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	Datacenter string `json:"datacenter"`
	Network    string `json:"network"`
	Server     string `json:"server"`
	Thumbprint string `json:"thumbprint"`
	Insecure   bool   `json:"insecure"`
}

// VSphereDatacenterConfigStatus defines the observed state of VSphereDatacenterConfig.
type VSphereDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
	// SpecValid is set to true if vspheredatacenterconfig is validated.
	SpecValid bool `json:"specValid,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// FailureMessage indicates that there is a fatal problem reconciling the
	// state, and will be set to a descriptive error message.
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VSphereDatacenterConfig is the Schema for the VSphereDatacenterConfigs API.
type VSphereDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereDatacenterConfigSpec   `json:"spec,omitempty"`
	Status VSphereDatacenterConfigStatus `json:"status,omitempty"`
}

func (v *VSphereDatacenterConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *VSphereDatacenterConfig) ExpectedKind() string {
	return VSphereDatacenterKind
}

func (v *VSphereDatacenterConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *VSphereDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *VSphereDatacenterConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

func (v *VSphereDatacenterConfig) SetDefaults() {
	v.Spec.Network = generateFullVCenterPath(networkFolderType, v.Spec.Network, v.Spec.Datacenter)

	if v.Spec.Insecure {
		logger.Info("Warning: VSphereDatacenterConfig configured in insecure mode")
		v.Spec.Thumbprint = ""
	}
}

func (v *VSphereDatacenterConfig) Validate() error {
	if len(v.Spec.Server) <= 0 {
		return errors.New("VSphereDatacenterConfig server is not set or is empty")
	}

	if len(v.Spec.Datacenter) <= 0 {
		return errors.New("VSphereDatacenterConfig datacenter is not set or is empty")
	}

	if len(v.Spec.Network) <= 0 {
		return errors.New("VSphereDatacenterConfig VM network is not set or is empty")
	}

	if err := validatePath(networkFolderType, v.Spec.Network, v.Spec.Datacenter); err != nil {
		return err
	}

	return nil
}

func (v *VSphereDatacenterConfig) ConvertConfigToConfigGenerateStruct() *VSphereDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if v.Namespace != "" {
		namespace = v.Namespace
	}
	config := &VSphereDatacenterConfigGenerate{
		TypeMeta: v.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        v.Name,
			Annotations: v.Annotations,
			Namespace:   namespace,
		},
		Spec: v.Spec,
	}

	return config
}

func (v *VSphereDatacenterConfig) Marshallable() Marshallable {
	return v.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as VSphereDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig.
type VSphereDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec VSphereDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereDatacenterConfigList contains a list of VSphereDatacenterConfig.
type VSphereDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereDatacenterConfig{}, &VSphereDatacenterConfigList{})
}
