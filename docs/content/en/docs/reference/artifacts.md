---
title: "Artifacts"
linkTitle: "Artifacts"
weight: 55
description: >
  Artifacts associated with this release: OVAs and images.
---

EKS Anywhere supports two different node operating systems:

* Bottlerocket
* Ubuntu

Bottlerocket OVAs and images are distributed by the EKS Anywhere project.
To build your own Ubuntu-based EKS Anywhere node, see [Building Ubuntu-based node images]({{< relref "#building-ubuntu-based-node-images">}}).

## Bare Metal artifacts

Artifacts for EKS Anyware Bare Metal clusters are listed below.
If you like, you can download these images and serve them locally to speed up cluster creation.
See descriptions of the [osImageURL]({{< relref "./clusterspec/baremetal/#osimageurl" >}}) and [`hookImagesURLPath`]({{< relref "./clusterspec/baremetal/#hookimagesurlpath" >}}) fields for details.

### Ubuntu OS images for Bare Metal

EKS Anywhere no longer distributes Ubuntu OS images.
However, see [Building Ubuntu-based node images]({{< relref "#building-ubuntu-based-node-images">}}) for information on how to build your own Ubuntu-based image to use with EKS Anywhere.

### Bottlerocket OS images for Bare Metal

Bottlerocket vends its Baremetal variant Images using a secure distribution tool called tuftool. Please refer [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket image.

### HookOS (kernel and initial ramdisk) for Bare Metal

kernel:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/14/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/vmlinuz-x86_64
```

initial ramdisk:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/14/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/initramfs-x86_64
```

## vSphere OVAs

### Bottlerocket OVAs

Bottlerocket vends its VMware variant OVAs using a secure distribution tool called tuftool. Please refer [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket OVA.

Bottlerocket Tags

OS Family - `os:bottlerocket`

EKS-D Release

1.23 - `eksdRelease:kubernetes-1-23-eks-5`

1.22 - `eksdRelease:kubernetes-1-22-eks-10`

1.21 - `eksdRelease:kubernetes-1-21-eks-18`

1.20 - `eksdRelease:kubernetes-1-20-eks-20`

### Ubuntu OVAs
EKS Anywhere no longer distributes Ubuntu OVAs for use with EKS Anywhere clusters.
Building your own Ubuntu-based nodes as described in [Building Ubuntu-based node images]({{< relref "#building-ubuntu-based-node-images">}}) is the only supported way to get that functionality.

## Download Bottlerocket node images
Bottlerocket vends its VMware variant OVAs and Baremetal variants images using a secure distribution tool called tuftool. Please follow instructions down below to
download Bottlerocket node images.
1. Install Rust and Cargo
```
curl https://sh.rustup.rs -sSf | sh
```
2. Install tuftool using Cargo
```
CARGO_NET_GIT_FETCH_WITH_CLI=true cargo install --force tuftool
```
3. Download the root role tuftool will use to download the bottlerocket images
```
curl -O "https://cache.bottlerocket.aws/root.json"
sha512sum -c <<<"b81af4d8eb86743539fbc4709d33ada7b118d9f929f0c2f6c04e1d41f46241ed80423666d169079d736ab79965b4dd25a5a6db5f01578b397496d49ce11a3aa2  root.json"
```
4. Export the desired Kubernetes Version. EKS Anywhere currently supports 1.23, 1.22, 1.21 and 1.20
```
export KUBEVERSION="1.23"
```
5. Download Bottlerocket node image

    a. To download VMware variant Bottlerocket OVA
    ```
    OVA="bottlerocket-vmware-k8s-${KUBEVERSION}-x86_64-v1.9.2.ova"
    tuftool download ${TMPDIR:-/tmp/bottlerocket-ovas} --target-name "${OVA}" \
       --root ./root.json \
       --metadata-url "https://updates.bottlerocket.aws/2020-07-07/vmware-k8s-${KUBEVERSION}/x86_64/" \
       --targets-url "https://updates.bottlerocket.aws/targets/"
    ```
   The above command will download Bottlerocket OVA. Please refer [Deploy an OVA Template]({{< relref "vsphere/vsphere-preparation/#deploy-an-ova-template">}}) to proceed with the downloaded OVA.

    b. To download Baremetal variant Bottlerocket image
    ```
    IMAGE="bottlerocket-metal-k8s-${KUBEVERSION}-x86_64-v1.9.0.img.lz4"
    tuftool download ${TMPDIR:-/tmp/bottlerocket-metal} --target-name "${IMAGE}" \
       --root ./root.json \
       --metadata-url "https://updates.bottlerocket.aws/2020-07-07/metal-k8s-${KUBEVERSION}/x86_64/" \
       --targets-url "https://updates.bottlerocket.aws/targets/"
    ```
   The above command will download Bottlerocket lz4 compressed image. Decompress and gzip the image with the following
   commands and host the image on a webserver for using it for an EKS Anywhere Baremetal cluster.
   ```
   lz4 --decompress ${TMPDIR:-/tmp/bottlerocket-metal}/${IMAGE} ${TMPDIR:-/tmp/bottlerocket-metal}/bottlerocket.img
   gzip ${TMPDIR:-/tmp/bottlerocket-metal}/bottlerocket.img
   ```

## Building Ubuntu-based node images

The `image-builder` CLI lets you build your own Ubuntu-based vSphere OVAs or Bare Metal gzip images to use in EKS Anywhere clusters.
When you run `image-builder` it will pull in all components needed to create images to use for nodes in an EKS Anywhere cluster, including the lastest Ubuntu, Kubernetes, and EKS Distro security updates, bug fixes, and patches.
With this tool, when you build an image you get to choose:

* Operating system type (for example, ubuntu)
* Provider (vsphere or baremetal)
* Release channel for EKS Distro (generally aligning with Kubernetes releases)
* vSphere only: configuration file providing information needed to access your vSphere setup

Because `image-builder` creates images in the same way that the EKS Anywhere project does for their own testing, images built with that tool are supported.
The following procedure describes how to use `image-builder` to build images for EKS Anywhere on a vSphere or Bare Metal provider.

### Prerequisites

To use `image-builder` you must meet the following prerequisites:

* Run on Ubuntu 22.04 or later 
* Machine requirements:
  * AMD 64-bit architecture
  * 50 GB disk space
  * 2 vCPUs
  * 8 GB RAM
  * Bare Metal only: Run on a bare metal machine with virtualization enabled
* Network access to:
  * vCenter endpoint (vSphere only)
  * public.ecr.aws (to download container images from EKS Anywhere)
  * anywhere-assets.eks.amazonaws.com (to download the EKS Anywhere binaries, manifests and OVAs)
  * distro.eks.amazonaws.com (to download EKS Distro binaries and manifests)
  * d2glxqk2uabbnd.cloudfront.net (for EKS Anywhere and EKS Distro ECR container images)
  * api.ecr.us-west-2.amazonaws.com (for EKS Anywhere package authentication matching your region)
  * d5l0dvt14r5h8.cloudfront.net (for EKS Anywhere package ECR container)
* vSphere only:
  * Required vSphere user permissions:
    * Inventory:
      * Create new
    * Configuration:
      * Change configuration
      * Add new disk
      * Add or remove device
      * Change memory
      * Change settings
      * Set annotation
    * Interaction:
      * Power on
      * Power off
      * Console interaction
      * Configure CD media
      * Device connection
    * Snapshot management:
      * Create snapshot
    * Provisioning
      * Mark as template
    * Resource Pool
      * Assign vm to resource pool
    * Datastore
      * Allocate space
      * Browse data
      * Low level file operations
    * Network
      * Assign network to vm

### Optional Proxy configuration
You can use a proxy server to route outbound requests to the internet. To configure `image-builder` tool to use a proxy server, export these proxy environment variables:
  ```
  export HTTP_PROXY=<HTTP proxy URL e.g. http://proxy.corp.com:80>
  export HTTPS_PROXY=<HTTPS proxy URL e.g. http://proxy.corp.com:443>
  export NO_PROXY=<No proxy>
  ```

### Build vSphere OVA node images

1. Create a linux user for running image-builder.
   ```
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
2. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```
   sudo usermod -aG sudo image-builder
   su image-builder
   ```
3. Install packages and prepare environment:
   ```
   sudo apt update -y
   sudo apt install jq unzip make ansible -y
   sudo snap install yq
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   ```
4. Get `image-builder`:
   ```bash
   cd /tmp
   sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/14/artifacts/image-builder/0.1.0/image-builder-linux-amd64.tar.gz
   sudo tar xvf image-builder*.tar.gz
   sudo cp image-builder /usr/local/bin
   ```
5. Create a content library on vSphere:
   ```
   govc library.create "<library name>"
   ```
6. Create a vsphere configuration file (for example, `vsphere-connection.json`):
   ```json
   {
     "cluster": "<vsphere cluster used for image building>",
     "convert_to_template": "false",
     "create_snapshot": "<creates a snapshot on base OVA after building if set to true>",
     "datacenter": "<vsphere datacenter used for image building>",
     "datastore": "<datastore used to store template/for image building>",
     "folder": "<folder on vsphere to create temporary vm>",
     "insecure_connection": "true",
     "linked_clone": "false",
     "network": "<vsphere network used for image building>",
     "password": "<vcenter password>",
     "resource_pool": "<resource pool used for image building vm>",
     "username": "<vcenter username>",
     "vcenter_server": "<vcenter fqdn>",
     "vsphere_library_name": "<vsphere content library name>"
   }
   ```

7. Run `image-builder` with the following options:

   * `--os`: Currently only `ubuntu` is supported.
   * `--hypervisor`: For vSphere use `vsphere`
   * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, and 1-23.
   * `--vsphere-config`: vSphere configuration file (`vsphere-connection.json` in this example)

   ```bash
   image-builder build --os ubuntu --hypervisor vsphere --release-channel 1-23 --vsphere-config vsphere-connection.json
   ```
### Build Bare Metal node images
1. Create a linux user for running image-builder.
   ```
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
2. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```
   sudo usermod -aG sudo image-builder
   su image-builder
   ```
1. Install packages and prepare environment:
   ```
   sudo apt update -y
   sudo apt install jq make qemu-kvm libvirt-daemon-system libvirt-clients virtinst cpu-checker libguestfs-tools libosinfo-bin unzip ansible -y
   sudo snap install yq
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   ```
1. Get `image-builder`:
    ```bash
    cd /tmp
    sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/14/artifacts/image-builder/0.1.0/image-builder-linux-amd64.tar.gz
    sudo tar xvf image-builder*.tar.gz
    sudo cp image-builder /usr/local/bin
    ```
1. Run `image-builder` with the following options:

    * `--os`: Currently only `ubuntu` is supported.
    * `--hypervisor`: For Bare Metal use `baremetal`
    * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, and 1-23.

   ```bash
   image-builder build --os ubuntu --hypervisor baremetal --release-channel 1-23
   ```
1. To consume the resulting gzip Ubuntu-based image, serve the image from an accessible Web server. For example, add the image to a server called `my-web-server`:
   ```
   my-web-server
   ├── hook
   │   ├── initramfs-x86_64
   │   └── vmlinuz-x86_64
   └── my-ubuntu-v1.23.9-eks-a-17-amd64.gz
   ```

1. Then create the [Bare metal configuration]({{< relref "./clusterspec/baremetal/" >}}) file, setting the `osImageURL` field to the location of the image. For example:

   ```
   osImageURL: "http://my-web-server/my-ubuntu-v1.23.9-eks-a-17-amd64.gz"
   ```

   See descriptions of [osImageURL]({{< relref "./clusterspec/baremetal/#osimageurl" >}}) for further information.
 
## Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
