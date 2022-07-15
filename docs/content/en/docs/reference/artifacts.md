---
title: "Artifacts"
linkTitle: "Artifacts"
weight: 55
description: >
  Artifacts associated with this release: OVAs and images.
---

# Bare Metal artifacts

Artifacts for EKS Anyware Bare Metal clusters are listed below.
If you like, you can download these images and serve them locally to speed up cluster creation.
See descriptions of the [osImageURL]({{< relref "./clusterspec/baremetal/#osimageurl" >}}) and [`hookImagesURLPath`]({{< relref "./clusterspec/baremetal/#hookimagesurlpath" >}}) fields for details.

## Ubuntu OS images for Bare Metal

Kubernetes 1.20:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/raw/1-20/ubuntu-v1.20.15-eks-d-1-20-18-eks-a-12-amd64.gz
```

Kubernetes 1.21:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/raw/1-21/ubuntu-v1.21.13-eks-d-1-21-16-eks-a-12-amd64.gz
```

Kubernetes 1.22:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/raw/1-22/ubuntu-v1.22.10-eks-d-1-22-9-eks-a-12-amd64.gz
```

## Bottlerocket OS images for Bare Metal

Kubernetes 1.21:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/raw/1-21/bottlerocket-v1.21.13-eks-d-1-21-16-eks-a-12-amd64.img.gz
```

Kubernetes 1.22:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/raw/1-22/bottlerocket-v1.22.10-eks-d-1-22-9-eks-a-12-amd64.img.gz
```

## HookOS (kernel and initial ramdisk) for Bare Metal

kernel:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/vmlinuz-x86_64
```

initial ramdisk:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/initramfs-x86_64
```

# vSphere OVAs

## Bottlerocket OVAs

Bottlerocket vends its VMware variant OVAs using a secure distribution tool called tuftool. Please follow instructions down below to
download Bottlerocket OVA.
1. Install Rust and Cargo
```
curl https://sh.rustup.rs -sSf | sh
```
2. Install tuftool using Cargo
```
CARGO_NET_GIT_FETCH_WITH_CLI=true cargo install --force tuftool
```
3. Download the root role tuftool will use to download the OVA
```
curl -O "https://cache.bottlerocket.aws/root.json"
sha512sum -c <<<"e9b1ea5f9b4f95c9b55edada4238bf00b12845aa98bdd2d3edb63ff82a03ada19444546337ec6d6806cbf329027cf49f7fde31f54d551c5e02acbed7efe75785  root.json"
```
4. Export the desired Kubernetes Version. EKS Anywhere currently supports 1.22, 1.21 and 1.20
```
export KUBEVERSION="1.22"
```
5. Download the OVA
```
OVA="bottlerocket-vmware-k8s-${KUBEVERSION}-x86_64-v1.8.0.ova"
tuftool download ${TMPDIR:-/tmp/bottlerocket-ovas} --target-name "${OVA}" \
   --root ./root.json \
   --metadata-url "https://updates.bottlerocket.aws/2020-07-07/vmware-k8s-${KUBEVERSION}/x86_64/" \
   --targets-url "https://updates.bottlerocket.aws/targets/"
```

Bottlerocket Tags

OS Family - `os:bottlerocket`

EKS-D Release

1.22 - `eksdRelease:kubernetes-1-22-eks-9`

1.21 - `eksdRelease:kubernetes-1-21-eks-16`

1.20 - `eksdRelease:kubernetes-1-20-eks-18`

## Ubuntu with Kubernetes 1.22 OVA

* https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/ova/1-22/ubuntu-v1.22.10-eks-d-1-22-9-eks-a-12-amd64.ova
* `os:ubuntu`
* `eksdRelease:kubernetes-1-22-eks-9`

## Ubuntu with Kubernetes 1.21 OVA

* https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/ova/1-21/ubuntu-v1.21.13-eks-d-1-21-16-eks-a-12-amd64.ova
* `os:ubuntu`
* `eksdRelease:kubernetes-1-21-eks-16`

## Ubuntu with Kubernetes 1.20 OVA

* https://anywhere-assets.eks.amazonaws.com/releases/bundles/12/artifacts/ova/1-20/ubuntu-v1.20.15-eks-d-1-20-18-eks-a-12-amd64.ova
* `os:ubuntu`
* `eksdRelease:kubernetes-1-20-eks-18`

## Building your own Ubuntu OVA for vSphere
The EKS Anywhere project OVA building process leverages upstream [image-builder repository.](https://github.com/kubernetes-sigs/image-builder)
If you want to build an OVA with a custom Ubuntu base image to use for an EKS Anywhere cluster, please follow the instructions below.

Having access to a vSphere environment and docker running locally are prerequisites for building your own images.

### Required vSphere Permissions
#### Virtual machine
Inventory:
* Create new

Configuration:
* Change configuration
* Add new disk
* Add or remove device
* Change memory
* Change settings
* Set annotation

Interaction:
* Power on
* Power off
* Console interaction
* Configure CD media
* Device connection

Snapshot management:
* Create snapshot

Provisioning
* Mark as template

#### Resource Pool
* Assign vm to resource pool

#### Datastore
* Allocate space
* Browse data
* Low level file operations

#### Network
* Assign network to vm

### Steps to build an OVA
1. Spin up a builder-base docker container and exec into it. Please use the most recent tag for the image on its repository [here](https://gallery.ecr.aws/eks-distro-build-tooling/builder-base)
```
docker exec -it public.ecr.aws/eks-distro-build-tooling/builder-base:latest bash
```
2. Clone the [eks-anywhere-build-tooling repo.](https://github.com/aws/eks-anywhere-build-tooling)
```
git clone https://github.com/aws/eks-anywhere-build-tooling.git
```
3. Navigate to the image-builder directory.
```
cd eks-anywhere-build-tooling/projects/kubernetes-sigs/image-builder
```
4. Get the vSphere connection details and create a json file named `vsphere.json` with the following template.
```
{
    "cluster": "<vSphere cluster name>",
    "datacenter": "<datacenter name on vSphere>",
    "datastore": "<datastore to be used on vSphere>",
    "folder": "<folder path to use for building ova>",
    "network": "<dhcp enabled network name>",
    "resource_pool": "<vSphere resource pool to use>",
    "vcenter_server": "<vSphere server URL>",
    "username": "<vSphere username>",
    "password": "<vSphere password>",
    "template": "",
    "insecure_connection": "false",
    "linked_clone": "false",
    "convert_to_template": "false",
    "create_snapshot": "true"
}

```
4. Export the vSphere connection data file, escaping all the quotes
```
export VSPHERE_CONNECTION_DATA=\"$(cat vsphere.json | jq -c . | sed 's/"/\\"/g')\"
```
5. Download the most recent release bundle manifest and get the latest URLs for `etcdadm` and `crictl` for the intended Kubernetes version.
```
wget https://anywhere-assets.eks.amazonaws.com/bundle-release.yaml
```
7. Export the CRICTL_URL and ETCADM_HTTP_SOURCE environment variables with the URLs from previous step.
```
export CRICTL_URL=<crictl url>
export ETCDADM_HTTP_SOURCE=<etcdadm url>
```
7. Create a library on vSphere for image-builder.
```
govc library.create "CodeBuild"
```
8. Update the Ubuntu configuration file with the new custom ISO URL and its checksum at
`image-builder/images/capi/packer/ova/ubuntu-2004.json`
9. Setup image-builder and run the OVA build for the Kubernetes version.
```
RELEASE_BRANCH=1-22 make release-ova-ubuntu-2004
```

# Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
