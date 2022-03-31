package v1alpha1

func (vb *VersionsBundle) Manifests() map[string][]*string {
	return map[string][]*string{
		"cluster-api-provider-aws": {
			&vb.Aws.Components.URI,
			&vb.Aws.ClusterTemplate.URI,
			&vb.Aws.Metadata.URI,
		},
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
		vb.EksD.Ova.Bottlerocket.Archive,
		vb.EksD.Ova.Ubuntu.Archive,
	}
}

func (vb *VersionsBundle) CloudStackImages() []Image {
	return []Image{
		vb.CloudStack.ClusterAPIController,
		vb.CloudStack.KubeVip,
	}
}

func (vb *VersionsBundle) VsphereImages() []Image {
	return []Image{
		vb.VSphere.ClusterAPIController,
		vb.VSphere.Driver,
		vb.VSphere.KubeProxy,
		vb.VSphere.KubeVip,
		vb.VSphere.Manager,
		vb.VSphere.Syncer,
	}
}

func (vb *VersionsBundle) DockerImages() []Image {
	return []Image{
		vb.Docker.KubeProxy,
		vb.Docker.Manager,
	}
}

func (vb *VersionsBundle) SharedImages() []Image {
	return []Image{
		vb.Bootstrap.Controller,
		vb.Bootstrap.KubeProxy,
		vb.BottleRocketBootstrap.Bootstrap,
		vb.BottleRocketAdmin.Admin,
		vb.CertManager.Acmesolver,
		vb.CertManager.Cainjector,
		vb.CertManager.Controller,
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
		vb.Flux.HelmController,
		vb.Flux.KustomizeController,
		vb.Flux.NotificationController,
		vb.Flux.SourceController,
		vb.ExternalEtcdBootstrap.Controller,
		vb.ExternalEtcdBootstrap.KubeProxy,
		vb.ExternalEtcdController.Controller,
		vb.ExternalEtcdController.KubeProxy,
		vb.Haproxy.Image,
	}
}

func (vb *VersionsBundle) Images() []Image {
	shared := vb.SharedImages()
	docker := vb.DockerImages()
	vsphere := vb.VsphereImages()
	cloudstack := vb.CloudStackImages()

	images := make([]Image, 0, len(shared)+len(docker)+len(vsphere)+len(cloudstack))
	images = append(images, shared...)
	images = append(images, docker...)
	images = append(images, vsphere...)
	images = append(images, cloudstack...)

	return images
}

func (vb *VersionsBundle) Charts() map[string]*Image {
	return map[string]*Image{
		"cilium":                &vb.Cilium.HelmChart,
		"eks-anywhere-packages": &vb.PackageController.HelmChart,
	}
}
