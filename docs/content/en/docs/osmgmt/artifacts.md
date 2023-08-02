---
title: "Artifacts"
linkTitle: "Artifacts"
weight: 55
aliases:
    /docs/reference/artifacts/
description: >
  Artifacts associated with this release: OVAs and images.
---

EKS Anywhere supports three different node operating systems:

* Bottlerocket: For vSphere and Bare Metal providers
* Ubuntu: For vSphere, Bare Metal, Nutanix, and Snow providers
* Red Hat Enterprise Linux (RHEL): For vSphere, CloudStack, and Bare Metal providers

Bottlerocket OVAs and images are distributed by the EKS Anywhere project.
To build your own Ubuntu-based or RHEL-based EKS Anywhere node, see [Building node images]({{< relref "#building-node-images">}}).

## Bare Metal artifacts

Artifacts for EKS Anyware Bare Metal clusters are listed below.
If you like, you can download these images and serve them locally to speed up cluster creation.
See descriptions of the [osImageURL]({{< relref "../getting-started/baremetal/bare-spec/#osimageurl" >}}) and [`hookImagesURLPath`]({{< relref "../getting-started/baremetal/bare-spec#hookimagesurlpath" >}}) fields for details.

### Ubuntu or RHEL OS images for Bare Metal

EKS Anywhere does not distribute Ubuntu or RHEL OS images.
However, see [Building node images]({{< relref "#building-node-images">}}) for information on how to build EKS Anywhere images from those Linux distributions.

### Bottlerocket OS images for Bare Metal

Bottlerocket vends its Baremetal variant Images using a secure distribution tool called `tuftool`. Please refer to [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket image. You can also get the download URIs for Bottlerocket Baremetal images from the bundle release by running the following commands:

Using the latest EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
```

OR

Using a specific EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=v0.16.3
```

```bash
BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[].eksD.raw.bottlerocket.uri"
```

### HookOS (kernel and initial ramdisk) for Bare Metal
Using the latest EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
```

OR

Using a specific EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=v0.16.0
```

kernel:
```bash
BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].tinkerbell.tinkerbellStack.hook.vmlinuz.amd.uri"
```

initial ramdisk:
```bash
BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].tinkerbell.tinkerbellStack.hook.initramfs.amd.uri"
```

## vSphere artifacts

### Bottlerocket OVAs

Bottlerocket vends its VMware variant OVAs using a secure distribution tool called `tuftool`. Please refer [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket OVA. You can also get the download URIs for Bottlerocket OVAs from the bundle release by running the following commands:

Using the latest EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
```

OR

Using a specific EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=v0.16.3
```

```bash
BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[].eksD.ova.bottlerocket.uri"
```

#### Bottlerocket Template Tags

OS Family - `os:bottlerocket`

EKS Distro Release

1.27 - `eksdRelease:kubernetes-1-27-eks-4`

1.26 - `eksdRelease:kubernetes-1-26-eks-5`

1.25 - `eksdRelease:kubernetes-1-25-eks-10`

1.24 - `eksdRelease:kubernetes-1-24-eks-14`

1.23 - `eksdRelease:kubernetes-1-23-eks-19`


### Ubuntu OVAs
EKS Anywhere no longer distributes Ubuntu OVAs for use with EKS Anywhere clusters.
Building your own Ubuntu-based nodes as described in [Building node images]({{< relref "#building-node-images">}}) is the only supported way to get that functionality.

## Download Bottlerocket node images
Bottlerocket vends its VMware variant OVAs and Baremetal variants images using a secure distribution tool called `tuftool`. Please follow instructions down below to
download Bottlerocket node images.
1. Install Rust and Cargo
```bash
curl https://sh.rustup.rs -sSf | sh
```
2. Install `tuftool` using Cargo
```bash
CARGO_NET_GIT_FETCH_WITH_CLI=true cargo install --force tuftool
```
3. Download the root role that will be used by `tuftool` to download the Bottlerocket images
```bash
curl -O "https://cache.bottlerocket.aws/root.json"
sha512sum -c <<<"b81af4d8eb86743539fbc4709d33ada7b118d9f929f0c2f6c04e1d41f46241ed80423666d169079d736ab79965b4dd25a5a6db5f01578b397496d49ce11a3aa2  root.json"
```
4. Export the desired Kubernetes version. EKS Anywhere currently supports 1.23, 1.24, 1.25, 1.26, and 1.27.
```bash
export KUBEVERSION="1.27"
```
5. Programmatically retrieve the Bottlerocket version corresponding to this release of EKS-A and Kubernetes version and export it.
   
   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.0
   ```

   Set the Bottlerocket image format to the desired value (`ova` for the VMware variant or `raw` for the Baremetal variant)
   ```bash
   export BOTTLEROCKET_IMAGE_FORMAT="ova"
   ```

   ```bash
   BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   BUILD_TOOLING_COMMIT=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.gitCommit")
   export BOTTLEROCKET_VERSION=$(curl -sL https://raw.githubusercontent.com/aws/eks-anywhere-build-tooling/$BUILD_TOOLING_COMMIT/projects/kubernetes-sigs/image-builder/BOTTLEROCKET_RELEASES | yq ".$(echo $KUBEVERSION | tr '.' '-').$BOTTLEROCKET_IMAGE_FORMAT-release-version")
   ```

6. Download Bottlerocket node image

    a. To download VMware variant Bottlerocket OVA
    ```bash
    OVA="bottlerocket-vmware-k8s-${KUBEVERSION}-x86_64-${BOTTLEROCKET_VERSION}.ova"
    tuftool download ${TMPDIR:-/tmp/bottlerocket-ovas} --target-name "${OVA}" \
       --root ./root.json \
       --metadata-url "https://updates.bottlerocket.aws/2020-07-07/vmware-k8s-${KUBEVERSION}/x86_64/" \
       --targets-url "https://updates.bottlerocket.aws/targets/"
    ```
   The above command will download a Bottlerocket OVA. Please refer [Deploy an OVA Template]({{< relref "../getting-started/vsphere/vsphere-preparation/#deploy-an-ova-template">}}) to proceed with the downloaded OVA.

    b. To download Baremetal variant Bottlerocket image
    ```bash
    IMAGE="bottlerocket-metal-k8s-${KUBEVERSION}-x86_64-${BOTTLEROCKET_VERSION}.img.lz4"
    tuftool download ${TMPDIR:-/tmp/bottlerocket-metal} --target-name "${IMAGE}" \
       --root ./root.json \
       --metadata-url "https://updates.bottlerocket.aws/2020-07-07/metal-k8s-${KUBEVERSION}/x86_64/" \
       --targets-url "https://updates.bottlerocket.aws/targets/"
    ```
   The above command will download a Bottlerocket lz4 compressed image. Decompress and gzip the image with the following
   commands and host the image on a webserver for using it for an EKS Anywhere Baremetal cluster.
   ```bash
   lz4 --decompress ${TMPDIR:-/tmp/bottlerocket-metal}/${IMAGE} ${TMPDIR:-/tmp/bottlerocket-metal}/bottlerocket.img
   gzip ${TMPDIR:-/tmp/bottlerocket-metal}/bottlerocket.img
   ```

## Building node images

The `image-builder` CLI lets you build your own Ubuntu-based vSphere OVAs, Nutanix qcow2 images, RHEL-based qcow2 images, or Bare Metal gzip images to use in EKS Anywhere clusters.
When you run `image-builder`, it will pull in all components needed to build images to be used as Kubernetes nodes in an EKS Anywhere cluster, including the latest operating system, Kubernetes control plane components, and EKS Distro security updates, bug fixes, and patches.
When building an image using this tool, you get to choose:

* Operating system type (for example, ubuntu, redhat)
* Provider (vsphere, cloudstack, baremetal, ami, nutanix)
* Release channel for EKS Distro (generally aligning with Kubernetes releases)
* **vSphere only:** configuration file providing information needed to access your vSphere setup
* **CloudStack only:** configuration file providing information needed to access your CloudStack setup
* **Snow AMI only:** configuration file providing information needed to customize your Snow AMI build parameters
* **Nutanix only:** configuration file providing information needed to access Nutanix Prism Central

Because `image-builder` creates images in the same way that the EKS Anywhere project does for their own testing, images built with that tool are supported.

The table below shows the support matrix for the hypervisor and OS combinations that `image-builder` supports.

|            | vSphere | Baremetal | CloudStack | Nutanix | Snow |
|:----------:|:-------:|:---------:|:----------:|:-------:|:----:|
| **Ubuntu** |    ✓    |     ✓     |            |    ✓    |   ✓  |
|  **RHEL**  |    ✓    |     ✓     |     ✓      |         |      |


### Prerequisites

To use `image-builder`, you must meet the following prerequisites:

#### System requirements

`image-builder` has been tested on Ubuntu (20.04, 21.04, 22.04), RHEL 8 and Amazon Linux 2 machines. The following system requirements should be met for the machine on which `image-builder` is run:
* AMD 64-bit architecture
* 50 GB disk space
* 2 vCPUs
* 8 GB RAM
* **Baremetal only:** Run on a bare metal machine with virtualization enabled

#### Network connectivity requirements
* public.ecr.aws (to download container images from EKS Anywhere)
* anywhere-assets.eks.amazonaws.com (to download the EKS Anywhere artifacts such as binaries, manifests and OS images)
* distro.eks.amazonaws.com (to download EKS Distro binaries and manifests)
* d2glxqk2uabbnd.cloudfront.net (to pull the EKS Anywhere and EKS Distro ECR container images)
* api.ecr.us-west-2.amazonaws.com (for EKS Anywhere package authentication matching your region)
* d5l0dvt14r5h8.cloudfront.net (for EKS Anywhere package ECR container)
* github.com (to download binaries and tools required for image builds from GitHub releases)
* objects.githubusercontent.com (to download binaries and tools required for image builds from GitHub releases)
* raw.githubusercontent.com (to download binaries and tools required for image builds from GitHub releases)
* releases.hashicorp.com (to download Packer binary for image builds)
* galaxy.ansible.com (to download Ansible packages from Ansible Galaxy)
* **vSphere only:** VMware vCenter endpoint
* **CloudStack only:** Apache CloudStack endpoint
* **Nutanix only:** Nutanix Prism Central endpoint
* **Red Hat only:** dl.fedoraproject.org (to download RPMs and GPG keys for RHEL image builds)
* **Ubuntu only:** cdimage.ubuntu.com (to download Ubuntu server ISOs for Ubuntu image builds)

#### vSphere requirements

image-builder uses the Hashicorp [vsphere-iso](https://developer.hashicorp.com/packer/plugins/builders/vsphere/vsphere-iso#packer-builder-for-vmware-vsphere) Packer Builder for building vSphere OVAs.

##### Permissions

Configure a user with a role containing the following permissions.

The role can be configured programmatically with the `govc` command below, or configured in the vSphere UI using the table below as reference.

Note that no matter how the role is created, it must be assigned to the user or user group at the **Global Permissions** level.

Unfortunately there is no API for managing vSphere Global Permissions, so they must be set on the user via the UI under `Administration > Access Control > Global Permissions`.

To generate a role named EKSAImageBuilder with the required privileges via `govc`, run the following:
```bash
govc role.create "EKSAImageBuilder" $(curl https://raw.githubusercontent.com/aws/eks-anywhere/main/pkg/config/static/imageBuilderPrivs.json | jq .[] | tr '\n' ' ' | tr -d '"')
```

If creating a role with these privileges via the UI, refer to the table below.

| Category | UI Privilege | Programmatic Privilege |
| --- | ----------- | ---- |
| Content Library | Add library item | ContentLibrary.AddLibraryItem |
| Content Library | Delete library item | ContentLibrary.DeleteLibraryItem |
| Content Library | Download files | ContentLibrary.DownloadSession |
| Content Library | Evict library item | ContentLibrary.EvictLibraryItem |
| Content Library | Update library item | ContentLibrary.UpdateLibraryItem |
| Datastore | Allocate space | Datastore.AllocateSpace |
| Datastore | Browse datastore | Datastore.Browse |
| Datastore | Low level file operations | Datastore.FileManagement |
| Network | Assign network | Network.Assign |
| Resource | Assign virtual machine to resource pool | Resource.AssignVMToPool |
| vApp | Export | vApp.Export |
| VirtualMachine | Configuration > Add new disk | VirtualMachine.Config.AddNewDisk |
| VirtualMachine | Configuration > Add or remove device | VirtualMachine.Config.AddRemoveDevice |
| VirtualMachine | Configuration > Advanced configuration | VirtualMachine.Config.AdvancedConfiguration |
| VirtualMachine | Configuration > Change CPU count | VirtualMachine.Config.CPUCount |
| VirtualMachine | Configuration > Change memory | VirtualMachine.Config.Memory |
| VirtualMachine | Configuration > Change settings | VirtualMachine.Config.Settings |
| VirtualMachine | Configuration > Change Resource | VirtualMachine.Config.Resource |
| VirtualMachine | Configuration > Set annotation | VirtualMachine.Config.Annotation |
| VirtualMachine | Edit Inventory > Create from existing | VirtualMachine.Inventory.CreateFromExisting |
| VirtualMachine | Edit Inventory > Create new | VirtualMachine.Inventory.Create |
| VirtualMachine | Edit Inventory > Remove | VirtualMachine.Inventory.Delete |
| VirtualMachine | Interaction > Configure CD media | VirtualMachine.Interact.SetCDMedia |
| VirtualMachine | Interaction > Configure floppy media | VirtualMachine.Interact.SetFloppyMedia |
| VirtualMachine | Interaction > Connect devices | VirtualMachine.Interact.DeviceConnection |
| VirtualMachine | Interaction > Inject USB HID scan codes | VirtualMachine.Interact.PutUsbScanCodes |
| VirtualMachine | Interaction > Power off | VirtualMachine.Interact.PowerOff |
| VirtualMachine | Interaction > Power on | VirtualMachine.Interact.PowerOn |
| VirtualMachine | Interaction > Create template from virtual machine | VirtualMachine.Provisioning.CreateTemplateFromVM |
| VirtualMachine | Interaction > Mark as template | VirtualMachine.Provisioning.MarkAsTemplate |
| VirtualMachine | Interaction > Mark as virtual machine | VirtualMachine.Provisioning.MarkAsVM |
| VirtualMachine | State > Create snapshot | VirtualMachine.State.CreateSnapshot |


#### CloudStack requirements
Refer to the [CloudStack Permissions for CAPC](https://github.com/kubernetes-sigs/cluster-api-provider-cloudstack/blob/main/docs/book/src/topics/cloudstack-permissions.md) doc for required CloudStack user permissions.

#### Snow AMI requirements
Packer will require prior authentication with your AWS account to launch EC2 instances for the Snow AMI build. Refer to the [Authentication guide for Amazon EBS Packer builder](https://developer.hashicorp.com/packer/plugins/builders/amazon#authentication) for possible modes of authentication. We recommend that you run `image-builder` on a pre-existing Ubuntu EC2 instance and use an [IAM instance role with the required permissions](https://developer.hashicorp.com/packer/plugins/builders/amazon#iam-task-or-instance-role).

#### Nutanix permissions

Prism Central Administrator permissions are required to build a Nutanix image using `image-builder`.

### Optional Proxy configuration
You can use a proxy server to route outbound requests to the internet. To configure `image-builder` tool to use a proxy server, export these proxy environment variables:
  ```bash
  export HTTP_PROXY=<HTTP proxy URL e.g. http://proxy.corp.com:80>
  export HTTPS_PROXY=<HTTPS proxy URL e.g. http://proxy.corp.com:443>
  export NO_PROXY=<No proxy>
  ```

### Build vSphere OVA node images

These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for vSphere. Before proceeding, ensure that the above system-level, network-level and vSphere-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a linux user for running image-builder.
   ```bash
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```bash
   sudo usermod -aG sudo image-builder
   su image-builder
   cd /home/$USER
   ```
1. Install packages and prepare environment:
   {{< tabpane >}}

   {{< tab header="Ubuntu" lang="bash" >}}
   sudo apt update -y
   sudo apt install jq unzip make ansible python3-pip -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make python3-pip wget -y
   python3 -m pip install --user ansible
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make python3-pip ansible wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}
1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.3
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo cp ./image-builder /usr/local/bin
   cd -
   ```
1. Get the latest version of `govc`:
   ```bash
   curl -L -o - "https://github.com/vmware/govmomi/releases/latest/download/govc_$(uname -s)_$(uname -m).tar.gz" | sudo tar -C /usr/local/bin -xvzf - govc
   ```
1. Create a content library on vSphere:
   ```bash
   govc library.create "<library name>"
   ```
1. Create a vsphere configuration file (for example, `vsphere-connection.json`):
   ```json
   {
     "cluster": "<vsphere cluster used for image building>",
     "convert_to_template": "false",
     "create_snapshot": "<creates a snapshot on base OVA after building if set to true>",
     "datacenter": "<vsphere datacenter used for image building>",
     "datastore": "<datastore used to store template/for image building>",
     "folder": "<folder on vsphere to create temporary VM>",
     "insecure_connection": "true",
     "linked_clone": "false",
     "network": "<vsphere network used for image building>",
     "password": "<vcenter password>",
     "resource_pool": "<resource pool used for image building VM>",
     "username": "<vcenter username>",
     "vcenter_server": "<vcenter fqdn>",
     "vsphere_library_name": "<vsphere content library name>"
   }
   ```
   For RHEL images, add the following fields:
   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<rhsm username>",
     "rhel_password": "<rhsm password>"
   }
   ```

1. Create an Ubuntu or Redhat image:

   * To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23, 1-24 and 1-25.
      * `--vsphere-config`: vSphere configuration file (`vsphere-connection.json` in this example)

      ```bash
      image-builder build --os ubuntu --hypervisor vsphere --release-channel 1-25 --vsphere-config vsphere-connection.json
      ```
   * To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23, 1-24 and 1-25.
      * `--vsphere-config`: vSphere configuration file (`vsphere-connection.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor vsphere --release-channel 1-25 --vsphere-config vsphere-connection.json
      ```
### Build Bare Metal node images
These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for Bare Metal. Before proceeding, ensure that the above system-level, network-level and baremetal-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a linux user for running image-builder.
   ```bash
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
2. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```bash
   sudo usermod -aG sudo image-builder
   su image-builder
   cd /home/$USER
   ```
1. Install packages and prepare environment:
   {{< tabpane >}}

   {{< tab header="Ubuntu" lang="bash" >}}
   sudo apt update -y
   sudo apt install jq make python3-pip qemu-kvm libvirt-daemon-system libvirt-clients virtinst cpu-checker libguestfs-tools libosinfo-bin unzip ansible -y
   sudo snap install yq
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq make python3-pip qemu-kvm libvirt virtinst cpu-checker libguestfs-tools libosinfo unzip wget -y
   python3 -m pip install --user ansible
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq make python3-pip qemu-kvm libvirt libvirt-clients libguestfs-tools unzip ansible wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}
1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo cp ./image-builder /usr/local/bin
   cd -
   ```

1. Create an Ubuntu or Red Hat image.

   **Ubuntu**

   Run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--hypervisor`: `baremetal`
      * `--release-channel`: A [supported EKS Distro release](https://anywhere.eks.amazonaws.com/docs/reference/support/support-versions/) 
      formatted as "[major]-[minor]"; for example "1-25"

      ```bash
      image-builder build --os ubuntu --hypervisor baremetal --release-channel 1-25
      ```

   **Red Hat Enterprise Linux (RHEL)**

   RHEL images require a configuration file to identify the location of the RHEL 8 ISO image and
   Red Hat subscription information. The `image-builder` command will temporarily consume a Red
   Hat subscription that is returned once the image is built.

   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<rhsm username>",
     "rhel_password": "<rhsm password>",
     "extra_rpms": "<space-separated list of RPM packages; useful for adding required drivers or other packages>"
   }
   ```

   Run the `image-builder` with the following options:

      * `--os`: `redhat`
      * `--hypervisor`: `baremetal`
      * `--release-channel`: A [supported EKS Distro release](https://anywhere.eks.amazonaws.com/docs/reference/support/support-versions/) 
      formatted as "[major]-[minor]"; for example "1-25"
      * `--baremetal-config`: Bare metal config file

      ```bash
      image-builder build --os redhat --hypervisor baremetal --release-channel 1-25 --baremetal-config baremetal.json
      ```

1. To consume the image, serve it from an accessible web server, then create the [bare metal cluster spec]({{< relref "../getting-started/baremetal/bare-spec/" >}}) 
   configuring the `osImageURL` field URL of the image. For example:

   ```
   osImageURL: "http://<artifact host address>/my-ubuntu-v1.23.9-eks-a-17-amd64.gz"
   ```

   See descriptions of [osImageURL]({{< relref "../getting-started/baremetal/bare-spec/#osimageurl" >}}) for further information.

### Build CloudStack node images

These steps use `image-builder` to create a RHEL-based image for CloudStack. Before proceeding, ensure that the above system-level, network-level and CloudStack-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a linux user for running image-builder.
   ```bash
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```bash
   sudo usermod -aG sudo image-builder
   su image-builder
   cd /home/$USER
   ```
1. Install packages and prepare environment:
   {{< tabpane >}}

   {{< tab header="Ubuntu" lang="bash" >}}
   sudo apt update -y
   sudo apt install jq make python3-pip qemu-kvm libvirt-daemon-system libvirt-clients virtinst cpu-checker libguestfs-tools libosinfo-bin unzip ansible -y
   sudo snap install yq
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq make python3-pip qemu-kvm libvirt virtinst cpu-checker libguestfs-tools libosinfo unzip wget -y
   python3 -m pip install --user ansible
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq make python3-pip qemu-kvm libvirt libvirt-clients libguestfs-tools unzip ansible wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}
1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.3
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo cp ./image-builder /usr/local/bin
   cd -
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
      * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23, 1-24 and 1-25.
      * `--cloudstack-config`: CloudStack configuration file (`cloudstack.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor cloudstack --release-channel 1-25 --cloudstack-config cloudstack.json
      ```

1. To consume the resulting RHEL-based image, add it as a template to your CloudStack setup as described in [Preparing CloudStack]({{< relref "../getting-started/cloudstack/cloudstack-preparation" >}}).

### Build Snow node images

These steps use `image-builder` to create an Ubuntu-based Amazon Machine Image (AMI) that is backed by EBS volumes for Snow. Before proceeding, ensure that the above system-level, network-level and AMI-specific [prerequisites]({{< relref "#prerequisites">}}) have been met

1. Create a linux user for running image-builder.
   ```bash
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add the `image-builder` user to the `sudo` group and switch user to `image-builder`, providing in the password from previous step when prompted.
   ```bash
   sudo usermod -aG sudo image-builder
   su image-builder
   cd /home/$USER
   ```
1. Install packages and prepare environment:
   {{< tabpane >}}

   {{< tab header="Ubuntu" lang="bash" >}}
   sudo apt update -y
   sudo apt install jq unzip make ansible python3-pip -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make python3-pip wget -y
   python3 -m pip install --user ansible
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make python3-pip ansible wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}
1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo cp ./image-builder /usr/local/bin
   cd /home/$USER
   ```
1. Create an AMI configuration file (for example, `ami.json`) that contains various AMI parameters. For example:

   ```json
   {
      "ami_filter_name": "ubuntu/images/*ubuntu-focal-20.04-amd64-server-*",
      "ami_filter_owners": "679593333241",
      "ami_regions": "us-east-2",
      "aws_region": "us-east-2",
      "ansible_extra_vars": "@/home/image-builder/eks-anywhere-build-tooling/projects/kubernetes-sigs/image-builder/packer/ami/ansible_extra_vars.yaml",
      "builder_instance_type": "t3.small",
      "custom_role_name_list" : ["/home/image-builder/eks-anywhere-build-tooling/projects/kubernetes-sigs/image-builder/ansible/roles/load_additional_files"],
      "manifest_output": "/home/image-builder/manifest.json",
      "root_device_name": "/dev/sda1",
      "volume_size": "25",
      "volume_type": "gp3",
   }
   ```

   ##### **ami_filter_name**
   Regular expression to filter a source AMI. (default: `ubuntu/images/*ubuntu-focal-20.04-amd64-server-*`).

   ##### **ami_filter_owners**
   AWS account ID or AWS owner alias such as 'amazon', 'aws-marketplace', etc. (default: `679593333241` - the AWS Marketplace AWS account ID).

   ##### **ami_regions**
   A list of AWS regions to copy the AMI to. (default: `us-west-2`).

   ##### **aws_region**
   The AWS region in which to launch the EC2 instance to create the AMI. (default: `us-west-2`).

   ##### **ansible_extra_vars**
   The absolute path to the additional variables to pass to Ansible. These are converted to the `--extra-vars` command-line argument. This path must be prefix with '@'. (default: `@/home/image-builder/eks-anywhere-build-tooling/projects/kubernetes-sigs/image-builder/packer/ami/ansible_extra_vars.yaml`)

   ##### **builder_instance_type**
   The EC2 instance type to use while building the AMI. (default: `t3.small`).

   ##### **custom_role_name_list**
   Array of strings representing the absolute paths of custom Ansible roles to run. This field is mutually exclusive with `custom_role_names`.

   ##### **custom_role_names**
   Space-delimited string of the custom roles to run. This field is mutually exclusive with `custom_role_name_list` and is provided for compatibility with Ansible's input format.

   ##### **manifest_output**
   The absolute path to write the build artifacts manifest to. If you wish to export the AMI using this manifest, ensure that you provide a path that is not inside the `/home/$USER/eks-anywhere-build-tooling` path since that will be cleaned up when the build finishes. (default: `/home/image-builder/manifest.json`).

   ##### **root_device_name**
   The device name used by EC2 for the root EBS volume attached to the instance. (default: `/dev/sda1`).

   ##### **subnet_id**
   The ID of the subnet where Packer will launch the EC2 instance. This field is required when using a non-default VPC.

   ##### **volume_size**
   The size of the root EBS volume in GiB. (default: `25`).

   ##### **volume_type**
   The type of root EBS volume, such as gp2, gp3, io1, etc. (default: `gp3`).

1. To create an Ubuntu-based image, run `image-builder` with the following options:

   * `--os`: `ubuntu`
   * `--hypervisor`: For AMI, use `ami`
   * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23 and 1-24.
   * `--ami-config`: AMI configuration file (`ami.json` in this example)

   ```bash
   image-builder build --os ubuntu --hypervisor ami --release-channel 1-24 --ami-config ami.json
   ```
1. After the build, the Ubuntu AMI will be available in your AWS account in the AWS region specified in your AMI configuration file. If you wish to export it as a Raw image, you can achieve this using the AWS CLI.
   ```
   ARTIFACT_ID=$(cat <manifest output location> | jq -r '.builds[0].artifact_id')
   AMI_ID=$(echo $ARTIFACT_ID | cut -d: -f2)
   IMAGE_FORMAT=raw
   AMI_EXPORT_BUCKET_NAME=<S3 bucket to export the AMI to>
   AMI_EXPORT_PREFIX=<S3 prefix for the exported AMI object>
   EXPORT_RESPONSE=$(aws ec2 export-image --disk-image-format $IMAGE_FORMAT --s3-export-location S3Bucket=$AMI_EXPORT_BUCKET_NAME,S3Prefix=$AMI_EXPORT_PREFIX --image-id $AMI_ID)
   EXPORT_TASK_ID=$(echo $EXPORT_RESPONSE | jq -r '.ExportImageTaskId')
   ```
   The exported image will be available at the location `s3://$AMI_EXPORT_BUCKET_NAME/$AMI_EXPORT_PREFIX/$EXPORT_IMAGE_TASK_ID.raw`.

### Build Nutanix node images

These steps use `image-builder` to create a Ubuntu-based image for Nutanix AHV and import it into the AOS Image Service. Before proceeding, ensure that the above system-level, network-level and Nutanix-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Download an [Ubuntu cloud image](https://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img) for the build and upload it to the AOS Image Service using Prism. You will need to specify this image name as the `source_image_name` in the `nutanix-connection.json` config file specified below.

1. Create a linux user for running image-builder.
   ```bash
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add image-builder user to the sudo group and change user as image-builder providing in the password from previous step when prompted.
   ```bash
   sudo usermod -aG sudo image-builder
   su image-builder
   cd /home/$USER
   ```
1. Install packages and prepare environment:
   {{< tabpane >}}

   {{< tab header="Ubuntu" lang="bash" >}}
   sudo apt update -y
   sudo apt install jq unzip make ansible python3-pip -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make python3-pip wget -y
   python3 -m pip install --user ansible
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make python3-pip ansible wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}
1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.16.3
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo cp ./image-builder /usr/local/bin
   cd -
   ```
1. Create a `nutanix-connection.json` config file. More details on values can be found in the [image-builder documentation](https://image-builder.sigs.k8s.io/capi/providers/nutanix.html). See example below:
   ```json
   {
     "nutanix_cluster_name": "Name of PE Cluster",
     "source_image_name": "Name of Source Image",
     "image_name": "Name of Destination Image",
     "nutanix_subnet_name": "Name of Subnet",
     "nutanix_endpoint": "Prism Central IP / FQDN",
     "nutanix_insecure": "false",
     "nutanix_port": "9440",
     "nutanix_username": "PrismCentral_Username",
     "nutanix_password": "PrismCentral_Password"
   }
   ```

1. Run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--hypervisor`: For Nutanix use `nutanix`
      * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23, 1-24 and 1-25.
      * `--nutanix-config`: Nutanix configuration file (`nutanix-connection.json` in this example)

      ```bash
      cd /home/$USER
      image-builder build --os ubuntu --hypervisor nutanix --release-channel 1-25 --nutanix-config nutanix-connection.json
      ```

## Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
