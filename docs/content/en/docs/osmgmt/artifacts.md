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

## Prerequisites

Several code snippets on this page use `curl` and `yq` commands. Refer to the [Tools section]({{< relref "../getting-started/install/#tools" >}}) to learn how to install them.

## Bare Metal artifacts

Artifacts for EKS Anywhere Bare Metal clusters are listed below.
If you like, you can download these images and serve them locally to speed up cluster creation.
See descriptions of the [osImageURL]({{< relref "../getting-started/baremetal/bare-spec/#osimageurl" >}}) and [`hookImagesURLPath`]({{< relref "../getting-started/baremetal/bare-spec#hookimagesurlpath" >}}) fields for details.

### Ubuntu or RHEL OS images for Bare Metal

EKS Anywhere does not distribute Ubuntu or RHEL OS images.
However, see [Building node images]({{< relref "#building-node-images">}}) for information on how to build EKS Anywhere images from those Linux distributions.  Note:  if you utilize your Admin Host to build images, you will need to review  the DHCP integration provided by Libvirtd and ensure it is disabled.  If the Libvirtd DHCP is enabled, the "boots container" will detect a port conflict and terminate.

### Bottlerocket OS images for Bare Metal

Bottlerocket vends its Baremetal variant Images using a secure distribution tool called `tuftool`. Please refer to [Download Bottlerocket node images]({{< relref "#download-bottlerocket-node-images">}}) to download Bottlerocket image. You can also get the download URIs for Bottlerocket Baremetal images from the bundle release by running the following commands:

Using the latest EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
```

OR

Using a specific EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=v0.18.0
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
EKSA_RELEASE_VERSION=v0.18.0
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
EKSA_RELEASE_VERSION=v0.18.0
```

```bash
BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[].eksD.ova.bottlerocket.uri"
```

#### Bottlerocket Template Tags

There are two categories of tags to be attached to the Bottlerocket templates in vCenter.

**os:** This category represents the OS corresponding to this template. The value for this tag must be `os:bottlerocket`.

**eksdRelease:** This category represents the EKS Distro release corresponding to this template. The value for this tag can be obtained programmatically as follows.

Using the latest EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
```

OR

Using a specific EKS Anywhere version
```bash
EKSA_RELEASE_VERSION=v0.18.0
```

```bash
KUBEVERSION=1.27 # Replace this with the Kubernetes version you wish to use

BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
curl -sL $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[] | select(.kubeVersion==\"$KUBEVERSION\").eksD.name"
```

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
sha512sum -c <<<"a3c58bc73999264f6f28f3ed9bfcb325a5be943a782852c7d53e803881968e0a4698bd54c2f125493f4669610a9da83a1787eb58a8303b2ee488fa2a3f7d802f  root.json"
```
4. Export the desired Kubernetes version. EKS Anywhere currently supports 1.23, 1.24, 1.25, 1.26, 1.27, and 1.28.
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
   EKSA_RELEASE_VERSION=v0.18.0
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

* Operating system type (for example, ubuntu, redhat) and version (Ubuntu only)
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
|  **RHEL**  |    ✓    |     ✓     |     ✓      |    ✓    |      |

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

### Build vSphere OVA node images

These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for vSphere. Before proceeding, ensure that the above system-level, network-level and vSphere-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a Linux user for running image-builder.
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
   sudo apt install jq unzip make -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}

   * Starting with `image-builder` version `v0.3.0`, the minimum required Python version is Python 3.9. However, many Linux distros ship only up to Python 3.8, so you will need to install Python 3.9 from external sources. Refer to the `pyenv` [installation](https://github.com/pyenv/pyenv#installation) and [usage](https://github.com/pyenv/pyenv#usage) documentation to install Python 3.9 and make it the default Python version.
   * Once you have Python 3.9, you can install Ansible using `pip`.
     ```bash
     python3 -m pip install --user ansible
     ```

1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo install -m 0755 ./image-builder /usr/local/bin/image-builder
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
1. Create a vSphere configuration file (for example, `vsphere.json`):
   ```json
   {
     "cluster": "<vsphere cluster used for image building>",
     "convert_to_template": "false",
     "create_snapshot": "<creates a snapshot on base OVA after building if set to true>",
     "datacenter": "<vsphere datacenter used for image building>",
     "datastore": "<datastore used to store template/for image building>",
     "folder": "<folder on vSphere to create temporary VM>",
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
     "rhel_username": "<RHSM username>",
     "rhel_password": "<RHSM password>"
   }
   ```
1. Create an Ubuntu or Redhat image:

   **Ubuntu**

   To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--os-version`: `20.04` or `22.04` (default: `20.04`)
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
      * `--vsphere-config`: vSphere configuration file (`vsphere.json` in this example)

      ```bash
      image-builder build --os ubuntu --hypervisor vsphere --release-channel 1-28 --vsphere-config vsphere.json
      ```

   **Red Hat Enterprise Linux**

   To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--os-version`: `8` (default: `8`)
      * `--hypervisor`: For vSphere use `vsphere`
      * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
      * `--vsphere-config`: vSphere configuration file (`vsphere.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor vsphere --release-channel 1-28 --vsphere-config vsphere.json
      ```

### Build Bare Metal node images
These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for Bare Metal. Before proceeding, ensure that the above system-level, network-level and baremetal-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a Linux user for running image-builder.
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
   sudo apt install jq make qemu-kvm libvirt-daemon-system libvirt-clients virtinst cpu-checker libguestfs-tools libosinfo-bin unzip -y
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
   sudo dnf install jq make qemu-kvm libvirt virtinst cpu-checker libguestfs-tools libosinfo unzip wget -y
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
   sudo yum install jq make qemu-kvm libvirt libvirt-clients libguestfs-tools unzip wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}

   * Starting with `image-builder` version `v0.3.0`, the minimum required Python version is Python 3.9. However, many Linux distros ship only up to Python 3.8, so you will need to install Python 3.9 from external sources. Refer to the `pyenv` [installation](https://github.com/pyenv/pyenv#installation) and [usage](https://github.com/pyenv/pyenv#usage) documentation to install Python 3.9 and make it the default Python version.
   * Once you have Python 3.9, you can install Ansible using `pip`.
     ```bash
     python3 -m pip install --user ansible
     ```

1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo install -m 0755 ./image-builder /usr/local/bin/image-builder   
   cd -
   ```

1. Create an Ubuntu or Red Hat image:

   **Ubuntu**
   
   To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--os-version`: `20.04` or `22.04` (default: `20.04`)
      * `--hypervisor`: `baremetal`
      * `--release-channel`: A [supported EKS Distro release](https://anywhere.eks.amazonaws.com/docs/reference/support/support-versions/)
      formatted as "[major]-[minor]"; for example "1-27"
      * `--baremetal-config`: baremetal config file if using proxy

      ```bash
      image-builder build --os ubuntu --hypervisor baremetal --release-channel 1-27
      ```

   **Red Hat Enterprise Linux (RHEL)**

   RHEL images require a configuration file to identify the location of the RHEL 8 ISO image and
   Red Hat subscription information. The `image-builder` command will temporarily consume a Red
   Hat subscription that is removed once the image is built.

   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<RHSM username>",
     "rhel_password": "<RHSM password>",
     "extra_rpms": "<space-separated list of RPM packages; useful for adding required drivers or other packages>"
   }
   ```

   Run the `image-builder` with the following options:

      * `--os`: `redhat`
      * `--os-version`: `8` (default: `8`)
      * `--hypervisor`: `baremetal`
      * `--release-channel`: A [supported EKS Distro release](https://anywhere.eks.amazonaws.com/docs/reference/support/support-versions/)
      formatted as "[major]-[minor]"; for example "1-27"
      * `--baremetal-config`: Bare metal config file

      ```bash
      image-builder build --os redhat --hypervisor baremetal --release-channel 1-28 --baremetal-config baremetal.json
      ```

1. To consume the image, serve it from an accessible web server, then create the [bare metal cluster spec]({{< relref "../getting-started/baremetal/bare-spec/" >}})
   configuring the `osImageURL` field URL of the image. For example:

   ```
   osImageURL: "http://<artifact host address>/my-ubuntu-v1.23.9-eks-a-17-amd64.gz"
   ```

   See descriptions of [osImageURL]({{< relref "../getting-started/baremetal/bare-spec/#osimageurl" >}}) for further information.

### Build CloudStack node images

These steps use `image-builder` to create a RHEL-based image for CloudStack. Before proceeding, ensure that the above system-level, network-level and CloudStack-specific [prerequisites]({{< relref "#prerequisites">}}) have been met.

1. Create a Linux user for running image-builder.
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
   sudo apt install jq make qemu-kvm libvirt-daemon-system libvirt-clients virtinst cpu-checker libguestfs-tools libosinfo-bin unzip -y
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
   sudo dnf install jq make qemu-kvm libvirt virtinst cpu-checker libguestfs-tools libosinfo unzip wget -y
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
   sudo yum install jq make qemu-kvm libvirt libvirt-clients libguestfs-tools unzip wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   sudo usermod -a -G kvm $USER
   sudo chmod 666 /dev/kvm
   sudo chown root:kvm /dev/kvm
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}

   * Starting with `image-builder` version `v0.3.0`, the minimum required Python version is Python 3.9. However, many Linux distros ship only up to Python 3.8, so you will need to install Python 3.9 from external sources. Refer to the `pyenv` [installation](https://github.com/pyenv/pyenv#installation) and [usage](https://github.com/pyenv/pyenv#usage) documentation to install Python 3.9 and make it the default Python version.
   * Once you have Python 3.9, you can install Ansible using `pip`.
     ```bash
     python3 -m pip install --user ansible
     ```

1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo install -m 0755 ./image-builder /usr/local/bin/image-builder
   cd -
   ```
1. Create a CloudStack configuration file (for example, `cloudstack.json`) to provide the location of a Red Hat Enterprise Linux 8 ISO image and related checksum and Red Hat subscription information:
   ```json
   {
     "iso_url": "<https://endpoint to RHEL ISO endpoint or path to file>",
     "iso_checksum": "<for example: ea5f349d492fed819e5086d351de47261c470fc794f7124805d176d69ddf1fcd>",
     "iso_checksum_type": "<for example: sha256>",
     "rhel_username": "<RHSM username>",
     "rhel_password": "<RHSM password>"
   }
   ```
   >**_NOTE_**: To build the RHEL-based image, `image-builder` temporarily consumes a Red Hat subscription. That subscription is removed once the image is built.

1. To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--os-version`: `8` (default: `8`)
      * `--hypervisor`: For CloudStack use `cloudstack`
      * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
      * `--cloudstack-config`: CloudStack configuration file (`cloudstack.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor cloudstack --release-channel 1-28 --cloudstack-config cloudstack.json
      ```

1. To consume the resulting RHEL-based image, add it as a template to your CloudStack setup as described in [Preparing CloudStack]({{< relref "../getting-started/cloudstack/cloudstack-preparation" >}}).

### Build Snow node images

These steps use `image-builder` to create an Ubuntu-based Amazon Machine Image (AMI) that is backed by EBS volumes for Snow. Before proceeding, ensure that the above system-level, network-level and AMI-specific [prerequisites]({{< relref "#prerequisites">}}) have been met

1. Create a Linux user for running image-builder.
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
   sudo apt install jq unzip make -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}

   * Starting with `image-builder` version `v0.3.0`, the minimum required Python version is Python 3.9. However, many Linux distros ship only up to Python 3.8, so you will need to install Python 3.9 from external sources. Refer to the `pyenv` [installation](https://github.com/pyenv/pyenv#installation) and [usage](https://github.com/pyenv/pyenv#usage) documentation to install Python 3.9 and make it the default Python version.
   * Once you have Python 3.9, you can install Ansible using `pip`.
     ```bash
     python3 -m pip install --user ansible
     ```

1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo install -m 0755 ./image-builder /usr/local/bin/image-builder
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
   * `--os-version`: `20.04` or `22.04` (default: `20.04`)
   * `--hypervisor`: For AMI, use `ami`
   * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
   * `--ami-config`: AMI configuration file (`ami.json` in this example)

   ```bash
   image-builder build --os ubuntu --hypervisor ami --release-channel 1-28 --ami-config ami.json
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

1. Download an [Ubuntu cloud image](https://cloud-images.ubuntu.com/releases) or [RHEL cloud image](https://access.redhat.com/downloads/content/rhel) pertaining to your desired OS and OS version and upload it to the AOS Image Service using Prism. You will need to specify the image's name in AOS as the `source_image_name` in the `nutanix.json` config file specified below. You can also skip this step and directly use the `image_url` field in the config file to provide the URL of a publicly accessible image as source.

1. Create a Linux user for running image-builder.
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
   sudo apt install jq unzip make -y
   sudo snap install yq
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="RHEL" lang="bash" >}}
   sudo dnf update -y
   sudo dnf install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< tab header="Amazon Linux 2" lang="bash" >}}
   sudo yum update -y
   sudo yum install jq unzip make wget -y
   sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
   mkdir -p /home/$USER/.ssh
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   {{< /tab >}}

   {{< /tabpane >}}

   * Starting with `image-builder` version `v0.3.0`, the minimum required Python version is Python 3.9. However, many Linux distros ship only up to Python 3.8, so you will need to install Python 3.9 from external sources. Refer to the `pyenv` [installation](https://github.com/pyenv/pyenv#installation) and [usage](https://github.com/pyenv/pyenv#usage) documentation to install Python 3.9 and make it the default Python version.
   * Once you have Python 3.9, you can install Ansible using `pip`.
     ```bash
     python3 -m pip install --user ansible
     ``` 

1. Get `image-builder`:

   Using the latest EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")
   ```

   OR

   Using a specific EKS Anywhere version
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0
   ```

   ```bash
   cd /tmp
   BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   IMAGEBUILDER_TARBALL_URI=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksD.imagebuilder.uri")
   curl -s $IMAGEBUILDER_TARBALL_URI | tar xz ./image-builder
   sudo install -m 0755 ./image-builder /usr/local/bin/image-builder
   cd -
   ```
1. Create a `nutanix.json` config file. More details on values can be found in the [image-builder documentation](https://image-builder.sigs.k8s.io/capi/providers/nutanix.html). See example below:
   ```json
   {
     "nutanix_cluster_name": "Name of PE Cluster",
     "source_image_name": "Name of Source Image",
     "image_name": "Name of Destination Image",
     "image_url": "URL where the source image is hosted",
     "image_export": "Exports the raw image to disk if set to true",
     "nutanix_subnet_name": "Name of Subnet",
     "nutanix_endpoint": "Prism Central IP / FQDN",
     "nutanix_insecure": "false",
     "nutanix_port": "9440",
     "nutanix_username": "PrismCentral_Username",
     "nutanix_password": "PrismCentral_Password"
   }
   ```

   For RHEL images, add the following fields:
   ```json
   {
     "rhel_username": "<RHSM username>",
     "rhel_password": "<RHSM password>"
   }
   ```
1. Create an Ubuntu or Redhat image:

   **Ubuntu**

   To create an Ubuntu-based image, run `image-builder` with the following options:

      * `--os`: `ubuntu`
      * `--os-version`: `20.04` or `22.04` (default: `20.04`)
      * `--hypervisor`: For Nutanix use `nutanix`
      * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
      * `--nutanix-config`: Nutanix configuration file (`nutanix.json` in this example)

      ```bash
      image-builder build --os ubuntu --hypervisor nutanix --release-channel 1-28 --nutanix-config nutanix.json
      ```

   **Red Hat Enterprise Linux**

   To create a RHEL-based image, run `image-builder` with the following options:

      * `--os`: `redhat`
      * `--os-version`: `8` or `9` (default: `8`)
      * `--hypervisor`: For Nutanix use `nutanix`
      * `--release-channel`: Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28.
      * `--nutanix-config`: Nutanix configuration file (`nutanix.json` in this example)

      ```bash
      image-builder build --os redhat --hypervisor nutanix --release-channel 1-28 --nutanix-config nutanix.json
      ```

### Configuring OS version

`image-builder` supports an `os-version` option that allows you to configure which version of the OS you wish to build. If no OS version is supplied, it will build the default for that OS, according to the table below.

<table style>
    <thead align="center">
        <tr>
            <th><center>Operating system</th>
            <th><center>Supported versions</th>
            <th><center>Corresponding <code>os-version</code> value</th>
            <th><center>Default <code>os-version</code> value</th>
            <th><center>Hypervisors supported</th>
        </tr>
    </thead>
    <tbody align="center">
        <tr >
            <td rowspan=2><b>Ubuntu</b></td>
            <td>20.04.6</td>
            <td>20.04</td>
            <td rowspan=2>20.04</td>
            <td rowspan=2>All hypervisors except CloudStack</td>
        </tr>
        <tr>
            <td>22.04.3</td>
            <td>22.04</td>
        </tr>
        <tr >
            <td rowspan=2><b>RHEL</b></td>
            <td>8.8</td>
            <td>8</td>
            <td rowspan=2>8</td>
            <td>All hypervisors except AMI</td>
        </tr>
        <tr>
            <td>9.2</td>
            <td>9</td>
            <td>Nutanix only</td>
        </tr>
    </tbody>
</table>

Currently, Ubuntu is the only operating system that supports multiple `os-version` values.

### Building images for a specific EKS Anywhere version

This section provides information about the relationship between `image-builder` and EKS Anywhere CLI version, and provides instructions on building images pertaining to a specific EKS Anywhere version.

Every release of EKS Anywhere includes a new version of `image-builder` CLI. For EKS-A releases prior to `v0.16.3`, the corresponding `image-builder` CLI builds images for the latest version of EKS-A released thus far. The EKS-A version determines what artifacts will be bundled into the final OS image, i.e., the core Kubernetes components vended by EKS Distro as well as several binaries vended by EKS Anywhere, such as `crictl`, `etcdadm`, etc, and users may not always want the latest versions of these, and rather wish to bake in certain specific component versions into the image.

This was improved in `image-builder` released with EKS-A `v0.16.3` to `v0.16.5`. Now you can override the default latest build behavior to build images corresponding to a specific EKS-A release, including previous releases. This can be achieved by setting the `EKSA_RELEASE_VERSION` environment variable to the desired EKS-A release (`v0.16.0` and above).
For example, if you want to build an image for EKS-A version `v0.16.5`, you can run the following command.
   ```bash
   export EKSA_RELEASE_VERSION=v0.16.5
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json
   ```

With `image-builder` versions `v0.2.1` and above (released with EKS-A version `v0.17.0`), the `image-builder` CLI has the EKS-A version baked into it, so it will build images pertaining to that release of EKS Anywhere by default. You can override the default version using the `eksa-release` option.
   ```bash
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json --eksa-release v0.16.5
   ```

#### Building images corresponding to dev versions of EKS-A
{{% alert title="Note" color="warning" %}}
Please note that this is not a recommended method for building production images. Images built using this method use development versions of EKS-A and its dependencies. Consider this as an advanced use-case and proceed at your own discretion.
{{% /alert %}}

`image-builder` also provides the option to build images pertaining to dev releases of EKS-A. In the above cases, using a production release of `image-builder` leads to manifests and images being sourced from production locations. While this is usually the desired behavior, it is sometimes useful to build images pertaining to the development branch. Often, new features or enhancements are added to `image-builder` or other EKS-A dependency projects, but are only released to production weeks or months later, based on the release cadence. In other cases, users may want to build EKS-A node images for new Kubernetes versions that are available in dev EKS-A releases but have not been officially released yet. This feature of `image-builder` supports both these use-cases and other similar ones.

This can be achieved using an `image-builder` CLI that has the dev version of EKS-A (`v0.0.0-dev`) baked into it, or by passing in `v0.0.0-dev` to the `eksa-release` option. For both these methods, you need to set the environment `EKSA_USE_DEV_RELEASE` to `true`.

   **`image-builder` obtained from a production EKS-A release:**
   ```bash
   export EKSA_USE_DEV_RELEASE=true
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json --eksa-release v0.0.0-dev
   ```
   **`image-builder` obtained from a dev EKS-A release:**
   ```bash
   export EKSA_USE_DEV_RELEASE=true
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json
   ```

In both these above approaches, the artifacts embedded into the images will be obtained from the dev release bundle manifest instead of production. This manifest contains the latest artifacts built from the `main` branch, and is generally more up-to-date than production release artifact versions.

### UEFI support

`image-builder` supports UEFI-enabled images for Ubuntu OVA and Raw images. UEFI is turned on by default for Ubuntu Raw image builds, but the default firmware for Ubuntu OVAs is BIOS. This can be toggled with the `firmware` option.

For example, to build a Kubernetes v1.27 Ubuntu 22.04 OVA with UEFI enabled, you can run the following command.
   ```bash
   image-builder build --os ubuntu --hypervisor vsphere --os-version 22.04 --release-channel 1.27 --vsphere-config config.json --firmware efi
   ```

The table below shows the possible firmware options for the hypervisor and OS combinations that `image-builder` supports.

|            |       vSphere       |      Baremetal      | CloudStack | Nutanix | Snow |
|:----------:|:-------------------:|:-------------------:|:----------:|:-------:|:----:|
| **Ubuntu** | bios (default), efi | bios, efi (default) |    bios    |   bios  | bios |
|  **RHEL**  |         bios        |         bios        |    bios    |   bios  | bios |

### Mounting additional files

`image-builder` allows you to customize your image by adding files located on your host onto the image at build time. This is helpful when you want your image to have a custom DNS resolver configuration, systemd service unit-files, custom scripts and executables, etc. This option is suppported for all OS and Hypervisor combinations.

To do this, create a configuration file (say, `files.json`) containing the list of files you want to copy:
   ```json
   {
      "additional_files_list": [
         {
            "src": "<Absolute path of the file on the host machine>",
            "dest": "<Absolute path of the location you want to copy the file to on the image",
            "owner": "<Name of the user that should own the file>",
            "group": "<Name of the group that should own the file>",
            "mode": "<The permissions to apply to the file on the image>"
         },
         ...
      ]
   }
   ```

You can now run the `image-builder` CLI with the `files-config` option, with this configuration file as input.
   ```bash
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json --files-config files.json
   ```

### Using Proxy Server

`image-builder` supports proxy-enabled build environments. In order to use proxy server to route outbound requests to the Internet, add the following fields to the hypervisor or provider configuration file (e.g. `baremetal.json`)

  ```json
   {
     "http_proxy": "<http proxy endpoint, for example, http://username:passwd@proxyhost:port>",
     "https_proxy": "<https proxy endpoint, for example, https://proxyhost:port/>",
     "no_proxy": "<optional comma seperated list of domains that should be excluded from proxying>"
  }
  ```
In a proxy-enabled environment, `image-builder` uses `wget` to download artifacts instead of `curl`, as `curl` does not support reading proxy environment variables. In order to add `wget` to the node OS, add the following to the above json configuration file:
   ```json
  {
      "extra_rpms": "wget" #If the node OS being built is RedHat
      "extra_debs": "wget" #If the node OS being built is Ubuntu
  }
  ```

Run `image-builder` CLI with the hypervisor configuration file
  ```bash
  image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json
  ```

### Red Hat Satellite Support

While building Red Hat node images, `image-builder` uses public Red Hat subscription endpoints to register the build virtual machine with the provided Red Hat account and download required packages.

Alternatively, `image-builder` can also use a private Red Hat Satellite to register the build virtual machine and pull packages from the Satellite. 
In order to use Red Hat Satellite in the image build process follow the steps below.

#### Prerequisites
1. Ensure the host running `image-builder` has bi-directional network connectivity with the RedHat Satellite
2. Image builder flow only supports RedHat Satellite version >= 6.8
3. Add the following Red Hat repositories for the latest 8.x or 9.x (for Nutanix) version on the Satellite and initiate a sync to replicate required packages
   1. Base OS Rpms
   2. Base OS - Extended Update Support Rpms
   3. AppStream - Extended Update Support Rpms
4. Create an activation key on the Satellite and ensure library environment is enabled

#### Build Red Hat node images using Red Hat Satellite
1. Add the following fields to the hypervisor or provider configuration file
   ```json
   {
     "rhsm_server_hostname": "fqdn of Red Hat Satellite server",
     "rhsm_server_release_version": "Version of Red hat OS Packages to pull from Satellite. e.x. 8.8",
     "rhsm_activation_key": "activation key from Satellite",
     "rhsm_org_id": "org id from Satellite"
   }
   ```
   `rhsm_server_release_version` should always point to the latest 8.x or 9.x minor Red Hat release synced and available on Red Hat Satellite
2. Run `image-builder` CLI with the hypervisor configuration file
   ```bash
   image-builder build --os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json
   ```

### Air Gapped Image Building
`image-builder` supports building node OS images in an air-gapped environment. Currently only building Ubuntu-based node OS images for baremetal provider is supported in air-gapped building mode.

#### Prerequisites
1. Air-gapped image building requires
   - private artifacts server e.g. artifactory from JFrog 
   - private git server. 
3. Ensure the host running `image-builder` has bi-directional network connectivity with the artifacts server and git server 
4. Artifacts server should have the ability to host and serve, standalone artifacts and Ubuntu OS packages 

#### Building node images in an air-gapped environment
1. Identify the EKS-D release channel (generally aligning with Kubernetes version) to build. For example, 1.27 or 1.28
2. Identify the latest release of EKS-A from [changelog]({{< ref "/docs/whatsnew/changelog" >}}). For example, v0.18.0
3. Run `image-builder` CLI to download manifests in an environment with internet connectivity
   ```bash
   image-builder download manifests
   ```
   This command will download a tarball containing all currently released and supported EKS-A and EKS-D manifests that are required for image building in an air-gapped environment.
3. Create a local file named `download-airgapped-artifacts.sh` with the contents below. This script will download the required EKS-A and EKS-D artifacts required for image building.
   <details>
      <summary>Click to expand download-airgapped-artifacts.sh script</summary>
   
      ```shell
      #!/usr/bin/env bash
      set +o nounset
   
      function downloadArtifact() {
         local -r artifact_url=${1}
         local -r artifact_path_pre=${2}
   
         # Removes hostname from url
         artifact_path_post=$(echo ${artifact_url} | sed -E 's:[^/]*//[^/]*::')
         artifact_path="${artifact_path_pre}${artifact_path_post}"
         curl -sL ${artifact_url} --output ${artifact_path} --create-dirs
      }
   
      if [ -z "${EKSA_RELEASE_VERSION}" ]; then
         echo "EKSA_RELEASE_VERSION not set. Please refer https://anywhere.eks.amazonaws.com/docs/whatsnew/ or https://github.com/aws/eks-anywhere/releases to get latest EKS-A release"
         exit 1
      fi
   
      if [ -z "${RELEASE_CHANNEL}" ]; then
         echo "RELEASE_CHANNEL not set. Supported EKS Distro releases include 1-24, 1-25, 1-26, 1-27 and 1-28"
         exit 1
      fi
   
      # Convert RELEASE_CHANNEL to dot schema
      kube_version="${RELEASE_CHANNEL/-/.}"
      echo "Setting Kube Version: ${kube_version}"
   
      # Create a local directory to download the artifacts
      artifacts_dir="eks-a-d-artifacts"
      eks_a_artifacts_dir="eks-a-artifacts"
      eks_d_artifacts_dir="eks-d-artifacts"
      echo "Creating artifacts directory: ${artifacts_dir}"
      mkdir ${artifacts_dir}
   
      # Download EKS-A bundle manifest
      cd ${artifacts_dir}
      bundles_url=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
      echo "Identified EKS-A Bundles URL: ${bundles_url}"
      echo "Downloading EKS-A Bundles manifest file"
      bundles_file_data=$(curl -sL "${bundles_url}" | yq)
   
      # Download EKS-A artifacts
      eks_a_artifacts="containerd crictl etcdadm"
      for eks_a_artifact in ${eks_a_artifacts}; do
         echo "Downloading EKS-A artifact: ${eks_a_artifact}"
         artifact_url=$(echo "${bundles_file_data}" | yq e ".spec.versionsBundles[] | select(.kubeVersion==\"${kube_version}\").eksD.${eks_a_artifact}.uri" -)
         downloadArtifact ${artifact_url} ${eks_a_artifacts_dir}
      done
   
      # Download EKS-D artifacts
      echo "Downloading EKS-D manifest file"
      eks_d_manifest_url=$(echo "${bundles_file_data}" | yq e ".spec.versionsBundles[] | select(.kubeVersion==\"${kube_version}\").eksD.manifestUrl" -)
      eks_d_manifest_file_data=$(curl -sL "${eks_d_manifest_url}" | yq)
   
      # Get EKS-D kubernetes base url from kube-apiserver
      eks_d_kube_tag=$(echo "${eks_d_manifest_file_data}" | yq e ".status.components[] | select(.name==\"kubernetes\").gitTag" -)
      echo "EKS-D Kube Tag: ${eks_d_kube_tag}"
      api_server_artifact="bin/linux/amd64/kube-apiserver.tar"
      api_server_artifact_url=$(echo "${eks_d_manifest_file_data}" | yq e ".status.components[] | select(.name==\"kubernetes\").assets[] | select(.name==\"${api_server_artifact}\").archive.uri")
      eks_d_base_url=$(echo "${api_server_artifact_url}" | sed -E "s,/${eks_d_kube_tag}/${api_server_artifact}.*,,")
      echo "EKS-D Kube Base URL: ${eks_d_base_url}"
   
      # Downloading EKS-D Kubernetes artifacts
      eks_d_k8s_artifacts="kube-apiserver.tar kube-scheduler.tar kube-controller-manager.tar kube-proxy.tar pause.tar coredns.tar etcd.tar kubeadm kubelet kubectl"
      for eks_d_k8s_artifact in ${eks_d_k8s_artifacts}; do
         echo "Downloading EKS-D artifact: Kubernetes - ${eks_d_k8s_artifact}"
         artifact_url="${eks_d_base_url}/${eks_d_kube_tag}/bin/linux/amd64/${eks_d_k8s_artifact}"
         downloadArtifact ${artifact_url} ${eks_d_artifacts_dir}
      done
   
      # Downloading EKS-D etcd artifacts
      eks_d_extra_artifacts="etcd cni-plugins"
      for eks_d_extra_artifact in ${eks_d_extra_artifacts}; do
         echo "Downloading EKS-D artifact: ${eks_d_extra_artifact}"
         eks_d_artifact_tag=$(echo "${eks_d_manifest_file_data}" | yq e ".status.components[] | select(.name==\"${eks_d_extra_artifact}\").gitTag" -)
         artifact_url=$(echo "${eks_d_manifest_file_data}" | yq e ".status.components[] | select(.name==\"${eks_d_extra_artifact}\").assets[] | select(.name==\"${eks_d_extra_artifact}-linux-amd64-${eks_d_artifact_tag}.tar.gz\").archive.uri")
         downloadArtifact ${artifact_url} ${eks_d_artifacts_dir}
      done

      ```
   
   </details>
4. Change mode of the saved file `download-airgapped-artifacts.sh` to an executable
   ```bash
   chmod +x download-airgapped-artifacts.sh
   ```
5. Set EKS-A release version and EKS-D release channel as environment variables and execute the script
   ```bash
   EKSA_RELEASE_VERSION=v0.18.0 RELEASE_CHANNEL=1-28 ./download-airgapped-artifacts.sh
   ```
   Executing this script will create a local directory `eks-a-d-artifacts` and download the required EKS-A and EKS-D artifacts.
6. Create two repositories, one for EKS-A and one for EKS-D on the private artifacts server.
   Upload the contents of `eks-a-d-artifacts/eks-a-artifacts` to the EKS-A repository. Similarly upload the contents of `eks-a-d-artifacts/eks-d-artifacts` to the EKS-D repository on the private artifacts server.
   Please note, the path of artifacts inside the downloaded directories must be preserved while hosted on the artifacts server.
7. Download and host the base ISO image to the artifacts server.
8. Replicate the following public git repositories to private artifacts server or git servers. Make sure to sync all branches and tags to the private git repo.
   - [eks-anywhere-build-tooling](https://github.com/aws/eks-anywhere-build-tooling)
   - [image-builder](https://github.com/kubernetes-sigs/image-builder)
9. Replicate public Ubuntu packages to private artifacts server. Please refer your artifact server's documentation for more detailed instructions.
10. Create a sources.list file that will configure apt commands to use private artifacts server for OS packages
    ```bash
    deb [trusted=yes] http://<private-artifacts-server>/debian focal main restricted universe multiverse
    deb [trusted=yes] http://<private-artifacts-server>/debian focal-updates main restricted universe multiverse
    deb [trusted=yes] http://<private-artifacts-server>/debian focal-backports main restricted universe multiverse
    deb [trusted=yes] http://<private-artifacts-server>/debian focal-security main restricted universe multiverse
    ```
    `focal` in the above file refers to the name of the Ubuntu OS for version 20.04. If using Ubuntu version 22.04 replace `focal` with `jammy`.
11. Create a provider or hypervisor configuration file and add the following fields
    ```json
    {
       "eksa_build_tooling_repo_url": "https://internal-git-host/eks-anywhere-build-tooling.git",
       "image_builder_repo_url": "https://internal-repos/image-builder.git",
       "private_artifacts_eksd_fqdn": "http://private-artifacts-server/artifactory/eks-d-artifacts",
       "private_artifacts_eksa_fqdn": "http://private-artifacts-server:8081/artifactory/eks-a-artifacts",
       "extra_repos": "<full path of sources.list>",
       "disable_public_repos": "true",
       "iso_url": "http://<private-base-iso-url>/ubuntu-20.04.1-legacy-server-amd64.iso",
       "iso_checksum": "<sha256 of the base iso",
       "iso_checksum_type": "sha256"
    }
    ```
12. Run `image-builder` CLI with the hypervisor configuration file and the downloaded manifest tarball
    ```bash
    image-builder build -os <OS> --hypervisor <hypervisor> --release-channel <release channel> --<hypervisor>-config config.json --airgapped --manifest-tarball <path to eks-a-manifests.tar>
    ```
    

## Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
