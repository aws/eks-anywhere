---
title: "Artifacts"
linkTitle: "Artifacts"
weight: 55
description: >
  Artifacts associated with this release: OVAs and images.
---

EKS Anywhere supports three different node operating systems:

* Bottlerocket: For vSphere and Bare Metal providers
* Ubuntu: For vSphere, Bare Metal, and Nutanix providers
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

1.24 - `eksdRelease:kubernetes-1-24-eks-3`

1.23 - `eksdRelease:kubernetes-1-23-eks-8`

1.22 - `eksdRelease:kubernetes-1-22-eks-13`

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

The `image-builder` CLI lets you build your own Ubuntu-based vSphere OVAs, Nutanix qcow2 images, RHEL-based qcow2 images, or Bare Metal gzip images to use in EKS Anywhere clusters.
When you run `image-builder` it will pull in all components needed to create images to use for nodes in an EKS Anywhere cluster, including the lastest operating system, Kubernetes, and EKS Distro security updates, bug fixes, and patches.
With this tool, when you build an image you get to choose:

* Operating system type (for example, ubuntu)
* Provider (vsphere, cloudstack, baremetal, ami, nutanix)
* Release channel for EKS Distro (generally aligning with Kubernetes releases)
* vSphere only: configuration file providing information needed to access your vSphere setup
* CloudStack only: configuration file providing information needed to access your Cloudstack setup
* AMI only: configuration file providing information needed to customize your AMI build parameters
* Nutanix only: configuration file providing information needed to access Prism Central

Because `image-builder` creates images in the same way that the EKS Anywhere project does for their own testing, images built with that tool are supported.
The following procedure describes how to use `image-builder` to build images for EKS Anywhere on a vSphere, Bare Metal, or Nutanix provider.

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
  * Prism Central endpoint (Nutanix only)
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
* AMI only: Packer will require prior authentication with your AWS account to launch EC2 instances for the AMI build. See [Authentication guide for Amazon EBS Packer builder](https://developer.hashicorp.com/packer/plugins/builders/amazon#authentication) for possible modes of authentication. We recommend that you run `image-builder` on a pre-existing Ubuntu EC2 instance and use an [IAM instance role with the required permissions](https://developer.hashicorp.com/packer/plugins/builders/amazon#iam-task-or-instance-role).
* Nutanix only: Prism Admin permissions

### Optional Proxy configuration
You can use a proxy server to route outbound requests to the internet. To configure `image-builder` tool to use a proxy server, export these proxy environment variables:
  ```
  export HTTP_PROXY=<HTTP proxy URL e.g. http://proxy.corp.com:80>
  export HTTPS_PROXY=<HTTPS proxy URL e.g. http://proxy.corp.com:443>
  export NO_PROXY=<No proxy>
  ```

### Build vSphere OVA node images

These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for vSphere.

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
1. Get the latest version of `govc`:
   ```
   curl -L -o - "https://github.com/vmware/govmomi/releases/latest/download/govc_$(uname -s)_$(uname -m).tar.gz" | sudo tar -C /usr/local/bin -xvzf - govc
   ```
1. Create a content library on vSphere:
   ```
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
These steps use `image-builder` to create an Ubuntu-based or RHEL-based image for Bare Metal.

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
     "rhel_password": "<rhsm password>",
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

### Building AMI node images

These steps use `image-builder` to create an Ubuntu-based Amazon Machine Image (AMI) that is backed by EBS volumes.

1. Create a linux user for running image-builder.
   ```
   sudo adduser image-builder
   ```
   Follow the prompt to provide a password for the image-builder user.
1. Add the `image-builder` user to the `sudo` group and switch user to `image-builder`, providing in the password from previous step when prompted.
   ```
   sudo usermod -aG sudo image-builder
   su image-builder
   ```
1. Install packages and prepare environment:
   ```
   sudo apt update -y
   sudo apt install gh jq unzip make ansible python3-pip -y
   sudo snap install yq
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   ```
1. Get `image-builder`:
   >**_NOTE_**: The version of `image-builder` CLI that includes support for building AMIs has not yet been released to production, so the steps below correspond to the development version of the CLI.
   
   ```bash
    cd /tmp
    LATEST_WEEKLY_RELEASE=$(gh release list --repo aws/eks-anywhere | grep Weekly | awk 'NR==1 {print $5}')
    gh release download $LATEST_WEEKLY_RELEASE --pattern "image-builder*.tar.gz"
    sudo tar xvf image-builder*.tar.gz
    sudo cp image-builder /usr/local/bin
    ```
1. Create an AMI configuration file (for example, `ami.json`) that contains various AMI parameters.
   ```json
   {
     "ami_filter_name": "<Regular expression to filter a source AMI (default: ubuntu/images/*ubuntu-focal-20.04-amd64-server-*)>",
     "ami_filter_owners": "<AWS account ID or AWS owner alias such as 'amazon', 'aws-marketplace', etc (default: 679593333241 - the AWS Marketplace AWS account ID)>",
     "ami_regions": "<A list of AWS regions to copy the AMI to>",
     "aws_region": "The AWS region in which to launch the EC2 instance to create the AMI",
     "ansible_extra_vars": "<The absolute path to the additional variables to pass to Ansible. These are converted to the `--extra-vars` command-line argument. This path must be prefix with '@'>",
     "builder_instance_type": "<The EC2 instance type to use while building the AMI (default: t3.small)>",
     "custom_role": "<If set to true, this will run a custom Ansible role before the `sysprep` role to allow for further customization>",
     "custom_role_name_list" : "<Array of strings representing the absolute paths of custom Ansible roles to run. This field is mutually exclusive with custom_role_names>",
     "custom_role_names": "<Space-delimited string of the custom roles to run. This field is mutually exclusive with custom_role_name_list and is provided for compatibility with Ansible's input format>",
     "manifest_output": "<The absolute path to write the build artifacts manifest to. If you wish to export the AMI using this manifest, ensure that you provide a path that is not inside the '/home/$USER/eks-anywhere-build-tooling' path since that will be cleaned up when the build finishes>",
     "root_device_name": "<The device name used by EC2 for the root EBS volume attached to the instance>",
     "volume_size": "<The size of the root EBS volume in GiB>",
     "volume_type": "<The type of root EBS volume, such as gp2, gp3, io1, etc>",
   }
   ```
1. To create an Ubuntu-based image, run `image-builder` with the following options:

   * `--os`: `ubuntu`
   * `--hypervisor`: For AMI, use `ami`
   * `--release-channel`: Supported EKS Distro releases include 1-20, 1-21, 1-22, and 1-23.
   * `--ami-config`: AMI configuration file (`ami.json` in this example)

   ```bash
   image-builder build --os ubuntu --hypervisor ami --release-channel 1-23 --ami-config ami.json
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

These steps use `image-builder` to create a Ubuntu-based image for Nutanix AHV and import it into the AOS Image Service.

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
   ```
1. Install packages and prepare environment:
   ```bash
   sudo apt update -y
   sudo apt install jq unzip make ansible -y
   sudo snap install yq
   echo "HostKeyAlgorithms +ssh-rsa" >> /home/$USER/.ssh/config
   echo "PubkeyAcceptedKeyTypes +ssh-rsa" >> /home/$USER/.ssh/config
   ```
1. Get `image-builder`:
    ```bash
    cd /home/$USER
    sudo wget https://anywhere-assets.eks.amazonaws.com/releases/bundles/19/artifacts/image-builder/0.1.2/image-builder-linux-amd64.tar.gz
    sudo tar xvf image-builder*.tar.gz
    sudo cp image-builder /usr/local/bin
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
      * `--release-channel`: Supported EKS Distro releases include 1-21, 1-22, 1-23, and 1-24.
      * `--nutanix-config`: Nutanix configuration file (`nutanix-connection.json` in this example)

      ```bash
      cd /home/$USER
      image-builder build --os ubuntu --hypervisor nutanix --nutanix-config nutanix-connection.json --release-channel 1-24
      ```

## Images

The various images for EKS Anywhere can be found [in the EKS Anywhere ECR repository](https://gallery.ecr.aws/eks-anywhere/).
The various images for EKS Distro can be found [in the EKS Distro ECR repository](https://gallery.ecr.aws/eks-distro/).
