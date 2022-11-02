package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// OIDCConfig defines an OpenID Connect (OIDCConfigSpec) identity provider configuration

// OIDCConfigSpec defines the desired state of OIDCConfig.
type OIDCConfigSpec struct {
	// ClientId defines the client ID for the OpenID Connect client
	ClientId string `json:"clientId,omitempty"`
	// +kubebuilder:validation:Optional
	// GroupsClaim defines the name of a custom OpenID Connect claim for specifying user groups
	GroupsClaim string `json:"groupsClaim,omitempty"`
	// +kubebuilder:validation:Optional
	// GroupsPrefix defines a string to be prefixed to all groups to prevent conflicts with other authentication strategies
	GroupsPrefix string `json:"groupsPrefix,omitempty"`
	// IssuerUrl defines the URL of the OpenID issuer, only HTTPS scheme will be accepted
	IssuerUrl string `json:"issuerUrl,omitempty"`
	// +kubebuilder:validation:Optional
	// RequiredClaims defines a key=value pair that describes a required claim in the ID Token
	RequiredClaims []OIDCConfigRequiredClaim `json:"requiredClaims,omitempty"`
	// +kubebuilder:validation:Optional
	// UsernameClaim defines the OpenID claim to use as the user name. Note that claims other than the default ('sub') is not guaranteed to be unique and immutable
	UsernameClaim string `json:"usernameClaim,omitempty"`
	// +kubebuilder:validation:Optional
	// UsernamePrefix defines a string to prefixed to all usernames. If not provided, username claims other than 'email' are prefixed by the issuer URL to avoid clashes. To skip any prefixing, provide the value '-'.
	UsernamePrefix string `json:"usernamePrefix,omitempty"`
}

func (e *OIDCConfigSpec) Equal(n *OIDCConfigSpec) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	if e.ClientId != n.ClientId {
		return false
	}
	if e.GroupsClaim != n.GroupsClaim {
		return false
	}
	if e.GroupsPrefix != n.GroupsPrefix {
		return false
	}
	if e.IssuerUrl != n.IssuerUrl {
		return false
	}
	if e.UsernameClaim != n.UsernameClaim {
		return false
	}
	if e.UsernamePrefix != n.UsernamePrefix {
		return false
	}
	return RequiredClaimsSliceEqual(e.RequiredClaims, n.RequiredClaims)
}

func RequiredClaimsSliceEqual(a, b []OIDCConfigRequiredClaim) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v.Claim+v.Value]++
	}
	for _, v := range b {
		if _, ok := m[v.Claim+v.Value]; !ok {
			return false
		}
		m[v.Claim+v.Value] -= 1
		if m[v.Claim+v.Value] == 0 {
			delete(m, v.Claim+v.Value)
		}
	}
	return len(m) == 0
}

// IsManaged returns true if the oidcconfig is associated with a workload cluster.
func (c *OIDCConfig) IsManaged() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *OIDCConfig) SetManagedBy(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

type OIDCConfigRequiredClaim struct {
	Claim string `json:"claim,omitempty"`
	Value string `json:"value,omitempty"`
}

// OIDCConfigStatus defines the observed state of OIDCConfig.
type OIDCConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OIDCConfig is the Schema for the oidcconfigs API.
type OIDCConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OIDCConfigSpec   `json:"spec,omitempty"`
	Status OIDCConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as OIDCConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled.
type OIDCConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec OIDCConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// OIDCConfigList contains a list of OIDCConfig.
type OIDCConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OIDCConfig `json:"items"`
}

func (c *OIDCConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *OIDCConfig) ExpectedKind() string {
	return OIDCConfigKind
}

func (c *OIDCConfig) Validate() field.ErrorList {
	return validateOIDCConfig(c)
}

func (c *OIDCConfig) ConvertConfigToConfigGenerateStruct() *OIDCConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &OIDCConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}
	return config
}

func init() {
	SchemeBuilder.Register(&OIDCConfig{}, &OIDCConfigList{})
}
