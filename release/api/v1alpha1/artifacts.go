package v1alpha1

func (vb *VersionsBundle) Manifests() map[string][]Manifest {
	manifests := map[string][]Manifest{}

	// CAPA manifests
	manifests["cluster-api-provider-aws"] = []Manifest{
		vb.Aws.Components,
		vb.Aws.ClusterTemplate,
		vb.Aws.Metadata,
	}

	// Core CAPI manifests
	manifests["core-cluster-api"] = []Manifest{
		vb.ClusterAPI.Components,
		vb.ClusterAPI.Metadata,
	}

	// CAPI Kubeadm bootstrap manifests
	manifests["capi-kubeadm-bootstrap"] = []Manifest{
		vb.Bootstrap.Components,
		vb.Bootstrap.Metadata,
	}

	// CAPI Kubeadm Controlplane manifests
	manifests["capi-kubeadm-control-plane"] = []Manifest{
		vb.ControlPlane.Components,
		vb.ControlPlane.Metadata,
	}

	// CAPD manifests
	manifests["cluster-api-provider-docker"] = []Manifest{
		vb.Docker.Components,
		vb.Docker.ClusterTemplate,
		vb.Docker.Metadata,
	}

	// CAPV manifests
	manifests["cluster-api-provider-vsphere"] = []Manifest{
		vb.VSphere.Components,
		vb.VSphere.ClusterTemplate,
		vb.VSphere.Metadata,
	}

	// Cilium manifest
	manifests["cilium"] = []Manifest{vb.Cilium.Manifest}

	// EKS Anywhere CRD manifest
	manifests["eks-anywhere-cluster-controller"] = []Manifest{vb.Eksa.Components}

	// Etcdadm bootstrap provider manifests
	manifests["etcdadm-bootstrap-provider"] = []Manifest{
		vb.ExternalEtcdBootstrap.Components,
		vb.ExternalEtcdBootstrap.Metadata,
	}

	// Etcdadm controller manifests
	manifests["etcdadm-controller"] = []Manifest{
		vb.ExternalEtcdController.Components,
		vb.ExternalEtcdController.Metadata,
	}

	return manifests
}

func (vb *VersionsBundle) Ovas() []Archive {
	return []Archive{
		vb.EksD.Ova.Bottlerocket.Archive,
		vb.EksD.Ova.Ubuntu.Archive,
	}
}

func (vb *VersionsBundle) VsphereImages() []Image {
	images := []Image{}
	images = append(images, vb.VSphere.ClusterAPIController)
	images = append(images, vb.VSphere.Driver)
	images = append(images, vb.VSphere.KubeProxy)
	images = append(images, vb.VSphere.KubeVip)
	images = append(images, vb.VSphere.Manager)
	images = append(images, vb.VSphere.Syncer)

	return images
}

func (vb *VersionsBundle) DockerImages() []Image {
	images := []Image{}
	images = append(images, vb.Docker.KubeProxy)
	images = append(images, vb.Docker.Manager)

	return images
}

func (vb *VersionsBundle) SharedImages() []Image {
	images := []Image{}
	images = append(images, vb.Bootstrap.Controller)
	images = append(images, vb.Bootstrap.KubeProxy)

	images = append(images, vb.BottleRocketBootstrap.Bootstrap)
	images = append(images, vb.BottleRocketAdmin.Admin)

	images = append(images, vb.CertManager.Acmesolver)
	images = append(images, vb.CertManager.Cainjector)
	images = append(images, vb.CertManager.Controller)
	images = append(images, vb.CertManager.Webhook)

	images = append(images, vb.Cilium.Cilium)
	images = append(images, vb.Cilium.Operator)

	images = append(images, vb.ClusterAPI.Controller)
	images = append(images, vb.ClusterAPI.KubeProxy)

	images = append(images, vb.ControlPlane.Controller)
	images = append(images, vb.ControlPlane.KubeProxy)

	images = append(images, vb.EksD.KindNode)
	images = append(images, vb.Eksa.CliTools)
	images = append(images, vb.Eksa.ClusterController)

	images = append(images, vb.Flux.HelmController)
	images = append(images, vb.Flux.KustomizeController)
	images = append(images, vb.Flux.NotificationController)
	images = append(images, vb.Flux.SourceController)

	images = append(images, vb.ExternalEtcdBootstrap.Controller)
	images = append(images, vb.ExternalEtcdBootstrap.KubeProxy)

	images = append(images, vb.ExternalEtcdController.Controller)
	images = append(images, vb.ExternalEtcdController.KubeProxy)

	return images
}

func (vb *VersionsBundle) Images() []Image {
	images := []Image{}
	images = append(images, vb.SharedImages()...)
	images = append(images, vb.DockerImages()...)
	images = append(images, vb.VsphereImages()...)

	return images
}
