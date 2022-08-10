// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const DefaultCloudStackAZPrefix = "default-az"

// CloudStackDatacenterConfigSpec defines the desired state of CloudStackDatacenterConfig
type CloudStackDatacenterConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Domain contains a grouping of accounts. Domains usually contain multiple accounts that have some logical relationship to each other and a set of delegated administrators with some authority over the domain and its subdomains
	// This field is considered as a fully qualified domain name which is the same as the domain path without "ROOT/" prefix. For example, if "foo" is specified then a domain with "ROOT/foo" domain path is picked.
	// The value "ROOT" is a special case that points to "the" ROOT domain of the CloudStack. That is, a domain with a path "ROOT/ROOT" is not allowed.
	// +optional
	Domain string `json:"domain,omitempty"`
	// Zones is a list of one or more zones that are managed by a single CloudStack management endpoint.
	// +optional
	Zones []CloudStackZone `json:"zones,omitempty"`
	// Account typically represents a customer of the service provider or a department in a large organization. Multiple users can exist in an account, and all CloudStack resources belong to an account. Accounts have users and users have credentials to operate on resources within that account. If an account name is provided, a domain must also be provided.
	// +optional
	Account string `json:"account,omitempty"`
	// CloudStack Management API endpoint's IP. It is added to VM's noproxy list
	// +optional
	ManagementApiEndpoint string `json:"managementApiEndpoint,omitempty"`
	// AvailabilityZones list of different partitions to distribute VMs across - corresponds to a list of CAPI failure domains
	// +optional
	AvailabilityZones []CloudStackAvailabilityZone `json:"availabilityZones,omitempty"`
}

type CloudStackResourceIdentifier struct {
	// Id of a resource in the CloudStack environment. Mutually exclusive with Name
	// +optional
	Id string `json:"id,omitempty"`
	// Name of a resource in the CloudStack environment. Mutually exclusive with Id
	// +optional
	Name string `json:"name,omitempty"`
}

func (r *CloudStackResourceIdentifier) Equal(o *CloudStackResourceIdentifier) bool {
	if r == o {
		return true
	}
	if r == nil || o == nil {
		return false
	}
	if r.Id != o.Id {
		return false
	}
	return r.Id == "" && o.Id == "" && r.Name == o.Name
}

// CloudStackZone is an organizational construct typically used to represent a single datacenter, and all its physical and virtual resources exist inside that zone. It can either be specified as a UUID or name
type CloudStackZone struct {
	// Zone is the name or UUID of the CloudStack zone in which clusters should be created. Zones should be managed by a single CloudStack Management endpoint.
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Network is the name or UUID of the CloudStack network in which clusters should be created. It can either be an isolated or shared network. If it doesn’t already exist in CloudStack, it’ll automatically be created by CAPC as an isolated network. It can either be specified as a UUID or name
	// In multiple-zones situation, only 'Shared' network is supported.
	Network CloudStackResourceIdentifier `json:"network"`
}

// CloudStackAvailabilityZone maps to a CAPI failure domain to distribute machines across Cloudstack infrastructure
type CloudStackAvailabilityZone struct {
	// Name is used as a unique identifier for each availability zone
	Name string `json:"name"`
	// CredentialRef is used to reference a secret in the eksa-system namespace
	CredentialsRef string `json:"credentialsRef"`
	// Zone represents the properties of the CloudStack zone in which clusters should be created, like the network.
	Zone CloudStackZone `json:"zone"`
	// Domain contains a grouping of accounts. Domains usually contain multiple accounts that have some logical relationship to each other and a set of delegated administrators with some authority over the domain and its subdomains
	// This field is considered as a fully qualified domain name which is the same as the domain path without "ROOT/" prefix. For example, if "foo" is specified then a domain with "ROOT/foo" domain path is picked.
	// The value "ROOT" is a special case that points to "the" ROOT domain of the CloudStack. That is, a domain with a path "ROOT/ROOT" is not allowed.
	Domain string `json:"domain"`
	// Account typically represents a customer of the service provider or a department in a large organization. Multiple users can exist in an account, and all CloudStack resources belong to an account. Accounts have users and users have credentials to operate on resources within that account. If an account name is provided, a domain must also be provided.
	Account string `json:"account,omitempty"`
	// CloudStack Management API endpoint's IP. It is added to VM's noproxy list
	ManagementApiEndpoint string `json:"managementApiEndpoint"`
}

// CloudStackDatacenterConfigStatus defines the observed state of CloudStackDatacenterConfig
type CloudStackDatacenterConfigStatus struct { // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudStackDatacenterConfig is the Schema for the cloudstackdatacenterconfigs API
type CloudStackDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackDatacenterConfigSpec   `json:"spec,omitempty"`
	Status CloudStackDatacenterConfigStatus `json:"status,omitempty"`
}

func (v *CloudStackDatacenterConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *CloudStackDatacenterConfig) ExpectedKind() string {
	return CloudStackDatacenterKind
}

func (v *CloudStackDatacenterConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *CloudStackDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *CloudStackDatacenterConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

func (v *CloudStackDatacenterConfig) ConvertConfigToConfigGenerateStruct() *CloudStackDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if v.Namespace != "" {
		namespace = v.Namespace
	}
	config := &CloudStackDatacenterConfigGenerate{
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

func (v *CloudStackDatacenterConfig) Marshallable() Marshallable {
	return v.ConvertConfigToConfigGenerateStruct()
}

func (v *CloudStackDatacenterConfig) Validate() error {
	if v.Spec.Account != "" {
		return errors.New("account must be empty")
	}
	if v.Spec.Domain != "" {
		return errors.New("domain must be empty")
	}
	if v.Spec.ManagementApiEndpoint != "" {
		return errors.New("managementApiEndpoint must be empty")
	}
	if len(v.Spec.Zones) > 0 {
		return errors.New("zones must be empty")
	}
	if len(v.Spec.AvailabilityZones) == 0 {
		return errors.New("availabilityZones must not be empty")
	}
	azSet := make(map[string]bool)
	for _, az := range v.Spec.AvailabilityZones {
		if exists := azSet[az.Name]; exists {
			return fmt.Errorf("availabilityZone names must be unique. Duplicate name: %s", az.Name)
		}
		azSet[az.Name] = true
	}

	return nil
}

func (v *CloudStackDatacenterConfig) SetDefaults() {
	if v.Spec.AvailabilityZones == nil || len(v.Spec.AvailabilityZones) == 0 {
		v.Spec.AvailabilityZones = make([]CloudStackAvailabilityZone, 0, len(v.Spec.Zones))
		for index, csZone := range v.Spec.Zones {
			az := CloudStackAvailabilityZone{
				Name:                  fmt.Sprintf("%s-%d", DefaultCloudStackAZPrefix, index),
				Zone:                  csZone,
				Account:               v.Spec.Account,
				Domain:                v.Spec.Domain,
				ManagementApiEndpoint: v.Spec.ManagementApiEndpoint,
				CredentialsRef:        "global",
			}
			v.Spec.AvailabilityZones = append(v.Spec.AvailabilityZones, az)
		}
	}
	v.Spec.Zones = nil
	v.Spec.Domain = ""
	v.Spec.Account = ""
	v.Spec.ManagementApiEndpoint = ""
}

func (s *CloudStackDatacenterConfigSpec) Equal(o *CloudStackDatacenterConfigSpec) bool {
	if s == o {
		return true
	}
	if s == nil || o == nil {
		return false
	}
	if len(s.Zones) != len(o.Zones) {
		return false
	}
	for i, z := range s.Zones {
		if !z.Equal(&o.Zones[i]) {
			return false
		}
	}
	if len(s.AvailabilityZones) != len(o.AvailabilityZones) {
		return false
	}
	oAzsMap := map[string]CloudStackAvailabilityZone{}
	for _, oAz := range o.AvailabilityZones {
		oAzsMap[oAz.Name] = oAz
	}
	for _, sAz := range s.AvailabilityZones {
		oAz, found := oAzsMap[sAz.Name]
		if !found || !sAz.Equal(&oAz) {
			return false
		}
	}
	return s.ManagementApiEndpoint == o.ManagementApiEndpoint &&
		s.Domain == o.Domain &&
		s.Account == o.Account
}

func (z *CloudStackZone) Equal(o *CloudStackZone) bool {
	if z == o {
		return true
	}
	if z == nil || o == nil {
		return false
	}
	if z.Id == o.Id &&
		z.Name == o.Name &&
		z.Network.Id == o.Network.Id &&
		z.Network.Name == o.Network.Name {
		return true
	}
	return false
}

func (az *CloudStackAvailabilityZone) Equal(o *CloudStackAvailabilityZone) bool {
	if az == o {
		return true
	}
	if az == nil || o == nil {
		return false
	}
	return az.Zone.Equal(&o.Zone) &&
		az.Name == o.Name &&
		az.CredentialsRef == o.CredentialsRef &&
		az.Account == o.Account &&
		az.Domain == o.Domain &&
		az.ManagementApiEndpoint == o.ManagementApiEndpoint
}

// +kubebuilder:object:generate=false

// Same as CloudStackDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type CloudStackDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec CloudStackDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackDatacenterConfigList contains a list of CloudStackDatacenterConfig
type CloudStackDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackDatacenterConfig{}, &CloudStackDatacenterConfigList{})
}
