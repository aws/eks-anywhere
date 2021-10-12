// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDatacenterConfig) DeepCopyInto(out *AWSDatacenterConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDatacenterConfig.
func (in *AWSDatacenterConfig) DeepCopy() *AWSDatacenterConfig {
	if in == nil {
		return nil
	}
	out := new(AWSDatacenterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSDatacenterConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDatacenterConfigList) DeepCopyInto(out *AWSDatacenterConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSDatacenterConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDatacenterConfigList.
func (in *AWSDatacenterConfigList) DeepCopy() *AWSDatacenterConfigList {
	if in == nil {
		return nil
	}
	out := new(AWSDatacenterConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSDatacenterConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDatacenterConfigSpec) DeepCopyInto(out *AWSDatacenterConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDatacenterConfigSpec.
func (in *AWSDatacenterConfigSpec) DeepCopy() *AWSDatacenterConfigSpec {
	if in == nil {
		return nil
	}
	out := new(AWSDatacenterConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDatacenterConfigStatus) DeepCopyInto(out *AWSDatacenterConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDatacenterConfigStatus.
func (in *AWSDatacenterConfigStatus) DeepCopy() *AWSDatacenterConfigStatus {
	if in == nil {
		return nil
	}
	out := new(AWSDatacenterConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterList.
func (in *ClusterList) DeepCopy() *ClusterList {
	if in == nil {
		return nil
	}
	out := new(ClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetwork) DeepCopyInto(out *ClusterNetwork) {
	*out = *in
	in.Pods.DeepCopyInto(&out.Pods)
	in.Services.DeepCopyInto(&out.Services)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetwork.
func (in *ClusterNetwork) DeepCopy() *ClusterNetwork {
	if in == nil {
		return nil
	}
	out := new(ClusterNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	in.ControlPlaneConfiguration.DeepCopyInto(&out.ControlPlaneConfiguration)
	if in.WorkerNodeGroupConfigurations != nil {
		in, out := &in.WorkerNodeGroupConfigurations, &out.WorkerNodeGroupConfigurations
		*out = make([]WorkerNodeGroupConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.DatacenterRef = in.DatacenterRef
	if in.IdentityProviderRefs != nil {
		in, out := &in.IdentityProviderRefs, &out.IdentityProviderRefs
		*out = make([]Ref, len(*in))
		copy(*out, *in)
	}
	if in.GitOpsRef != nil {
		in, out := &in.GitOpsRef, &out.GitOpsRef
		*out = new(Ref)
		**out = **in
	}
	in.ClusterNetwork.DeepCopyInto(&out.ClusterNetwork)
	if in.ExternalEtcdConfiguration != nil {
		in, out := &in.ExternalEtcdConfiguration, &out.ExternalEtcdConfiguration
		*out = new(ExternalEtcdConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.ProxyConfiguration != nil {
		in, out := &in.ProxyConfiguration, &out.ProxyConfiguration
		*out = new(ProxyConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.RegistryMirrorConfiguration != nil {
		in, out := &in.RegistryMirrorConfiguration, &out.RegistryMirrorConfiguration
		*out = new(RegistryMirrorConfiguration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneConfiguration) DeepCopyInto(out *ControlPlaneConfiguration) {
	*out = *in
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(Endpoint)
		**out = **in
	}
	if in.MachineGroupRef != nil {
		in, out := &in.MachineGroupRef, &out.MachineGroupRef
		*out = new(Ref)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneConfiguration.
func (in *ControlPlaneConfiguration) DeepCopy() *ControlPlaneConfiguration {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DockerDatacenterConfig) DeepCopyInto(out *DockerDatacenterConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DockerDatacenterConfig.
func (in *DockerDatacenterConfig) DeepCopy() *DockerDatacenterConfig {
	if in == nil {
		return nil
	}
	out := new(DockerDatacenterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DockerDatacenterConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DockerDatacenterConfigList) DeepCopyInto(out *DockerDatacenterConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DockerDatacenterConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DockerDatacenterConfigList.
func (in *DockerDatacenterConfigList) DeepCopy() *DockerDatacenterConfigList {
	if in == nil {
		return nil
	}
	out := new(DockerDatacenterConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DockerDatacenterConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DockerDatacenterConfigSpec) DeepCopyInto(out *DockerDatacenterConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DockerDatacenterConfigSpec.
func (in *DockerDatacenterConfigSpec) DeepCopy() *DockerDatacenterConfigSpec {
	if in == nil {
		return nil
	}
	out := new(DockerDatacenterConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DockerDatacenterConfigStatus) DeepCopyInto(out *DockerDatacenterConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DockerDatacenterConfigStatus.
func (in *DockerDatacenterConfigStatus) DeepCopy() *DockerDatacenterConfigStatus {
	if in == nil {
		return nil
	}
	out := new(DockerDatacenterConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalEtcdConfiguration) DeepCopyInto(out *ExternalEtcdConfiguration) {
	*out = *in
	if in.MachineGroupRef != nil {
		in, out := &in.MachineGroupRef, &out.MachineGroupRef
		*out = new(Ref)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalEtcdConfiguration.
func (in *ExternalEtcdConfiguration) DeepCopy() *ExternalEtcdConfiguration {
	if in == nil {
		return nil
	}
	out := new(ExternalEtcdConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Flux) DeepCopyInto(out *Flux) {
	*out = *in
	out.Github = in.Github
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Flux.
func (in *Flux) DeepCopy() *Flux {
	if in == nil {
		return nil
	}
	out := new(Flux)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitOpsConfig) DeepCopyInto(out *GitOpsConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitOpsConfig.
func (in *GitOpsConfig) DeepCopy() *GitOpsConfig {
	if in == nil {
		return nil
	}
	out := new(GitOpsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitOpsConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitOpsConfigList) DeepCopyInto(out *GitOpsConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitOpsConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitOpsConfigList.
func (in *GitOpsConfigList) DeepCopy() *GitOpsConfigList {
	if in == nil {
		return nil
	}
	out := new(GitOpsConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitOpsConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitOpsConfigSpec) DeepCopyInto(out *GitOpsConfigSpec) {
	*out = *in
	out.Flux = in.Flux
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitOpsConfigSpec.
func (in *GitOpsConfigSpec) DeepCopy() *GitOpsConfigSpec {
	if in == nil {
		return nil
	}
	out := new(GitOpsConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitOpsConfigStatus) DeepCopyInto(out *GitOpsConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitOpsConfigStatus.
func (in *GitOpsConfigStatus) DeepCopy() *GitOpsConfigStatus {
	if in == nil {
		return nil
	}
	out := new(GitOpsConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Github) DeepCopyInto(out *Github) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Github.
func (in *Github) DeepCopy() *Github {
	if in == nil {
		return nil
	}
	out := new(Github)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfig) DeepCopyInto(out *OIDCConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfig.
func (in *OIDCConfig) DeepCopy() *OIDCConfig {
	if in == nil {
		return nil
	}
	out := new(OIDCConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigList) DeepCopyInto(out *OIDCConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OIDCConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigList.
func (in *OIDCConfigList) DeepCopy() *OIDCConfigList {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigRequiredClaim) DeepCopyInto(out *OIDCConfigRequiredClaim) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigRequiredClaim.
func (in *OIDCConfigRequiredClaim) DeepCopy() *OIDCConfigRequiredClaim {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigRequiredClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigSpec) DeepCopyInto(out *OIDCConfigSpec) {
	*out = *in
	if in.RequiredClaims != nil {
		in, out := &in.RequiredClaims, &out.RequiredClaims
		*out = make([]OIDCConfigRequiredClaim, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigSpec.
func (in *OIDCConfigSpec) DeepCopy() *OIDCConfigSpec {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigStatus) DeepCopyInto(out *OIDCConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigStatus.
func (in *OIDCConfigStatus) DeepCopy() *OIDCConfigStatus {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectMeta) DeepCopyInto(out *ObjectMeta) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectMeta.
func (in *ObjectMeta) DeepCopy() *ObjectMeta {
	if in == nil {
		return nil
	}
	out := new(ObjectMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pods) DeepCopyInto(out *Pods) {
	*out = *in
	if in.CidrBlocks != nil {
		in, out := &in.CidrBlocks, &out.CidrBlocks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pods.
func (in *Pods) DeepCopy() *Pods {
	if in == nil {
		return nil
	}
	out := new(Pods)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyConfiguration) DeepCopyInto(out *ProxyConfiguration) {
	*out = *in
	if in.NoProxy != nil {
		in, out := &in.NoProxy, &out.NoProxy
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyConfiguration.
func (in *ProxyConfiguration) DeepCopy() *ProxyConfiguration {
	if in == nil {
		return nil
	}
	out := new(ProxyConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ref) DeepCopyInto(out *Ref) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ref.
func (in *Ref) DeepCopy() *Ref {
	if in == nil {
		return nil
	}
	out := new(Ref)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryMirrorConfiguration) DeepCopyInto(out *RegistryMirrorConfiguration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryMirrorConfiguration.
func (in *RegistryMirrorConfiguration) DeepCopy() *RegistryMirrorConfiguration {
	if in == nil {
		return nil
	}
	out := new(RegistryMirrorConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Services) DeepCopyInto(out *Services) {
	*out = *in
	if in.CidrBlocks != nil {
		in, out := &in.CidrBlocks, &out.CidrBlocks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Services.
func (in *Services) DeepCopy() *Services {
	if in == nil {
		return nil
	}
	out := new(Services)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UserConfiguration) DeepCopyInto(out *UserConfiguration) {
	*out = *in
	if in.SshAuthorizedKeys != nil {
		in, out := &in.SshAuthorizedKeys, &out.SshAuthorizedKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UserConfiguration.
func (in *UserConfiguration) DeepCopy() *UserConfiguration {
	if in == nil {
		return nil
	}
	out := new(UserConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereDatacenterConfig) DeepCopyInto(out *VSphereDatacenterConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereDatacenterConfig.
func (in *VSphereDatacenterConfig) DeepCopy() *VSphereDatacenterConfig {
	if in == nil {
		return nil
	}
	out := new(VSphereDatacenterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereDatacenterConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereDatacenterConfigList) DeepCopyInto(out *VSphereDatacenterConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VSphereDatacenterConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereDatacenterConfigList.
func (in *VSphereDatacenterConfigList) DeepCopy() *VSphereDatacenterConfigList {
	if in == nil {
		return nil
	}
	out := new(VSphereDatacenterConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereDatacenterConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereDatacenterConfigSpec) DeepCopyInto(out *VSphereDatacenterConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereDatacenterConfigSpec.
func (in *VSphereDatacenterConfigSpec) DeepCopy() *VSphereDatacenterConfigSpec {
	if in == nil {
		return nil
	}
	out := new(VSphereDatacenterConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereDatacenterConfigStatus) DeepCopyInto(out *VSphereDatacenterConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereDatacenterConfigStatus.
func (in *VSphereDatacenterConfigStatus) DeepCopy() *VSphereDatacenterConfigStatus {
	if in == nil {
		return nil
	}
	out := new(VSphereDatacenterConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereMachineConfig) DeepCopyInto(out *VSphereMachineConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereMachineConfig.
func (in *VSphereMachineConfig) DeepCopy() *VSphereMachineConfig {
	if in == nil {
		return nil
	}
	out := new(VSphereMachineConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereMachineConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereMachineConfigList) DeepCopyInto(out *VSphereMachineConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VSphereMachineConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereMachineConfigList.
func (in *VSphereMachineConfigList) DeepCopy() *VSphereMachineConfigList {
	if in == nil {
		return nil
	}
	out := new(VSphereMachineConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereMachineConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereMachineConfigSpec) DeepCopyInto(out *VSphereMachineConfigSpec) {
	*out = *in
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make([]UserConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereMachineConfigSpec.
func (in *VSphereMachineConfigSpec) DeepCopy() *VSphereMachineConfigSpec {
	if in == nil {
		return nil
	}
	out := new(VSphereMachineConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereMachineConfigStatus) DeepCopyInto(out *VSphereMachineConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereMachineConfigStatus.
func (in *VSphereMachineConfigStatus) DeepCopy() *VSphereMachineConfigStatus {
	if in == nil {
		return nil
	}
	out := new(VSphereMachineConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkerNodeGroupConfiguration) DeepCopyInto(out *WorkerNodeGroupConfiguration) {
	*out = *in
	if in.MachineGroupRef != nil {
		in, out := &in.MachineGroupRef, &out.MachineGroupRef
		*out = new(Ref)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkerNodeGroupConfiguration.
func (in *WorkerNodeGroupConfiguration) DeepCopy() *WorkerNodeGroupConfiguration {
	if in == nil {
		return nil
	}
	out := new(WorkerNodeGroupConfiguration)
	in.DeepCopyInto(out)
	return out
}
