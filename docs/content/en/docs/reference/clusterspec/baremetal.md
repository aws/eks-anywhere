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

To generate your own cluster configuration, follow instructions from [Create production cluster]({{< relref "../../getting-started/production-environment/" >}}) and modify it using descriptions below.

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
  tinkerbellCertURL: ""
  tinkerbellGRPCAuth: ""
  tinkerbellHegelURL: ""
  tinkerbellIP: "193.17.0.50"
  tinkerbellPBnJGRPCAuth: ""

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  osFamily: ubuntu
  templateRef:
    kind: TinkerbellTemplateConfig
    name: my-cluster-name
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
  osFamily: ubuntu
  templateRef:
    kind: TinkerbellTemplateConfig
    name: my-cluster-name
  users:
  - name: ec2-user
    sshAuthorizedKeys:
    - ssh-rsa AAAAB3NzaC1yc2... jwjones@833efcab1482.home.example.com

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

---
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

## TinkerBellMachineConfig Fields

### osFamily (required)
Operating system on the machine. For example, `ubuntu`.

### templateRef
Identifies the template that defines the actions that will be applied to the TinkerbellMachineConfig.
See TinkerbellTemplateConfig fields below.

### users
The name of the user you want to configure to access your virtual machines through SSH.

The default is `ec2-user`.

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your machines through SSH (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster machines so you can SSH into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<machine-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value


## TinkerbellTemplateConfig Fields

The values in the `TinkerbellTemplateConfig` fields are created from the contents of the CSV file used to generate this configuration.
The template contains actions that are performed on a Bare Metal machine when it first boots up to be provisioned.
For advanced users, you can modify these fields if you have special needs to do so.

### template.global_timeout

NEED INFO. Set to 6000 in example (100 minutes?). Different for different OS families?

### template.id

NEED INFO

### template.tasks

Within the TinkerbellTemplateConfig `template` under `tasks` is a set of actions.
The following descriptions cover the actions shown in the example template:

### template.tasks.actions.name.stream-image
The `stream-image` action streams the selected image to machine you are provisioning. It identifies:

* environment.COMPRESSED: NEED INFO
* environment.DEST_DISK: The hard disk on which the operating system is desployed. The default is the first SCSI disk (/dev/sda), but can be changed for other disk types.
* environment.IMG_URL: The operating system tarball (ubuntu or other) to stream to the machine you are configuring.
* image: Container image needed to perform the steps needed by this action.
* timeout: NEED INFO (Set to 360 (six minutes???). Is the timout to get the image???)

### template.tasks.actions.name.write-netplan
The `write-netplan` action writes network configuration information to the machine. It identifies:

* environment.CONTENTS.network.version: NEED INFO. CNI version, cilium version??? 
* environment.CONTENTS.network.renderer: NEED INFO  What is "networkd"????
* environment.CONTENTS.network.ethernets: Network interface to external network (eno1, by default) and whether or not to use dhcp4 (true, by default).
* environment.DEST_DISK: Destination block storage device partition where the operating system is copied. By default, /dev/sda2 is used (sda1 is the UEFI partition). 
* environment.DEST_PATH: File where the networking configuration is written (/etc/netplan/config.yaml, by default).
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0755, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0644, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout:

### template.tasks.actions.add-tink-cloud-init-config
The `add-tink-cloud-init-config` action configures cloud-init features to further configure the operating system. It identifies:

* environment.CONTENTS.datasource: NEED INFO. Ec2.metadata_urls: [] Ec2.strict_id: false
* environment.CONTENTS.system_info: Creates the `tink` user and gives it administrative group privileges (wheel, adm) and passwordless sudo privileges, and set the default shell (/bin/bash).
* environment.CONTENTS.manage_etc_hosts NEED INFO.  Set to localhost.
* environment.CONTENTS.warnings: NEED INFO. Sets dsid_missing_source: off
* environment.DEST_DISK: Destination block storage device partition where the operating system is located (/dev/sda2, by default).
* environment.DEST_PATH: Location of the cloud-init configuration file on disk (/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg, by default)
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0700, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0600, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout: NEED INFO. Set to 90, by default

### template.tasks.actions.add-tink-cloud-init-ds-config
The `add-tink-cloud-init-ds-config` action configures cloud-init data store features. This identifies the location of your metadata source once the machine is up and running. It identifies:

* environment.CONTENTS.datasource: NEED INFO. Set to Ec2.
* environment.DEST_DISK: Destination block storage device partition where the operating system is located (/dev/sda2, by default).
* environment.DEST_PATH: Location of the data store identity configuration file on disk (/etc/cloud/ds-identify.cfg, by default) 
* environment.DIRMODE: Linux directory permissions bits to use when creating directories (0700, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* environment.GID: The Linux group ID to set on file. Set to 0 (root group) by default.
* environment.MODE: The Linux permission bits to set on file (0600, by default).
* environment.UID: The Linux user ID to set on file. Set to 0 (root user) by default.
* image: Container image used to perform the steps needed by this action.
* timeout: NEED INFO. Set to 90, by default

### template.tasks.actions.kexec-image
The `kexec-image` action performs provisioning activities on the machine, then allows kexec to pivot the kernel to use the system installed on disk. This action identifies:

* environment.BLOCK_DEVICE: Disk partition on which the operating system is installed (/dev/sda2, by default)
* environment.FS_TYPE: Type of filesystem on the partition (ext4, by default).
* image: Container image used to perform the steps needed by this action.
* pid: NEED INFO. Set to host.
* timeout: NEED INFO. Set to 90.
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

* version
     "0.1"
