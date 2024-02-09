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

func (vb *VersionsBundle) Manifests() map[string][]*string {
	return map[string][]*string{
		"core-cluster-api": {
			&vb.ClusterAPI.Components.URI,
			&vb.ClusterAPI.Metadata.URI,
		},
		"capi-kubeadm-bootstrap": {
			&vb.Bootstrap.Components.URI,
			&vb.Bootstrap.Metadata.URI,
		},
		"capi-kubeadm-control-plane": {
			&vb.ControlPlane.Components.URI,
			&vb.ControlPlane.Metadata.URI,
		},
		"cert-manager": {
			&vb.CertManager.Manifest.URI,
		},
		"cluster-api-provider-docker": {
			&vb.Docker.Components.URI,
			&vb.Docker.ClusterTemplate.URI,
			&vb.Docker.Metadata.URI,
		},
		"cluster-api-provider-vsphere": {
			&vb.VSphere.Components.URI,
			&vb.VSphere.ClusterTemplate.URI,
			&vb.VSphere.Metadata.URI,
		},
		"cluster-api-provider-cloudstack": {
			&vb.CloudStack.Components.URI,
			&vb.CloudStack.Metadata.URI,
		},
		"cluster-api-provider-tinkerbell": {
			&vb.Tinkerbell.Components.URI,
			&vb.Tinkerbell.ClusterTemplate.URI,
			&vb.Tinkerbell.Metadata.URI,
		},
		"cluster-api-provider-snow": {
			&vb.Snow.Components.URI,
			&vb.Snow.Metadata.URI,
		},
		"cluster-api-provider-nutanix": {
			&vb.Nutanix.Components.URI,
			&vb.Nutanix.ClusterTemplate.URI,
			&vb.Nutanix.Metadata.URI,
		},
		"cilium": {
			&vb.Cilium.Manifest.URI,
		},
		"kindnetd": {
			&vb.Kindnetd.Manifest.URI,
		},
		"eks-anywhere-cluster-controller": {
			&vb.Eksa.Components.URI,
		},
		"etcdadm-bootstrap-provider": {
			&vb.ExternalEtcdBootstrap.Components.URI,
			&vb.ExternalEtcdBootstrap.Metadata.URI,
		},
		"etcdadm-controller": {
			&vb.ExternalEtcdController.Components.URI,
			&vb.ExternalEtcdController.Metadata.URI,
		},
		"eks-distro": {
			&vb.EksD.Components,
			&vb.EksD.EksDReleaseUrl,
		},
	}
}

func (vb *VersionsBundle) Ovas() []Archive {
	return []Archive{
		vb.EksD.Ova.Bottlerocket,
	}
}

func (vb *VersionsBundle) CloudStackImages() []Image {
	return []Image{
		vb.CloudStack.ClusterAPIController,
		vb.CloudStack.KubeRbacProxy,
		vb.CloudStack.KubeVip,
	}
}

func (vb *VersionsBundle) VsphereImages() []Image {
	return []Image{
		vb.VSphere.ClusterAPIController,
		vb.VSphere.KubeProxy,
		vb.VSphere.KubeVip,
		vb.VSphere.Manager,
	}
}

func (vb *VersionsBundle) DockerImages() []Image {
	return []Image{
		vb.Docker.KubeProxy,
		vb.Docker.Manager,
	}
}

func (vb *VersionsBundle) SnowImages() []Image {
	i := make([]Image, 0, 2)
	if vb.Snow.KubeVip.URI != "" {
		i = append(i, vb.Snow.KubeVip)
	}
	if vb.Snow.Manager.URI != "" {
		i = append(i, vb.Snow.Manager)
	}
	if vb.Snow.BottlerocketBootstrapSnow.URI != "" {
		i = append(i, vb.Snow.BottlerocketBootstrapSnow)
	}

	return i
}

func (vb *VersionsBundle) TinkerbellImages() []Image {
	return []Image{
		vb.Tinkerbell.ClusterAPIController,
		vb.Tinkerbell.KubeVip,
		vb.Tinkerbell.Envoy,
		vb.Tinkerbell.TinkerbellStack.Actions.Cexec,
		vb.Tinkerbell.TinkerbellStack.Actions.Kexec,
		vb.Tinkerbell.TinkerbellStack.Actions.ImageToDisk,
		vb.Tinkerbell.TinkerbellStack.Actions.OciToDisk,
		vb.Tinkerbell.TinkerbellStack.Actions.WriteFile,
		vb.Tinkerbell.TinkerbellStack.Actions.Reboot,
		vb.Tinkerbell.TinkerbellStack.Boots,
		vb.Tinkerbell.TinkerbellStack.Hegel,
		vb.Tinkerbell.TinkerbellStack.Hook.Bootkit,
		vb.Tinkerbell.TinkerbellStack.Hook.Docker,
		vb.Tinkerbell.TinkerbellStack.Hook.Kernel,
		vb.Tinkerbell.TinkerbellStack.Rufio,
		vb.Tinkerbell.TinkerbellStack.Tink.TinkController,
		vb.Tinkerbell.TinkerbellStack.Tink.TinkServer,
		vb.Tinkerbell.TinkerbellStack.Tink.TinkWorker,
	}
}

func (vb *VersionsBundle) NutanixImages() []Image {
	i := make([]Image, 0, 2)
	if vb.Nutanix.ClusterAPIController.URI != "" {
		i = append(i, vb.Nutanix.ClusterAPIController)
	}

	if vb.Nutanix.CloudProvider.URI != "" {
		i = append(i, vb.Nutanix.CloudProvider)
	}

	return i
}

func (vb *VersionsBundle) SharedImages() []Image {
	return []Image{
		vb.Bootstrap.Controller,
		vb.Bootstrap.KubeProxy,
		vb.BottleRocketHostContainers.Admin,
		vb.BottleRocketHostContainers.Control,
		vb.BottleRocketHostContainers.KubeadmBootstrap,
		vb.CertManager.Acmesolver,
		vb.CertManager.Cainjector,
		vb.CertManager.Controller,
		vb.CertManager.Ctl,
		vb.CertManager.Webhook,
		vb.Cilium.Cilium,
		vb.Cilium.Operator,
		vb.ClusterAPI.Controller,
		vb.ClusterAPI.KubeProxy,
		vb.ControlPlane.Controller,
		vb.ControlPlane.KubeProxy,
		vb.EksD.KindNode,
		vb.Eksa.CliTools,
		vb.Eksa.ClusterController,
		vb.Eksa.DiagnosticCollector,
		vb.Flux.HelmController,
		vb.Flux.KustomizeController,
		vb.Flux.NotificationController,
		vb.Flux.SourceController,
		vb.ExternalEtcdBootstrap.Controller,
		vb.ExternalEtcdBootstrap.KubeProxy,
		vb.ExternalEtcdController.Controller,
		vb.ExternalEtcdController.KubeProxy,
		vb.Haproxy.Image,
		vb.PackageController.Controller,
		vb.PackageController.TokenRefresher,
		vb.Upgrader.Upgrader,
	}
}

func (vb *VersionsBundle) Images() []Image {
	groupedImages := [][]Image{
		vb.SharedImages(),
		vb.DockerImages(),
		vb.VsphereImages(),
		vb.CloudStackImages(),
		vb.SnowImages(),
		vb.TinkerbellImages(),
		vb.NutanixImages(),
	}

	size := 0
	for _, g := range groupedImages {
		size += len(g)
	}

	images := make([]Image, 0, size)
	for _, g := range groupedImages {
		images = append(images, g...)
	}

	return images
}

func (vb *VersionsBundle) Charts() map[string]*Image {
	return map[string]*Image{
		"cilium":                &vb.Cilium.HelmChart,
		"eks-anywhere-packages": &vb.PackageController.HelmChart,
		"tinkerbell-chart":      &vb.Tinkerbell.TinkerbellStack.TinkebellChart,
	}
}
