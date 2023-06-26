//go:build !ignore_autogenerated
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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActionsBundle) DeepCopyInto(out *ActionsBundle) {
	*out = *in
	in.Cexec.DeepCopyInto(&out.Cexec)
	in.Kexec.DeepCopyInto(&out.Kexec)
	in.ImageToDisk.DeepCopyInto(&out.ImageToDisk)
	in.OciToDisk.DeepCopyInto(&out.OciToDisk)
	in.WriteFile.DeepCopyInto(&out.WriteFile)
	in.Reboot.DeepCopyInto(&out.Reboot)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActionsBundle.
func (in *ActionsBundle) DeepCopy() *ActionsBundle {
	if in == nil {
		return nil
	}
	out := new(ActionsBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArchitectureBundle) DeepCopyInto(out *ArchitectureBundle) {
	*out = *in
	in.Amd64.DeepCopyInto(&out.Amd64)
	in.Arm64.DeepCopyInto(&out.Arm64)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArchitectureBundle.
func (in *ArchitectureBundle) DeepCopy() *ArchitectureBundle {
	if in == nil {
		return nil
	}
	out := new(ArchitectureBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Archive) DeepCopyInto(out *Archive) {
	*out = *in
	if in.Arch != nil {
		in, out := &in.Arch, &out.Arch
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Archive.
func (in *Archive) DeepCopy() *Archive {
	if in == nil {
		return nil
	}
	out := new(Archive)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsBundle) DeepCopyInto(out *AwsBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.ClusterTemplate = in.ClusterTemplate
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsBundle.
func (in *AwsBundle) DeepCopy() *AwsBundle {
	if in == nil {
		return nil
	}
	out := new(AwsBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BinaryBundle) DeepCopyInto(out *BinaryBundle) {
	*out = *in
	in.LinuxBinary.DeepCopyInto(&out.LinuxBinary)
	in.DarwinBinary.DeepCopyInto(&out.DarwinBinary)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BinaryBundle.
func (in *BinaryBundle) DeepCopy() *BinaryBundle {
	if in == nil {
		return nil
	}
	out := new(BinaryBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BottlerocketHostContainersBundle) DeepCopyInto(out *BottlerocketHostContainersBundle) {
	*out = *in
	in.Admin.DeepCopyInto(&out.Admin)
	in.Control.DeepCopyInto(&out.Control)
	in.KubeadmBootstrap.DeepCopyInto(&out.KubeadmBootstrap)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BottlerocketHostContainersBundle.
func (in *BottlerocketHostContainersBundle) DeepCopy() *BottlerocketHostContainersBundle {
	if in == nil {
		return nil
	}
	out := new(BottlerocketHostContainersBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Bundles) DeepCopyInto(out *Bundles) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Bundles.
func (in *Bundles) DeepCopy() *Bundles {
	if in == nil {
		return nil
	}
	out := new(Bundles)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Bundles) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BundlesList) DeepCopyInto(out *BundlesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Bundles, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BundlesList.
func (in *BundlesList) DeepCopy() *BundlesList {
	if in == nil {
		return nil
	}
	out := new(BundlesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BundlesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BundlesSpec) DeepCopyInto(out *BundlesSpec) {
	*out = *in
	if in.VersionsBundles != nil {
		in, out := &in.VersionsBundles, &out.VersionsBundles
		*out = make([]VersionsBundle, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BundlesSpec.
func (in *BundlesSpec) DeepCopy() *BundlesSpec {
	if in == nil {
		return nil
	}
	out := new(BundlesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BundlesStatus) DeepCopyInto(out *BundlesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BundlesStatus.
func (in *BundlesStatus) DeepCopy() *BundlesStatus {
	if in == nil {
		return nil
	}
	out := new(BundlesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertManagerBundle) DeepCopyInto(out *CertManagerBundle) {
	*out = *in
	in.Acmesolver.DeepCopyInto(&out.Acmesolver)
	in.Cainjector.DeepCopyInto(&out.Cainjector)
	in.Controller.DeepCopyInto(&out.Controller)
	in.Ctl.DeepCopyInto(&out.Ctl)
	in.Webhook.DeepCopyInto(&out.Webhook)
	out.Manifest = in.Manifest
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertManagerBundle.
func (in *CertManagerBundle) DeepCopy() *CertManagerBundle {
	if in == nil {
		return nil
	}
	out := new(CertManagerBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CiliumBundle) DeepCopyInto(out *CiliumBundle) {
	*out = *in
	in.Cilium.DeepCopyInto(&out.Cilium)
	in.Operator.DeepCopyInto(&out.Operator)
	out.Manifest = in.Manifest
	in.HelmChart.DeepCopyInto(&out.HelmChart)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CiliumBundle.
func (in *CiliumBundle) DeepCopy() *CiliumBundle {
	if in == nil {
		return nil
	}
	out := new(CiliumBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudStackBundle) DeepCopyInto(out *CloudStackBundle) {
	*out = *in
	in.ClusterAPIController.DeepCopyInto(&out.ClusterAPIController)
	in.KubeRbacProxy.DeepCopyInto(&out.KubeRbacProxy)
	in.KubeVip.DeepCopyInto(&out.KubeVip)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudStackBundle.
func (in *CloudStackBundle) DeepCopy() *CloudStackBundle {
	if in == nil {
		return nil
	}
	out := new(CloudStackBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CoreClusterAPI) DeepCopyInto(out *CoreClusterAPI) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CoreClusterAPI.
func (in *CoreClusterAPI) DeepCopy() *CoreClusterAPI {
	if in == nil {
		return nil
	}
	out := new(CoreClusterAPI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DockerBundle) DeepCopyInto(out *DockerBundle) {
	*out = *in
	in.Manager.DeepCopyInto(&out.Manager)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.ClusterTemplate = in.ClusterTemplate
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DockerBundle.
func (in *DockerBundle) DeepCopy() *DockerBundle {
	if in == nil {
		return nil
	}
	out := new(DockerBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EksARelease) DeepCopyInto(out *EksARelease) {
	*out = *in
	in.EksABinary.DeepCopyInto(&out.EksABinary)
	in.EksACLI.DeepCopyInto(&out.EksACLI)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EksARelease.
func (in *EksARelease) DeepCopy() *EksARelease {
	if in == nil {
		return nil
	}
	out := new(EksARelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EksDRelease) DeepCopyInto(out *EksDRelease) {
	*out = *in
	in.KindNode.DeepCopyInto(&out.KindNode)
	in.Ami.DeepCopyInto(&out.Ami)
	in.Ova.DeepCopyInto(&out.Ova)
	in.Raw.DeepCopyInto(&out.Raw)
	in.Etcdadm.DeepCopyInto(&out.Etcdadm)
	in.Crictl.DeepCopyInto(&out.Crictl)
	in.ImageBuilder.DeepCopyInto(&out.ImageBuilder)
	in.Containerd.DeepCopyInto(&out.Containerd)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EksDRelease.
func (in *EksDRelease) DeepCopy() *EksDRelease {
	if in == nil {
		return nil
	}
	out := new(EksDRelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EksaBundle) DeepCopyInto(out *EksaBundle) {
	*out = *in
	in.CliTools.DeepCopyInto(&out.CliTools)
	in.ClusterController.DeepCopyInto(&out.ClusterController)
	in.DiagnosticCollector.DeepCopyInto(&out.DiagnosticCollector)
	out.Components = in.Components
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EksaBundle.
func (in *EksaBundle) DeepCopy() *EksaBundle {
	if in == nil {
		return nil
	}
	out := new(EksaBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EtcdadmBootstrapBundle) DeepCopyInto(out *EtcdadmBootstrapBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EtcdadmBootstrapBundle.
func (in *EtcdadmBootstrapBundle) DeepCopy() *EtcdadmBootstrapBundle {
	if in == nil {
		return nil
	}
	out := new(EtcdadmBootstrapBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EtcdadmControllerBundle) DeepCopyInto(out *EtcdadmControllerBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EtcdadmControllerBundle.
func (in *EtcdadmControllerBundle) DeepCopy() *EtcdadmControllerBundle {
	if in == nil {
		return nil
	}
	out := new(EtcdadmControllerBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FluxBundle) DeepCopyInto(out *FluxBundle) {
	*out = *in
	in.SourceController.DeepCopyInto(&out.SourceController)
	in.KustomizeController.DeepCopyInto(&out.KustomizeController)
	in.HelmController.DeepCopyInto(&out.HelmController)
	in.NotificationController.DeepCopyInto(&out.NotificationController)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FluxBundle.
func (in *FluxBundle) DeepCopy() *FluxBundle {
	if in == nil {
		return nil
	}
	out := new(FluxBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HaproxyBundle) DeepCopyInto(out *HaproxyBundle) {
	*out = *in
	in.Image.DeepCopyInto(&out.Image)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HaproxyBundle.
func (in *HaproxyBundle) DeepCopy() *HaproxyBundle {
	if in == nil {
		return nil
	}
	out := new(HaproxyBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HookArch) DeepCopyInto(out *HookArch) {
	*out = *in
	in.Arm.DeepCopyInto(&out.Arm)
	in.Amd.DeepCopyInto(&out.Amd)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookArch.
func (in *HookArch) DeepCopy() *HookArch {
	if in == nil {
		return nil
	}
	out := new(HookArch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HookBundle) DeepCopyInto(out *HookBundle) {
	*out = *in
	in.Bootkit.DeepCopyInto(&out.Bootkit)
	in.Docker.DeepCopyInto(&out.Docker)
	in.Kernel.DeepCopyInto(&out.Kernel)
	in.Initramfs.DeepCopyInto(&out.Initramfs)
	in.Vmlinuz.DeepCopyInto(&out.Vmlinuz)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HookBundle.
func (in *HookBundle) DeepCopy() *HookBundle {
	if in == nil {
		return nil
	}
	out := new(HookBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Image) DeepCopyInto(out *Image) {
	*out = *in
	if in.Arch != nil {
		in, out := &in.Arch, &out.Arch
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Image.
func (in *Image) DeepCopy() *Image {
	if in == nil {
		return nil
	}
	out := new(Image)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KindnetdBundle) DeepCopyInto(out *KindnetdBundle) {
	*out = *in
	out.Manifest = in.Manifest
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KindnetdBundle.
func (in *KindnetdBundle) DeepCopy() *KindnetdBundle {
	if in == nil {
		return nil
	}
	out := new(KindnetdBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeadmBootstrapBundle) DeepCopyInto(out *KubeadmBootstrapBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeadmBootstrapBundle.
func (in *KubeadmBootstrapBundle) DeepCopy() *KubeadmBootstrapBundle {
	if in == nil {
		return nil
	}
	out := new(KubeadmBootstrapBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeadmControlPlaneBundle) DeepCopyInto(out *KubeadmControlPlaneBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	out.Components = in.Components
	out.Metadata = in.Metadata
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeadmControlPlaneBundle.
func (in *KubeadmControlPlaneBundle) DeepCopy() *KubeadmControlPlaneBundle {
	if in == nil {
		return nil
	}
	out := new(KubeadmControlPlaneBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Manifest) DeepCopyInto(out *Manifest) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Manifest.
func (in *Manifest) DeepCopy() *Manifest {
	if in == nil {
		return nil
	}
	out := new(Manifest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NutanixBundle) DeepCopyInto(out *NutanixBundle) {
	*out = *in
	in.ClusterAPIController.DeepCopyInto(&out.ClusterAPIController)
	in.KubeVip.DeepCopyInto(&out.KubeVip)
	out.Components = in.Components
	out.Metadata = in.Metadata
	out.ClusterTemplate = in.ClusterTemplate
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NutanixBundle.
func (in *NutanixBundle) DeepCopy() *NutanixBundle {
	if in == nil {
		return nil
	}
	out := new(NutanixBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OSImageBundle) DeepCopyInto(out *OSImageBundle) {
	*out = *in
	in.Bottlerocket.DeepCopyInto(&out.Bottlerocket)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OSImageBundle.
func (in *OSImageBundle) DeepCopy() *OSImageBundle {
	if in == nil {
		return nil
	}
	out := new(OSImageBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PackageBundle) DeepCopyInto(out *PackageBundle) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.TokenRefresher.DeepCopyInto(&out.TokenRefresher)
	in.HelmChart.DeepCopyInto(&out.HelmChart)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PackageBundle.
func (in *PackageBundle) DeepCopy() *PackageBundle {
	if in == nil {
		return nil
	}
	out := new(PackageBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PlatformBundle) DeepCopyInto(out *PlatformBundle) {
	*out = *in
	in.LinuxBinary.DeepCopyInto(&out.LinuxBinary)
	in.DarwinBinary.DeepCopyInto(&out.DarwinBinary)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PlatformBundle.
func (in *PlatformBundle) DeepCopy() *PlatformBundle {
	if in == nil {
		return nil
	}
	out := new(PlatformBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Release) DeepCopyInto(out *Release) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Release.
func (in *Release) DeepCopy() *Release {
	if in == nil {
		return nil
	}
	out := new(Release)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Release) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseList) DeepCopyInto(out *ReleaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Release, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseList.
func (in *ReleaseList) DeepCopy() *ReleaseList {
	if in == nil {
		return nil
	}
	out := new(ReleaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReleaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseSpec) DeepCopyInto(out *ReleaseSpec) {
	*out = *in
	if in.Releases != nil {
		in, out := &in.Releases, &out.Releases
		*out = make([]EksARelease, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseSpec.
func (in *ReleaseSpec) DeepCopy() *ReleaseSpec {
	if in == nil {
		return nil
	}
	out := new(ReleaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseStatus) DeepCopyInto(out *ReleaseStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseStatus.
func (in *ReleaseStatus) DeepCopy() *ReleaseStatus {
	if in == nil {
		return nil
	}
	out := new(ReleaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SnowBundle) DeepCopyInto(out *SnowBundle) {
	*out = *in
	in.Manager.DeepCopyInto(&out.Manager)
	in.KubeVip.DeepCopyInto(&out.KubeVip)
	out.Components = in.Components
	out.Metadata = in.Metadata
	in.BottlerocketBootstrapSnow.DeepCopyInto(&out.BottlerocketBootstrapSnow)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SnowBundle.
func (in *SnowBundle) DeepCopy() *SnowBundle {
	if in == nil {
		return nil
	}
	out := new(SnowBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TinkBundle) DeepCopyInto(out *TinkBundle) {
	*out = *in
	in.TinkController.DeepCopyInto(&out.TinkController)
	in.TinkServer.DeepCopyInto(&out.TinkServer)
	in.TinkWorker.DeepCopyInto(&out.TinkWorker)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TinkBundle.
func (in *TinkBundle) DeepCopy() *TinkBundle {
	if in == nil {
		return nil
	}
	out := new(TinkBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TinkerbellBundle) DeepCopyInto(out *TinkerbellBundle) {
	*out = *in
	in.ClusterAPIController.DeepCopyInto(&out.ClusterAPIController)
	in.KubeVip.DeepCopyInto(&out.KubeVip)
	in.Envoy.DeepCopyInto(&out.Envoy)
	out.Components = in.Components
	out.Metadata = in.Metadata
	out.ClusterTemplate = in.ClusterTemplate
	in.TinkerbellStack.DeepCopyInto(&out.TinkerbellStack)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TinkerbellBundle.
func (in *TinkerbellBundle) DeepCopy() *TinkerbellBundle {
	if in == nil {
		return nil
	}
	out := new(TinkerbellBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TinkerbellStackBundle) DeepCopyInto(out *TinkerbellStackBundle) {
	*out = *in
	in.Actions.DeepCopyInto(&out.Actions)
	in.Boots.DeepCopyInto(&out.Boots)
	in.Hegel.DeepCopyInto(&out.Hegel)
	in.TinkebellChart.DeepCopyInto(&out.TinkebellChart)
	in.Hook.DeepCopyInto(&out.Hook)
	in.Rufio.DeepCopyInto(&out.Rufio)
	in.Tink.DeepCopyInto(&out.Tink)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TinkerbellStackBundle.
func (in *TinkerbellStackBundle) DeepCopy() *TinkerbellStackBundle {
	if in == nil {
		return nil
	}
	out := new(TinkerbellStackBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereBundle) DeepCopyInto(out *VSphereBundle) {
	*out = *in
	in.ClusterAPIController.DeepCopyInto(&out.ClusterAPIController)
	in.KubeProxy.DeepCopyInto(&out.KubeProxy)
	in.Manager.DeepCopyInto(&out.Manager)
	in.KubeVip.DeepCopyInto(&out.KubeVip)
	out.Components = in.Components
	out.Metadata = in.Metadata
	out.ClusterTemplate = in.ClusterTemplate
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereBundle.
func (in *VSphereBundle) DeepCopy() *VSphereBundle {
	if in == nil {
		return nil
	}
	out := new(VSphereBundle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionsBundle) DeepCopyInto(out *VersionsBundle) {
	*out = *in
	in.EksD.DeepCopyInto(&out.EksD)
	in.CertManager.DeepCopyInto(&out.CertManager)
	in.ClusterAPI.DeepCopyInto(&out.ClusterAPI)
	in.Bootstrap.DeepCopyInto(&out.Bootstrap)
	in.ControlPlane.DeepCopyInto(&out.ControlPlane)
	in.VSphere.DeepCopyInto(&out.VSphere)
	in.CloudStack.DeepCopyInto(&out.CloudStack)
	in.Docker.DeepCopyInto(&out.Docker)
	in.Eksa.DeepCopyInto(&out.Eksa)
	in.Cilium.DeepCopyInto(&out.Cilium)
	out.Kindnetd = in.Kindnetd
	in.Flux.DeepCopyInto(&out.Flux)
	in.PackageController.DeepCopyInto(&out.PackageController)
	in.BottleRocketHostContainers.DeepCopyInto(&out.BottleRocketHostContainers)
	in.ExternalEtcdBootstrap.DeepCopyInto(&out.ExternalEtcdBootstrap)
	in.ExternalEtcdController.DeepCopyInto(&out.ExternalEtcdController)
	in.Tinkerbell.DeepCopyInto(&out.Tinkerbell)
	in.Haproxy.DeepCopyInto(&out.Haproxy)
	in.Snow.DeepCopyInto(&out.Snow)
	in.Nutanix.DeepCopyInto(&out.Nutanix)
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(AwsBundle)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionsBundle.
func (in *VersionsBundle) DeepCopy() *VersionsBundle {
	if in == nil {
		return nil
	}
	out := new(VersionsBundle)
	in.DeepCopyInto(out)
	return out
}
