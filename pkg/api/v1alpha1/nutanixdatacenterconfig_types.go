// Important: Run "make generate" to regenerate code after modifying this file
// json tags are required; new fields must have json tags for the fields to be serialized

package v1alpha1

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// NutanixDatacenterConfigSpec defines the desired state of NutanixDatacenterConfig.
type NutanixDatacenterConfigSpec struct {
	// Endpoint is the Endpoint of Nutanix Prism Central
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Port is the Port of Nutanix Prism Central
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Default=9440
	Port int `json:"port"`

	// AdditionalTrustBundle is the optional PEM-encoded certificate bundle for
	// users that configured their Prism Central with certificates from non-publicly
	// trusted CAs
	AdditionalTrustBundle string `json:"additionalTrustBundle,omitempty"`

	// Insecure is the optional flag to skip TLS verification. Nutanix Prism
	// Central installation by default ships with a self-signed certificate
	// that will fail TLS verification because the certificate is not issued by
	// a public CA and does not have the IP SANs with the Prism Central endpoint.
	// To accommodate the scenario where the user has not changed the default
	// Certificate that ships with Prism Central, we allow the user to skip TLS
	// verification. This is not recommended for production use.
	Insecure bool `json:"insecure,omitempty"`

	// CredentialRef is the reference to the secret name that contains the credentials
	// for the Nutanix Prism Central. The namespace for the secret is assumed to be a constant i.e. eksa-system.
	// +optional
	CredentialRef *Ref `json:"credentialRef,omitempty"`
}

// NutanixDatacenterConfigStatus defines the observed state of NutanixDatacenterConfig.
type NutanixDatacenterConfigStatus struct{}

// NutanixDatacenterConfig is the Schema for the NutanixDatacenterConfigs API
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type NutanixDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixDatacenterConfigSpec   `json:"spec,omitempty"`
	Status NutanixDatacenterConfigStatus `json:"status,omitempty"`
}

func (in *NutanixDatacenterConfig) Kind() string {
	return in.TypeMeta.Kind
}

func (in *NutanixDatacenterConfig) ExpectedKind() string {
	return NutanixDatacenterKind
}

func (in *NutanixDatacenterConfig) PauseReconcile() {
	if in.Annotations == nil {
		in.Annotations = map[string]string{}
	}
	in.Annotations[pausedAnnotation] = "true"
}

func (in *NutanixDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := in.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (in *NutanixDatacenterConfig) ClearPauseAnnotation() {
	if in.Annotations != nil {
		delete(in.Annotations, pausedAnnotation)
	}
}

func (in *NutanixDatacenterConfig) ConvertConfigToConfigGenerateStruct() *NutanixDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if in.Namespace != "" {
		namespace = in.Namespace
	}
	config := &NutanixDatacenterConfigGenerate{
		TypeMeta: in.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        in.Name,
			Annotations: in.Annotations,
			Namespace:   namespace,
		},
		Spec: in.Spec,
	}

	return config
}

func (in *NutanixDatacenterConfig) Marshallable() Marshallable {
	return in.ConvertConfigToConfigGenerateStruct()
}

func (in *NutanixDatacenterConfig) Validate() error {
	if len(in.Spec.Endpoint) <= 0 {
		return errors.New("NutanixDatacenterConfig endpoint is not set or is empty")
	}

	if in.Spec.Port == 0 {
		return errors.New("NutanixDatacenterConfig port is not set or is empty")
	}

	if len(in.Spec.AdditionalTrustBundle) > 0 {
		certPem := []byte(in.Spec.AdditionalTrustBundle)
		block, _ := pem.Decode(certPem)
		if block == nil {
			return errors.New("NutanixDatacenterConfig additionalTrustBundle is not valid: could not find a PEM block in the certificate")
		}
		if _, err := x509.ParseCertificates(block.Bytes); err != nil {
			return fmt.Errorf("NutanixDatacenterConfig additionalTrustBundle is not valid: %s", err)
		}
	}

	if in.Spec.CredentialRef != nil {
		if in.Spec.CredentialRef.Kind != constants.SecretKind {
			return fmt.Errorf("NutanixDatacenterConfig credentialRef Kind (%s) is not a secret", in.Spec.CredentialRef.Kind)
		}

		if len(in.Spec.CredentialRef.Name) <= 0 {
			return errors.New("NutanixDatacenterConfig credentialRef name is not set or is empty")
		}
	}

	return nil
}

// SetDefaults sets default values for the NutanixDatacenterConfig object.
func (in *NutanixDatacenterConfig) SetDefaults() {
	if in.Spec.CredentialRef == nil {
		in.Spec.CredentialRef = &Ref{
			Kind: constants.SecretKind,
			Name: constants.NutanixCredentialsName,
		}
	}
}

// NutanixDatacenterConfigGenerate is same as NutanixDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
//
// +kubebuilder:object:generate=false
type NutanixDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec NutanixDatacenterConfigSpec `json:"spec,omitempty"`
}

// NutanixDatacenterConfigList contains a list of NutanixDatacenterConfig
//
// +kubebuilder:object:root=true
type NutanixDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixDatacenterConfig{}, &NutanixDatacenterConfigList{})
}
