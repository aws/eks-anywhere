package v1alpha1

func (vb *VersionsBundle) Manifests() map[string][]Manifest {
	return map[string][]Manifest{
		"cluster-api-provider-aws": {
			vb.Aws.Components,
			vb.Aws.ClusterTemplate,
			vb.Aws.Metadata,
		},
		"core-cluster-api": {
			vb.ClusterAPI.Components,
			vb.ClusterAPI.Metadata,
		},
		"capi-kubeadm-bootstrap": {
			vb.Bootstrap.Components,
			vb.Bootstrap.Metadata,
		},
		"capi-kubeadm-control-plane": {
			vb.ControlPlane.Components,
			vb.ControlPlane.Metadata,
		},
		"cluster-api-provider-docker": {
			vb.Docker.Components,
			vb.Docker.ClusterTemplate,
			vb.Docker.Metadata,
		},
		"cluster-api-provider-vsphere": {
			vb.VSphere.Components,
			vb.VSphere.ClusterTemplate,
			vb.VSphere.Metadata,
		},
		"cluster-api-provider-cloudstack": {
			vb.CloudStack.Components,
			vb.CloudStack.Metadata,
		},
		"cilium": {
			vb.Cilium.Manifest,
		},
		"kindnetd": {
			vb.Kindnetd.Manifest,
		},
		"eks-anywhere-cluster-controller": {
			vb.Eksa.Components,
		},
		"etcdadm-bootstrap-provider": {
			vb.ExternalEtcdBootstrap.Components,
			vb.ExternalEtcdBootstrap.Metadata,
		},
		"etcdadm-controller": {
			vb.ExternalEtcdController.Components,
			vb.ExternalEtcdController.Metadata,
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
		vb.CloudStack.KubeProxy,
		vb.CloudStack.Manager,
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
