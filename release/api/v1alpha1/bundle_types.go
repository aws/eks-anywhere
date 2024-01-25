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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BundlesSpec defines the desired state of Bundles.
type BundlesSpec struct {
	// Monotonically increasing release number
	Number          int              `json:"number"`
	CliMinVersion   string           `json:"cliMinVersion"`
	CliMaxVersion   string           `json:"cliMaxVersion"`
	VersionsBundles []VersionsBundle `json:"versionsBundles"`
}

// BundlesStatus defines the observed state of Bundles.
type BundlesStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Bundles is the Schema for the bundles API.
type Bundles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BundlesSpec   `json:"spec,omitempty"`
	Status BundlesStatus `json:"status,omitempty"`
}

func (b *Bundles) DefaultEksAToolsImage() Image {
	return b.Spec.VersionsBundles[0].Eksa.CliTools
}

//+kubebuilder:object:root=true

// BundlesList contains a list of Bundles.
type BundlesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bundles `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bundles{}, &BundlesList{})
}

type VersionsBundle struct {
	KubeVersion                string                           `json:"kubeVersion"`
	EksD                       EksDRelease                      `json:"eksD"`
	CertManager                CertManagerBundle                `json:"certManager"`
	ClusterAPI                 CoreClusterAPI                   `json:"clusterAPI"`
	Bootstrap                  KubeadmBootstrapBundle           `json:"bootstrap"`
	ControlPlane               KubeadmControlPlaneBundle        `json:"controlPlane"`
	VSphere                    VSphereBundle                    `json:"vSphere"`
	CloudStack                 CloudStackBundle                 `json:"cloudStack,omitempty"`
	Docker                     DockerBundle                     `json:"docker"`
	Eksa                       EksaBundle                       `json:"eksa"`
	Cilium                     CiliumBundle                     `json:"cilium"`
	Kindnetd                   KindnetdBundle                   `json:"kindnetd"`
	Flux                       FluxBundle                       `json:"flux"`
	PackageController          PackageBundle                    `json:"packageController"`
	BottleRocketHostContainers BottlerocketHostContainersBundle `json:"bottlerocketHostContainers"`
	ExternalEtcdBootstrap      EtcdadmBootstrapBundle           `json:"etcdadmBootstrap"`
	ExternalEtcdController     EtcdadmControllerBundle          `json:"etcdadmController"`
	Tinkerbell                 TinkerbellBundle                 `json:"tinkerbell,omitempty"`
	Haproxy                    HaproxyBundle                    `json:"haproxy,omitempty"`
	Snow                       SnowBundle                       `json:"snow,omitempty"`
	Nutanix                    NutanixBundle                    `json:"nutanix,omitempty"`
	Upgrader                   UpgraderBundle                   `json:"upgrader,omitempty"`
	// This field has been deprecated
	Aws *AwsBundle `json:"aws,omitempty"`
}

type EksDRelease struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	// Release branch of the EKS-D release like 1-19, 1-20
	ReleaseChannel string `json:"channel,omitempty"`

	// +kubebuilder:validation:Required
	// Release number of EKS-D release
	KubeVersion string `json:"kubeVersion,omitempty"`

	// +kubebuilder:validation:Required
	// Url pointing to the EKS-D release manifest using which
	// assets where created
	EksDReleaseUrl string `json:"manifestUrl,omitempty"`

	// +kubebuilder:validation:Required
	// Git commit the component is built from, before any patches
	GitCommit string `json:"gitCommit,omitempty"`

	// KindNode points to a kind image built with this eks-d version
	KindNode Image `json:"kindNode,omitempty"`

	// Ami points to a collection of AMIs built with this eks-d version
	Ami OSImageBundle `json:"ami,omitempty"`

	// Ova points to a collection of OVAs built with this eks-d version
	Ova OSImageBundle `json:"ova,omitempty"`

	// Raw points to a collection of Raw images built with this eks-d version
	Raw OSImageBundle `json:"raw,omitempty"`

	// Components refers to the url that points to the EKS-D release CRD
	Components string `json:"components,omitempty"`

	// Etcdadm points to the etcdadm binary/tarball built for this eks-d kube version
	Etcdadm Archive `json:"etcdadm,omitempty"`

	// Crictl points to the crictl binary/tarball built for this eks-d kube version
	Crictl Archive `json:"crictl,omitempty"`

	// ImageBuilder points to the image-builder binary used to build eks-D based node images
	ImageBuilder Archive `json:"imagebuilder,omitempty"`

	// Containerd points to the containerd binary baked into this eks-D based node image
	Containerd Archive `json:"containerd,omitempty"`
}

// UpgraderBundle is a In-place Kubernetes version upgrader bundle.
type UpgraderBundle struct {
	Upgrader Image `json:"upgrader"`
}

type OSImageBundle struct {
	Bottlerocket Archive `json:"bottlerocket,omitempty"`
}

type BottlerocketHostContainersBundle struct {
	Admin            Image `json:"admin"`
	Control          Image `json:"control"`
	KubeadmBootstrap Image `json:"kubeadmBootstrap"`
}

type CertManagerBundle struct {
	Version    string   `json:"version,omitempty"`
	Acmesolver Image    `json:"acmesolver"`
	Cainjector Image    `json:"cainjector"`
	Controller Image    `json:"controller"`
	Ctl        Image    `json:"ctl"`
	Webhook    Image    `json:"webhook"`
	Manifest   Manifest `json:"manifest"`
}

type CoreClusterAPI struct {
	Version    string   `json:"version"`
	Controller Image    `json:"controller"`
	KubeProxy  Image    `json:"kubeProxy"`
	Components Manifest `json:"components"`
	Metadata   Manifest `json:"metadata"`
}

type KubeadmBootstrapBundle struct {
	Version    string   `json:"version"`
	Controller Image    `json:"controller"`
	KubeProxy  Image    `json:"kubeProxy"`
	Components Manifest `json:"components"`
	Metadata   Manifest `json:"metadata"`
}

type KubeadmControlPlaneBundle struct {
	Version    string   `json:"version"`
	Controller Image    `json:"controller"`
	KubeProxy  Image    `json:"kubeProxy"`
	Components Manifest `json:"components"`
	Metadata   Manifest `json:"metadata"`
}

type AwsBundle struct {
	Version         string   `json:"version"`
	Controller      Image    `json:"controller"`
	KubeProxy       Image    `json:"kubeProxy"`
	Components      Manifest `json:"components"`
	ClusterTemplate Manifest `json:"clusterTemplate"`
	Metadata        Manifest `json:"metadata"`
}

type VSphereBundle struct {
	Version              string   `json:"version"`
	ClusterAPIController Image    `json:"clusterAPIController"`
	KubeProxy            Image    `json:"kubeProxy"`
	Manager              Image    `json:"manager"`
	KubeVip              Image    `json:"kubeVip"`
	Components           Manifest `json:"components"`
	Metadata             Manifest `json:"metadata"`
	ClusterTemplate      Manifest `json:"clusterTemplate"`
	// This field has been deprecated
	Driver *Image `json:"driver,omitempty"`
	// This field has been deprecated
	Syncer *Image `json:"syncer,omitempty"`
}

type DockerBundle struct {
	Version         string   `json:"version"`
	Manager         Image    `json:"manager"`
	KubeProxy       Image    `json:"kubeProxy"`
	Components      Manifest `json:"components"`
	ClusterTemplate Manifest `json:"clusterTemplate"`
	Metadata        Manifest `json:"metadata"`
}

type CloudStackBundle struct {
	Version              string   `json:"version"`
	ClusterAPIController Image    `json:"clusterAPIController"`
	KubeRbacProxy        Image    `json:"kubeRbacProxy"`
	KubeVip              Image    `json:"kubeVip"`
	Components           Manifest `json:"components"`
	Metadata             Manifest `json:"metadata"`
}

type CiliumBundle struct {
	Version   string   `json:"version,omitempty"`
	Cilium    Image    `json:"cilium"`
	Operator  Image    `json:"operator"`
	Manifest  Manifest `json:"manifest"`
	HelmChart Image    `json:"helmChart,omitempty"`
}

type KindnetdBundle struct {
	Version  string   `json:"version,omitempty"`
	Manifest Manifest `json:"manifest"`
}

type FluxBundle struct {
	Version                string `json:"version,omitempty"`
	SourceController       Image  `json:"sourceController"`
	KustomizeController    Image  `json:"kustomizeController"`
	HelmController         Image  `json:"helmController"`
	NotificationController Image  `json:"notificationController"`
}

type PackageBundle struct {
	Version                   string `json:"version,omitempty"`
	Controller                Image  `json:"packageController"`
	TokenRefresher            Image  `json:"tokenRefresher"`
	CredentialProviderPackage Image  `json:"credentialProviderPackage,omitempty"`
	HelmChart                 Image  `json:"helmChart,omitempty"`
}

type EksaBundle struct {
	Version             string   `json:"version,omitempty"`
	CliTools            Image    `json:"cliTools"`
	ClusterController   Image    `json:"clusterController"`
	DiagnosticCollector Image    `json:"diagnosticCollector"`
	Components          Manifest `json:"components"`
}

type EtcdadmBootstrapBundle struct {
	Version    string   `json:"version"`
	Controller Image    `json:"controller"`
	KubeProxy  Image    `json:"kubeProxy"`
	Components Manifest `json:"components"`
	Metadata   Manifest `json:"metadata"`
}

type EtcdadmControllerBundle struct {
	Version    string   `json:"version"`
	Controller Image    `json:"controller"`
	KubeProxy  Image    `json:"kubeProxy"`
	Components Manifest `json:"components"`
	Metadata   Manifest `json:"metadata"`
}

type TinkerbellStackBundle struct {
	Actions        ActionsBundle `json:"actions"`
	Boots          Image         `json:"boots"`
	Hegel          Image         `json:"hegel"`
	TinkebellChart Image         `json:"tinkerbellChart"`
	Hook           HookBundle    `json:"hook"`
	Rufio          Image         `json:"rufio"`
	Tink           TinkBundle    `json:"tink"`
}

// Tinkerbell Template Actions.
type ActionsBundle struct {
	Cexec       Image `json:"cexec"`
	Kexec       Image `json:"kexec"`
	ImageToDisk Image `json:"imageToDisk"`
	OciToDisk   Image `json:"ociToDisk"`
	WriteFile   Image `json:"writeFile"`
	Reboot      Image `json:"reboot"`
}

type TinkBundle struct {
	TinkController Image `json:"tinkController"`
	TinkServer     Image `json:"tinkServer"`
	TinkWorker     Image `json:"tinkWorker"`
}

// Tinkerbell hook OS.
type HookBundle struct {
	Bootkit   Image    `json:"bootkit"`
	Docker    Image    `json:"docker"`
	Kernel    Image    `json:"kernel"`
	Initramfs HookArch `json:"initramfs"`
	Vmlinuz   HookArch `json:"vmlinuz"`
}

type HookArch struct {
	Arm Archive `json:"arm"`
	Amd Archive `json:"amd"`
}

type TinkerbellBundle struct {
	Version              string                `json:"version"`
	ClusterAPIController Image                 `json:"clusterAPIController"`
	KubeVip              Image                 `json:"kubeVip"`
	Envoy                Image                 `json:"envoy"`
	Components           Manifest              `json:"components"`
	Metadata             Manifest              `json:"metadata"`
	ClusterTemplate      Manifest              `json:"clusterTemplate"`
	TinkerbellStack      TinkerbellStackBundle `json:"tinkerbellStack,omitempty"`
}

type HaproxyBundle struct {
	Image Image `json:"image"`
}

type SnowBundle struct {
	Version                   string   `json:"version"`
	Manager                   Image    `json:"manager"`
	KubeVip                   Image    `json:"kubeVip"`
	Components                Manifest `json:"components"`
	Metadata                  Manifest `json:"metadata"`
	BottlerocketBootstrapSnow Image    `json:"bottlerocketBootstrapSnow"`
}

type NutanixBundle struct {
	ClusterAPIController Image    `json:"clusterAPIController"`
	CloudProvider        Image    `json:"cloudProvider,omitempty"`
	Version              string   `json:"version"`
	KubeVip              Image    `json:"kubeVip"`
	Components           Manifest `json:"components"`
	Metadata             Manifest `json:"metadata"`
	ClusterTemplate      Manifest `json:"clusterTemplate"`
}
