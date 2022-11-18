---
title: "Artifacts"
linkTitle: "Artifacts"
weight: 55
description: >
  Artifacts associated with this release: OVAs and images.
---

EKS Anywhere supports three different node operating systems:

* Bottlerocket: For vSphere and Bare Metal providers
* Ubuntu: For vSphere and Bare Metal providers
* Red Hat Enterprise Linux (RHEL): For vSphere, CloudStack, and Bare Metal providers

Bottlerocket OVAs and images are distributed by the EKS Anywhere project.
To build your own Ubuntu-based or RHEL-based EKS Anywhere node, see [Building node images]({{< relref "#building-node-images">}}).

## Bare Metal artifacts

Artifacts for EKS Anyware Bare Metal clusters are listed below.
If you like, you can download these images and serve them locally to speed up cluster creation.
See descriptions of the [osImageURL]({{< relref "./clusterspec/baremetal/#osimageurl" >}}) and [`hookImagesURLPath`]({{< relref "./clusterspec/baremetal/#hookimagesurlpath" >}}) fields for details.

### Ubuntu or RHEL OS images for Bare Metal

EKS Anywhere does not distribute Ubuntu or RHEL OS images.
However, see [Building node images]({{< relref "#building-node-images">}}) for information on how to build EKS Anywhere images from those Linux distributions.

### Bottlerocket OS images for Bare Metal

Bottlerocket vends its Baremetal variant Images using a secure distribution tool called tuftool. Please refer to [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket image.

### HookOS (kernel and initial ramdisk) for Bare Metal

kernel:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/21/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/vmlinuz-x86_64
```

initial ramdisk:
```bash
https://anywhere-assets.eks.amazonaws.com/releases/bundles/21/artifacts/hook/029ef8f0711579717bfd14ac5eb63cdc3e658b1d/initramfs-x86_64
```

## vSphere artifacts

### Bottlerocket OVAs

Bottlerocket vends its VMware variant OVAs using a secure distribution tool called tuftool. Please refer [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket OVA.

Bottlerocket Tags

OS Family - `os:bottlerocket`

EKS-D Release

1.24 - `eksdRelease:kubernetes-1-24-eks-2`

1.23 - `eksdRelease:kubernetes-1-23-eks-8`

1.22 - `eksdRelease:kubernetes-1-22-eks-12`

1.21 - `eksdRelease:kubernetes-1-21-eks-21`

1.20 - `eksdRelease:kubernetes-1-20-eks-22`

### Ubuntu OVAs
EKS Anywhere no longer distributes Ubuntu OVAs for use with EKS Anywhere clusters.
Building your own Ubuntu-based nodes as described in [Building node images]({{< relref "#building-node-images">}}) is the only supported way to get that functionality.

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
3. Download the root role that will be used by tuftool to download the Bottlerocket images
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
    OVA="bottlerocket-vmware-k8s-${KUBEVERSION}-x86_64-v1.10.1.ova"
    tuftool download ${TMPDIR:-/tmp/bottlerocket-ovas} --target-name "${OVA}" \
       --root ./root.json \
       --metadata-url "https://updates.bottlerocket.aws/2020-07-07/vmware-k8s-${KUBEVERSION}/x86_64/" \
       --targets-url "https://updates.bottlerocket.aws/targets/"
    ```
   The above command will download Bottlerocket OVA. Please refer [Deploy an OVA Template]({{< relref "vsphere/vsphere-preparation/#deploy-an-ova-template">}}) to proceed with the downloaded OVA.

    b. To download Baremetal variant Bottlerocket image
    ```
    IMAGE="bottlerocket-metal-k8s-${KUBEVERSION}-x86_64-v1.10.1.img.lz4"
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

## Building node images

The `image-builder` CLI lets you build your own Ubuntu-based vSphere OVAs, RHEL-based qcow2 images, or Bare Metal gzip images to use in EKS Anywhere clusters.
When you run `image-builder` it will pull in all components needed to create images to use for nodes in an EKS Anywhere cluster, including the lastest operating system, Kubernetes, and EKS Distro security updates, bug fixes, and patches.
With this tool, when you build an image you get to choose:

* Operating system type (for example, ubuntu)
* Provider (vsphere, cloudstack or baremetal)
* Release channel for EKS Distro (generally aligning with Kubernetes releases)
* vSphere only: configuration file providing information needed to access your vSphere setup
* CloudStack only: configuration file providing information needed to access your Cloudstack setup

Because `image-builder` creates images in the same way that the EKS Anywhere project does for their own testing, images built with that tool are supported.
The following procedure describes how to use `image-builder` to build images for EKS Anywhere on a vSphere or Bare Metal provider.

### Prerequisites

To use `image-builder` you must meet the following prerequisites:

* Run on Ubuntu 22.04 or later (for Ubuntu images) or RHEL 8.4 or later (for RHEL images)
* Machine requirements:
  * AMD 64-bit architecture
  * 50 GB disk space
  * 2 vCPUs
  * 8 GB RAM
  * Bare Metal only: Run on a bare metal machine with virtualization enabled
* Network access to:
  * vCenter endpoint (vSphere only)
  * CloudStack endpoint (CloudStack only)
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
* CloudStack only: See [CloudStack Permissions for CAPC](https://github.com/kubernetes-sigs/cluster-api-provider-cloudstack/blob/main/docs/book/src/topics/cloudstack-permissions.md) for required CloudStack user permissions.

### Optional Proxy configuration
You can use a proxy server to route outbound requests to the internet. To configure `image-builder` tool to use a proxy server, export these proxy environment variables:
  ```
  export HTTP_PROXY=<HTTP proxy URL e.g. http://proxy.corp.com:80>
  export HTTPS_PROXY=<HTTPS proxy URL e.g. http://proxy.corp.com:443>
  export NO_PROXY=<No proxy>
  ```

### Build vSphere OVA node images

These steps use `image-builder` to create a Ubuntu-based or RHEL-based image for vSphere.

1. Create a linux user for running image-builder.
   ```
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```
   sudo usermod -aG sudo image-builder
   su image-builder
   ```
1. Install packages and prepare environment:
   ```
   sudo apt update -y
   sudo apt install jq unzip make ansible -y
   sudo snap install yq
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   ```
1. Get `image-builder`:
   ```bash
   cd /tmp
   sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/21/artifacts/image-builder/0.1.2/image-builder-linux-amd64.tar.gz
   sudo tar xvf image-builder*.tar.gz
   sudo cp image-builder /usr/local/bin
   ```
1. Create a content library on vSphere:
   ```
   govc library.create "<library name>"
   ```
1. Create a vsphere configuration file (for example, `vsphere-connection.json`):
   ```json
   {
     "cluster":"<vsphere cluster used for image building>",
     "convert_to_template":"false",
     "create_snapshot":"<creates a snapshot on base OVA after building if set to true>",
     "datacenter":"<vsphere datacenter used for image building>",
     "datastore":"<datastore used to store template/for image building>",
     "folder":"<folder on vsphere to create temporary vm>",
     "insecure_connection":"true",
     "linked_clone":"false",
     "network":"<vsphere network used for image building>",
     "password":"<vcenter username>",
     "resource_pool":"<resource pool used for image building vm>",
     "username":"<vcenter username>",
     "vcenter_server":"<vcenter fqdn>",
     "vsphere_library_name": "<vsphere content library name>"
   }
   ```
   For RHEL images, add the following fields:
   ```json
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<rhsm username>",
     "rhel_password": "<rhsm password>"
   ```

1. Create an ubuntu or redhat image:

   * To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, 1-23, and 1-24.
      * `--vsphere-config`: vSphere configuration file (`vsphere-connection.json` in this example)

      ```bash
      image-builder build --os ubuntu --hypervisor vsphere --release-channel 1-24 --vsphere-config vsphere-connection.json
      ```
   * To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, 1-23, and 1-24.
      * `--vsphere-config`: vSphere configuration file (`vsphere-connection.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor vsphere --release-channel 1-24 --vsphere-config vsphere-connection.json
      ```
### Build Bare Metal node images
These steps use `image-builder` to create a Ubuntu-based or RHEL-based image for Bare Metal.

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
    sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/21/artifacts/image-builder/0.1.2/image-builder-linux-amd64.tar.gz
    sudo tar xvf image-builder*.tar.gz
    sudo cp image-builder /usr/local/bin
    ```
1. Create a Bare Metal configuration file (for example, `baremetal.json`) to identify the location of a Red Hat Enterprise Linux 8 ISO image and related checksum and Red Hat subscription information:
   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<rhsm username>",
     "rhel_password": "<rhsm password>"
     "extra_rpms": "<Space-separated list of RPM packages. Useful for adding required drivers or other packages>"
   }
   ```
   >**_NOTE_**: To build the RHEL-based image, `image-builder` temporarily consumes a Red Hat subscription. That subscription is returned once the image is built.

1. Create an ubuntu or redhat image:

   * To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--hypervisor`: For Bare Metal use `baremetal`
      * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, 1-23, and 1-24.
      * `--baremetal-config`: Bare Metal configuration file (`baremetal.json` in this example)

      ```bash
      image-builder build --os ubuntu --hypervisor baremetal --release-channel 1-24 --baremetal-config baremetal.json
      ```
   * To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--hypervisor`: For Bare Metal use `baremetal`
      * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, 1-23, and 1-24.
      * `--baremetal-config`: Bare Metal configuration file (`baremetal.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor baremetal --release-channel 1-24 --baremetal-config baremetal.json
      ```

1. To consume the resulting Ubuntu-based or RHEL-based image, serve the image from an accessible Web server. For example, add the image to a server called `my-web-server`:
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

### Build CloudStack node images

These steps use `image-builder` to create a RHEL-based image for CloudStack.

1. Create a linux user for running image-builder.
   ```
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
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
    sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/19/artifacts/image-builder/0.1.2/image-builder-linux-amd64.tar.gz
    sudo tar xvf image-builder*.tar.gz
    sudo cp image-builder /usr/local/bin
    ```
1. Create a CloudStack configuration file (for example, `cloudstack.json`) to identify the location of a Red Hat Enterprise Linux 8 ISO image and related checksum and Red Hat subscription information:
   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<rhsm username>",
     "rhel_password": "<rhsm password>"
   }
   ```
   >**_NOTE_**: To build the RHEL-based image, `image-builder` temporarily consumes a Red Hat subscription. That subscription is returned once the image is built.

1. To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--hypervisor`: For CloudStack use `cloudstack`
      * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, and 1-23.
      * `--cloudstack-config`: CloudStack configuration file (`cloudstack.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor cloudstack --release-channel 1-23 --cloudstack-config cloudstack.json
      ```

1. To consume the resulting RHEL-based image, add it as a template to your CloudStack setup as described in [Preparing Cloudstack]({{< relref "./cloudstack/cloudstack-preparation.md" >}}).

## Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
