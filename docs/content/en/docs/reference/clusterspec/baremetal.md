---
title: "Bare metal configuration"
linkTitle: "Bare metal configuration"
weight: 10
description: >
  Full EKS Anywhere configuration reference for a Bare Metal cluster.
---

This is a generic template with detailed descriptions below for reference.
The following additional optional configuration can also be included:

* [CNI]({{< relref "optional/cni.md" >}})
* [multus]({{< relref "optional/multus.md" >}})

To generate your own cluster configuration, follow instructions from the Bare Metal [Create production cluster]({{< relref "../../getting-started/production-environment/" >}}) section and modify it using descriptions below.
For information on how to add cluster configuration settings to this file for advanced node configuration, see [Advanced Bare Metal cluster configuration]({{< relref "#advanced-bare-metal-cluster-configuration" >}}).


```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  clusterNetwork:
    cniConfig:
      cilium: {}
    pods:
      cidrBlocks:
      - 192.168.0.0/16
    services:
      cidrBlocks:
      - 10.96.0.0/12
  controlPlaneConfiguration:              
    count: 1
    endpoint:
      host: "193.17.0.50"
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: my-cluster-name-cp
  datacenterRef:
    kind: TinkerbellDatacenterConfig
    name: my-cluster-name
  kubernetesVersion: "1.22"
  managementCluster:
    name: my-cluster-name
  workerNodeGroupConfigurations:
  - count: 1
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: my-cluster-name
    name: md-0

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellDatacenterConfig
metadata:
  name: my-cluster-name
spec:
  tinkerbellIP: "193.17.0.50"

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  hardwareSelector: {}
  osFamily: ubuntu
  templateRef: {}
  users:
  - name: ec2-user
    sshAuthorizedKeys:
    - ssh-rsa AAAAB3NzaC1yc2... jwjones@833efcab1482.home.example.com

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: my-cluster-name
spec:
  hardwareSelector: {}
  osFamily: ubuntu
  templateRef:
    kind: TinkerbellTemplateConfig
    name: my-cluster-name
  users:
  - name: ec2-user
    sshAuthorizedKeys:
    - ssh-rsa AAAAB3NzaC1yc2... jwjones@833efcab1482.home.example.com
```

## Cluster Fields

### name (required)
Name of your cluster (`my-cluster-name` in this example).

### clusterNetwork (required)
Specific network configuration for your Kubernetes cluster.

### clusterNetwork.cniConfig (required)
CNI plugin to be installed in the cluster. The only supported value at the moment is `cilium`.

### clusterNetwork.pods.cidrBlocks[0] (required)
Subnet used by pods in CIDR notation. Please note that only 1 custom pods CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the machines.

### clusterNetwork.services.cidrBlocks[0] (required)
Subnet used by services in CIDR notation. Please note that only 1 custom services CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the machines.

### clusterNetwork.dns.resolvConf.path (optional)
Path to the file with a custom DNS resolver configuration.

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other machines.

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration for your nodes. See `TinkerbellMachineConfig Fields` below.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. 

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint, `node-role.kubernetes.io/master`. The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint `node-role.kubernetes.io/master`.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
> 

### controlPlaneConfiguration.labels
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

### datacenterRef
Refers to the Kubernetes object with Tinkerbell-specific configuration. See `TinkerbellDatacenterConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.22`, `1.21`, `1.20`

### managementCluster
Identifies the name of the management cluster.
If this is a standalone cluster or if it were serving as the management cluster for other workload clusters, this will be the same as the cluster name.
Bare Metal EKS Anywhere clusters do not yet support the creation of separate workload clusters.

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers.
You may define one or more worker node groups.

### workerNodeGroupConfigurations.count (required)
Number of worker nodes

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration for your nodes. See `TinkerbellMachineConfig Fields` below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

## TinkerbellDatacenterConfig Fields

### tinkerbellIP
Optional field to identify the IP address of the Tinkerbell service.
Other TinkerbellDatacenterConfig fields are not yet supported.

## TinkerbellMachineConfig Fields
In the example, there are `TinkerbellMachineConfig` sections for control plane (`my-cluster-name-cp`) and worker (`my-cluster-name`) machine groups.
The following fields identify information needed to configure the nodes in each of those groups.
>**_NOTE:_** Currently, you can only have one machine group for all machines in the control plane and one for all machines in the worker group.
>
### hardwareSelector
Use fields under `hardwareSelector` to add key/value pair labels to match particular machines that you identified in the CSV file where you defined the machines in your cluster.
Choose any label name you like.
For example, if you had added the label `node=cp-machine` to the machines listed in your CSV file that you want to be control plane nodes, the following `hardwareSelector` field would cause those machines to be added to the control plane:
```bash
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  hardwareSelector:
    node: "cp-machine"
```
### osFamily (required)
Operating system on the machine. For example, `bottlerocker` or `ubuntu`.
### templateRef
Identifies the template that defines the actions that will be applied to the TinkerbellMachineConfig.
See TinkerbellTemplateConfig fields below.

### users
The name of the user you want to configure to access your virtual machines through SSH.

The default is `ec2-user`.
Currently, only one user is supported.

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your machines through SSH (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster machines so you can SSH into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<machine-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value.

## Advanced Bare Metal cluster configuration

When you generate a Bare Metal cluster configuration, the `TinkerbellTemplateConfig` is kept internally and not shown in the generated configuration file.
`TinkerbellTemplateConfig` settings define the actions done to install each node, such as get installation media, configure networking, add users, and otherwise configure the node.

Advanced users can override the default values set for `TinkerbellTemplateConfig`.
They can also add their own [Tinkerbell actions](https://docs.tinkerbell.org/actions/action-architecture/) to make personalized modifications to EKS Anywhere nodes.

The following shows two `TinkerbellTemplateConfig` examples that you can add to your cluster configuration file to override the values that EKS Anywhere sets: one for Ubuntu and one for Bottlerocket.
Most actions used differ for different operating systems.

### Ubuntu TinkerbellTemplateConfig example

```
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellTemplateConfig
metadata:
  name: my-cluster-name
spec:
  template:
    global_timeout: 6000
    id: ""
    name: my-cluster-name
    tasks:
    - actions:
      - environment:
          COMPRESSED: "true"
          DEST_DISK: /dev/sda
          IMG_URL: https://.../ubuntu-v1.22.9-eks-d...
        image: public.ecr.aws/.../image2disk:6c0f0d437bde2c...
        name: stream-image
        timeout: 360
      - environment:
          CONTENTS: |
            network:
              version: 2
              renderer: networkd
              ethernets:
                  eno1:
                      dhcp4: true
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/netplan/config.yaml
          DIRMODE: "0755"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/.../writefile:6c0f0d437bde2c...
        name: write-netplan
        timeout: 90
      - environment:
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: []
                strict_id: false
            system_info:
              default_user:
                name: tink
                groups: [wheel, adm]
                sudo: ["ALL=(ALL) NOPASSWD:ALL"]
                shell: /bin/bash
            manage_etc_hosts: localhost
            warnings:
              dsid_missing_source: off
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/.../writefile:6c0f0d437bde2c...
        name: add-tink-cloud-init-config
        timeout: 90
      - environment:
          CONTENTS: |
            datasource: Ec2
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/ds-identify.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/.../writefile:6c0f0d437bde2c...
        name: add-tink-cloud-init-ds-config
        timeout: 90
      - environment:
          BLOCK_DEVICE: /dev/sda2
          FS_TYPE: ext4
        image: public.ecr.aws/.../kexec:6c0f0d437bde2c...
        name: kexec-image
        pid: host
        timeout: 90
      name: my-cluster-name
      volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
      worker: '{{.device_1}}'
    version: "0.1"
```

### Bottlerocket TinkerbellTemplateConfig example

```
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellTemplateConfig
metadata:
  name: my-cluster-name
spec:
  template:
    global_timeout: 6000
    id: ""
    name: my-cluster-name
    tasks:
    - actions:
      - environment:
          COMPRESSED: "true"
          DEST_DISK: /dev/sda
          IMG_URL: https://.../bottlerocket-metal-k8s-1.22-x86_64-1.7.2-cf824404.img
        image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/image2disk:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-v0.0.0-dev-build.2166
        name: stream-image
        timeout: 360
      - environment:
          BOOTCONFIG_CONTENTS: |
            kernel {
                console = "tty0", "ttyS0,115200n8"
            }
          DEST_DISK: /dev/sda12
          DEST_PATH: /bootconfig.data
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-v0.0.0-dev-build.2878
        name: write-bootconfig
        timeout: 90
      - environment:
          CONTENTS: |
            # Version is required, it will change as we support
            # additional settings
            version = 1
            # "eno1" is the interface name
            # Users may turn on dhcp4 and dhcp6 via boolean
            [eno1]
            dhcp4 = true
            # Define this interface as the "primary" interface
            # for the system.  This IP is what kubelet will use
            # as the node IP.  If none of the interfaces has
            # "primary" set, we choose the first interface in
            # the file
            primary = true
          DEST_DISK: /dev/sda12
          DEST_PATH: /net.toml
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-v0.0.0-dev-build.2878
        name: write-netconfig
        timeout: 90
      - environment:
          HEGEL_URL: http://<hegel-ip>:50061
          DEST_DISK: /dev/sda12
          DEST_PATH: /user-data.toml
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-v0.0.0-dev-build.2878
        name: write-user-data
        timeout: 90
      - name: "reboot"
        image: public.ecr.aws/t0n3a9y4/reboot-action:latest
        timeout: 90
        volumes:
          - /worker:/worker
    version: "0.1"
```
## TinkerbellTemplateConfig Fields

The values in the `TinkerbellTemplateConfig` fields are created from the contents of the CSV file used to generate a configuration.
The template contains actions that are performed on a Bare Metal machine when it first boots up to be provisioned.
For advanced users, you can add these fields to your cluster configuration file if you have special needs to do so.

While there are a fields that apply to all provisioned operating systems, actions are specific to each operating system.
Examples below describe actions for Ubuntu and Bottlerocket operating systems.

### template.global_timeout

Sets the timeout value for completing the configuration. Set to 6000 (100 minutes) by default.

### template.id

Not set by default.

### template.tasks

Within the TinkerbellTemplateConfig `template` under `tasks` is a set of actions.
The following descriptions cover the actions shown in the example templates for Ubuntu and Bottlerocket:

### template.tasks.actions.name.stream-image (Ubuntu and Bottlerocket)
The `stream-image` action streams the selected image to the machine you are provisioning. It identifies:

* environment.COMPRESSED: When set to `true`, Tinkerbell expects `IMG_URL` to be a compressed image, which Tinkerbell will uncompress when it writes the contents to disk.
* environment.DEST_DISK: The hard disk on which the operating system is desployed. The default is the first SCSI disk (/dev/sda), but can be changed for other disk types.
* environment.IMG_URL: The operating system tarball (ubuntu or other) to stream to the machine you are configuring.
* image: Container image needed to perform the steps needed by this action.
* timeout: Sets the amount of time (in seconds) that Tinkerbell has to stream the image, uncompress it, and write it to disk before timing out. Consider increasing this limit from the default to a higher limit if this action is timing out.

## Ubuntu-specific actions

### template.tasks.actions.name.write-netplan (Ubuntu)
The `write-netplan` action writes Ubuntu network configuration information to the machine (see [Netplan](https://netplan.io/)) for details. It identifies:

* environment.CONTENTS.network.version: Identifies the network version.
* environment.CONTENTS.network.renderer: Defines the service to manage networking. By default, the `networkd` systemd service is used.
* environment.CONTENTS.network.ethernets: Network interface to external network (eno1, by default) and whether or not to use dhcp4 (true, by default).
* environment.DEST_DISK: Destination block storage device partition where the operating system is copied. By default, /dev/sda2 is used (sda1 is the UEFI partition). 
* environment.DEST_PATH: File where the networking configuration is written (/etc/netplan/config.yaml, by default).
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0755, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0644, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.add-tink-cloud-init-config (Ubuntu)
The `add-tink-cloud-init-config` action configures cloud-init features to further configure the operating system. See [cloud-init Documentation](https://cloudinit.readthedocs.io/en/latest/) for details. It identifies:

* environment.CONTENTS.datasource: Identifies Ec2 (Ec2.metadata_urls) as the data source and sets `Ec2.strict_id: false` to prevent could init from producing warnings about this datasource.
* environment.CONTENTS.system_info: Creates the `tink` user and gives it administrative group privileges (wheel, adm) and passwordless sudo privileges, and set the default shell (/bin/bash).
* environment.CONTENTS.manage_etc_hosts: Updates the system's `/etc/hosts` file with the hostname. Set to `localhost` by default.
* environment.CONTENTS.warnings: Sets dsid_missing_source to `off`.
* environment.DEST_DISK: Destination block storage device partition where the operating system is located (`/dev/sda2`, by default).
* environment.DEST_PATH: Location of the cloud-init configuration file on disk (`/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg`, by default)
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0700, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0600, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.add-tink-cloud-init-ds-config (Ubuntu)
The `add-tink-cloud-init-ds-config` action configures cloud-init data store features. This identifies the location of your metadata source once the machine is up and running. It identifies:

* environment.CONTENTS.datasource: Sets the datasource. Uses Ec2, by default.
* environment.DEST_DISK: Destination block storage device partition where the operating system is located (/dev/sda2, by default).
* environment.DEST_PATH: Location of the data store identity configuration file on disk (/etc/cloud/ds-identify.cfg, by default) 
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0700, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0600, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.kexec-image (Ubuntu)
The `kexec-image` action performs provisioning activities on the machine, then allows kexec to pivot the kernel to use the system installed on disk. This action identifies:

* environment.BLOCK_DEVICE: Disk partition on which the operating system is installed (/dev/sda2, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* image: Container image used to perform the steps needed by this action.
* pid: Process ID. Set to host, by default.
* timeout: Time needed to complete the action, in seconds.
* volumes: Identifies mount points that need to be remounted to point to locations in the installed system.

If your hardware requires a full reboot, you can change the kexec-image setting as follows:

```
actions:
- name: "reboot"
  image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/reboot-action:latest
  timeout: 90
  volumes:
  - /worker:/worker
```

## Bottlerocket-specific actions

### template.tasks.actions.write-bootconfig (Bottlerocket)
The write-bootconfig action identifies the location on the machine to put content needed to boot the system from disk????

* environment.BOOTCONFIG_CONTENTS.kernel: Add kernel parameters that are passed to the kernel when the system boots???
* environment.DEST_DISK: Identifies the block storage device that holds the boot partition???
* environment.DEST_PATH: Identifies the file holding boot configuration data (`/bootconfig.data` in this example).
* environment.DIRMODE: The Linux permissions assigned to the boot directory???
* environment.FS_TYPE: The filesystem type associated with the boot partition???
* environment.GID: The group ID associated with files and directories created on the boot partition???. GID 0 is the root group.
* environment.MODE: The Linux permissions assigned to files in the boot partition???
* environment.UID: The user ID associated with files and directories created on the boot partition???. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.write-netconfig (Bottlerocket)
The write-netconfig action configures networking for the system???

* environment.CONTENTS: Add network values, including: `version = 1` (version number), `[eno1]` (external network interface), `dhcp4 = true` (turns on dhcp4), and `primary = true` (identifies this interface as the primary interface used by kubelet).
* environment.DEST_DISK: Identifies the block storage device that holds the network configuration information ???
* environment.DEST_PATH: Identifies the file holding network configuration data (`/net.toml` in this example).
* environment.DIRMODE: The Linux permissions assigned to the directory holding network configuration settings???
* environment.FS_TYPE: The filesystem type associated with the partition holding network configuration settings???
* environment.GID: The group ID associated with files and directories created on the partition???. GID 0 is the root group.
* environment.MODE: The Linux permissions assigned to files in the partition???
* environment.UID: The user ID associated with files and directories created on the partition???. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.

### template.tasks.actions.write-user-data (Bottlerocket)
The write-user-data action configures the Tinkerbell Hegel service, which provides the metadata store for Tinkerbell.

* environment.HEGEL_URL: The IP address and port number of the Tinkerbell [Hegel](https://docs.tinkerbell.org/services/hegel/) service.
* environment.DEST_DISK: Identifies the block storage device that holds the network configuration information ???
* environment.DEST_PATH: Identifies the file holding network configuration data (`/net.toml` in this example).
* environment.DIRMODE: The Linux permissions assigned to the directory holding network configuration settings???
* environment.FS_TYPE: The filesystem type associated with the partition holding network configuration settings???
* environment.GID: The group ID associated with files and directories created on the partition???. GID 0 is the root group.
* environment.MODE: The Linux permissions assigned to files in the partition???
* environment.UID: The user ID associated with files and directories created on the partition???. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.reboot (Bottlerocket)
The reboot action defines how the system restarts to bring up the installed system.

* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.
* volumes: The volume (directory) to mount into the container from the installed system.??????
### version

Matches the current version of the Tinkerbell template.
