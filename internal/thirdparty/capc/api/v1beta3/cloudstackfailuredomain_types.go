/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta3

import (
	"crypto/md5" // #nosec G501 -- weak cryptographic primitive doesn't matter here. Not security related.
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FailureDomainHashedMetaName returns an MD5 name generated from the FailureDomain and Cluster name.
// FailureDomains must have a unique name even when potentially sharing a namespace with other clusters.
// In the future we may remove the ability to run multiple clusters in a single namespace, but today
// this is a consequence of being upstream of EKS-A which does run multiple clusters in a single namepace.
func FailureDomainHashedMetaName(fdName, clusterName string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fdName+clusterName))) // #nosec G401 -- weak cryptographic primitive doesn't matter here. Not security related.
}

const (
	FailureDomainFinalizer = "cloudstackfailuredomain.infrastructure.cluster.x-k8s.io"
	FailureDomainLabelName = "cloudstackfailuredomain.infrastructure.cluster.x-k8s.io/name"
)

const (
	NetworkTypeIsolated = "Isolated"
	NetworkTypeShared   = "Shared"
)

type Network struct {
	// Cloudstack Network ID the cluster is built in.
	// +optional
	ID string `json:"id,omitempty"`

	// Cloudstack Network Type the cluster is built in.
	// + optional
	Type string `json:"type,omitempty"`

	// Cloudstack Network Name the cluster is built in.
	Name string `json:"name"`
}

// CloudStackZoneSpec specifies a Zone's details.
type CloudStackZoneSpec struct {
	// Name.
	//+optional
	Name string `json:"name,omitempty"`

	// ID.
	//+optional
	ID string `json:"id,omitempty"`

	// The network within the Zone to use.
	Network Network `json:"network"`
}

// CloudStackFailureDomainSpec defines the desired state of CloudStackFailureDomain
type CloudStackFailureDomainSpec struct {
	// The failure domain unique name.
	Name string `json:"name"`

	// The ACS Zone for this failure domain.
	Zone CloudStackZoneSpec `json:"zone"`

	// CloudStack account.
	// +optional
	Account string `json:"account,omitempty"`

	// CloudStack domain.
	// +optional
	Domain string `json:"domain,omitempty"`

	// Apache CloudStack Endpoint secret reference.
	ACSEndpoint corev1.SecretReference `json:"acsEndpoint"`
}

// CloudStackFailureDomainStatus defines the observed state of CloudStackFailureDomain
type CloudStackFailureDomainStatus struct {
	// Reflects the readiness of the CloudStack Failure Domain.
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CloudStackFailureDomain is the Schema for the cloudstackfailuredomains API
type CloudStackFailureDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackFailureDomainSpec   `json:"spec"`
	Status CloudStackFailureDomainStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudStackFailureDomainList contains a list of CloudStackFailureDomain
type CloudStackFailureDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudStackFailureDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudStackFailureDomain{}, &CloudStackFailureDomainList{})
}
