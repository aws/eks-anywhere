---
title: "Configure for Bare Metal"
linkTitle: "Configuration"
weight: 50
aliases:
    /docs/reference/clusterspec/baremetal/
description: >
  Full EKS Anywhere configuration reference for a Bare Metal cluster.
---

This is a generic template with detailed descriptions below for reference.
The following additional optional configuration can also be included:

* [CNI]({{< relref "../optional/cni.md" >}})
* [Host OS Config]({{< relref "../optional/hostOSConfig.md" >}})
* [Proxy]({{< relref "../optional/proxy.md" >}})
* [Gitops]({{< relref "../optional/gitops.md" >}})
* [IAM Authenticator]({{< relref "../optional/iamauth.md" >}})
* [OIDC]({{< relref "../optional/oidc.md" >}})
* [Registry Mirror]({{< relref "../optional/registrymirror.md" >}})
* [Machine Health Check Timeouts]({{< relref "../optional/healthchecks.md" >}})

To generate your own cluster configuration, follow instructions from the [Create Bare Metal cluster]({{< relref "./baremetal-getstarted" >}}) section and modify it using descriptions below.
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
      host: "<Control Plane Endpoint IP>"
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: my-cluster-name-cp
  datacenterRef:
    kind: TinkerbellDatacenterConfig
    name: my-cluster-name
  kubernetesVersion: "1.28"
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
  tinkerbellIP: "<Tinkerbell IP>"

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  hardwareSelector: {}
  osFamily: bottlerocket
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
  osFamily: bottlerocket
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

{{% include "../_configuration/cluster_clusterNetwork.html" %}}

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes.
This number needs to be odd to maintain ETCD quorum.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other machines.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing.

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration for your nodes. See `TinkerbellMachineConfig Fields` below.

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint (For k8s versions prior to 1.24, `node-role.kubernetes.io/master`. For k8s versions 1.24+, `node-role.kubernetes.io/control-plane`). The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
>

### controlPlaneConfiguration.labels
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

### controlPlaneConfiguration.skipLoadBalancerDeployment
Optional field to skip deploying the control plane load balancer. Make sure your infrastructure can handle control plane load balancing when you set this field to true. In most cases, you should not set this field to true.

### datacenterRef
Refers to the Kubernetes object with Tinkerbell-specific configuration. See `TinkerbellDatacenterConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

### managementCluster
Identifies the name of the management cluster.
If your cluster spec is for a standalone or management cluster, this value is the same as the cluster name.

### workerNodeGroupConfigurations
This takes in a list of node groups that you can define for your workers.

You can omit `workerNodeGroupConfigurations` when creating Bare Metal clusters. If you omit `workerNodeGroupConfigurations`, control plane nodes will not be tainted and all pods will run on the control plane nodes. This mechanism can be used to deploy Bare Metal clusters on a single server. You can also run multi-node Bare Metal clusters without `workerNodeGroupConfigurations`.

>**_NOTE:_** Empty `workerNodeGroupConfigurations` is not supported when Kubernetes version <= 1.21.

### workerNodeGroupConfigurations.count
Number of worker nodes. Optional if autoscalingConfiguration is used, in which case count will default to `autoscalingConfiguration.minCount`.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration for your nodes. See `TinkerbellMachineConfig Fields` below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations.autoscalingConfiguration.minCount
Minimum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.autoscalingConfiguration.maxCount
Maximum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

### workerNodeGroupConfigurations.kubernetesVersion
The Kubernetes version you want to use for this worker node group. [Supported values]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}): `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

Must be less than or equal to the cluster `kubernetesVersion` defined at the root level of the cluster spec. The worker node kubernetesVersion must be no more than two minor Kubernetes versions lower than the cluster control plane's Kubernetes version. Removing `workerNodeGroupConfiguration.kubernetesVersion` will trigger an upgrade of the node group to the `kubernetesVersion` defined at the root level of the cluster spec.

## TinkerbellDatacenterConfig Fields

### tinkerbellIP
Required field to identify the IP address of the Tinkerbell service.
This IP address must be a unique IP in the network range that does not conflict with other IPs.
Once the Tinkerbell services move from the Admin machine to run on the target cluster, this IP address makes it possible for the stack to be used for future provisioning needs.
When separate management and workload clusters are supported in Bare Metal, the IP address becomes a necessity.

### osImageURL
Optional field to replace the default Bottlerocket operating system. EKS Anywhere can only auto-import Bottlerocket. In order to use Ubuntu or RHEL see [building baremetal node images]({{< relref "../../osmgmt/artifacts/#build-bare-metal-node-images" >}}). This field is also useful if you want to provide a customized operating system image or simply host the standard image locally. To upgrade a node or group of nodes to a new operating system version (ie. RHEL 8.7 to RHEL 8.8), modify this field to point to the new operating system image URL and run [upgrade cluster command]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/#upgrade-cluster-command" >}}).
The `osImageURL` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.24, the `osImageURL` name should include 1.24, 1_24, 1-24 or 124.

### hookImagesURLPath
Optional field to replace the HookOS image.
This field is useful if you want to provide a customized HookOS image or simply host the standard image locally.
See [Artifacts]({{< relref "../../osmgmt/artifacts/#hookos-kernel-and-initial-ramdisk-for-bare-metal" >}}) for details.

#### Example `TinkerbellDatacenterConfig.spec`
```yaml
spec:
  tinkerbellIP: "192.168.0.10"                                          # Available, routable IP
  osImageURL: "http://my-web-server/ubuntu-v1.23.7-eks-a-12-amd64.gz"   # Full URL to the OS Image hosted locally
  hookImagesURLPath: "http://my-web-server/hook"                        # Path to the hook images. This path must contain vmlinuz-x86_64 and initramfs-x86_64
```
This is the folder structure for `my-web-server`:
```
my-web-server
├── hook
│   ├── initramfs-x86_64
│   └── vmlinuz-x86_64
└── ubuntu-v1.23.7-eks-a-12-amd64.gz
```

### skipLoadBalancerDeployment
Optional field to skip deploying the default load balancer for Tinkerbell stack.

EKS Anywhere for Bare Metal uses `kube-vip` load balancer by default to expose the Tinkerbell stack externally.
You can disable this feature by setting this field to `true`.
>**_NOTE:_** If you skip load balancer deployment, you will have to ensure that the Tinkerbell stack is available at [tinkerbellIP]({{< relref "#tinkerbellip" >}}) once the cluster creation is finished. One way to achieve this is by using the [MetalLB]({{< relref "../../packages/metallb" >}}) package.

## TinkerbellMachineConfig Fields
In the example, there are `TinkerbellMachineConfig` sections for control plane (`my-cluster-name-cp`) and worker (`my-cluster-name`) machine groups.
The following fields identify information needed to configure the nodes in each of those groups.
>**_NOTE:_** Currently, you can only have one machine group for all machines in the control plane, although you can have multiple machine groups for the workers.
>
### hardwareSelector
Use fields under `hardwareSelector` to add key/value pair labels to match particular machines that you identified in the CSV file where you defined the machines in your cluster.
Choose any label name you like.
For example, if you had added the label `node=cp-machine` to the machines listed in your CSV file that you want to be control plane nodes, the following `hardwareSelector` field would cause those machines to be added to the control plane:
```yaml
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
Operating system on the machine. Permitted values: `bottlerocket`, `ubuntu`, `redhat` (Default: `bottlerocket`).

### osImageURL (optional)
Optional field to replace the default Bottlerocket operating system. EKS Anywhere can only auto-import Bottlerocket. In order to use Ubuntu or RHEL see [building baremetal node images]({{< relref "../../osmgmt/artifacts/#build-bare-metal-node-images" >}}). This field is also useful if you want to provide a customized operating system image or simply host the standard image locally. To upgrade a node or group of nodes to a new operating system version (ie. RHEL 8.7 to RHEL 8.8), modify this field to point to the new operating system image URL and run [upgrade cluster command]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/#upgrade-cluster-command" >}}).

>**_NOTE:_** If specified for a single `TinkerbellMachineConfig`, osImageURL has to be specified for all the `TinkerbellMachineConfigs`.
osImageURL field cannot be specified both in the `TinkerbellDatacenterConfig` and `TinkerbellMachineConfig` objects.

### templateRef (optional)
Identifies the template that defines the actions that will be applied to the TinkerbellMachineConfig.
See TinkerbellTemplateConfig fields below.
EKS Anywhere will generate default templates based on `osFamily` during the `create` command.
You can override this default template by providing your own template here.

### users
The name of the user you want to configure to access your virtual machines through SSH.

The default is `ec2-user`.
Currently, only one user is supported.

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your machines through SSH (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster machines so you can SSH into
them. The user will be what is defined under `name` above. For example:

```
ssh -i <private-key-file> <user>@<machine-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value.

### hostOSConfig (optional)
Optional host OS configurations for the EKS Anywhere Kubernetes nodes.
More information in the [Host OS Configuration]({{< relref "../optional/hostOSConfig.md" >}}) section.

## Advanced Bare Metal cluster configuration

When you generate a Bare Metal cluster configuration, the `TinkerbellTemplateConfig` is kept internally and not shown in the generated configuration file.
`TinkerbellTemplateConfig` settings define the actions done to install each node, such as get installation media, configure networking, add users, and otherwise configure the node.

Advanced users can override the default values set for `TinkerbellTemplateConfig`.
They can also add their own [Tinkerbell actions](https://docs.tinkerbell.org/actions/action-architecture/) to make personalized modifications to EKS Anywhere nodes.

The following shows two `TinkerbellTemplateConfig` examples that you can add to your cluster configuration file to override the values that EKS Anywhere sets: one for Ubuntu and one for Bottlerocket.
Most actions used differ for different operating systems.

>**_NOTE:_** For the `stream-image` action, `DEST_DISK` points to the device representing the entire hard disk (for example, `/dev/sda`).
For UEFI-enabled images, such as Ubuntu, write actions use `DEST_DISK` to point to the second partition (for example, `/dev/sda2`), with the first being the EFI partition.
For the Bottlerocket image, which has 12 partitions, `DEST_DISK` is partition 12 (for example, `/dev/sda12`).
Device names will be different for different disk types.
>

### Ubuntu TinkerbellTemplateConfig example

```yaml
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
          IMG_URL: https://my-file-server/ubuntu-v1.23.7-eks-a-12-amd64.gz
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/image2disk:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: stream-image
        timeout: 360
      - environment:
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/netplan/config.yaml
          STATIC_NETPLAN: true
          DIRMODE: "0755"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-netplan
        timeout: 90
      - environment:
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: [<admin-machine-ip>, <tinkerbell-ip-from-cluster-config>]
                strict_id: false
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
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: add-tink-cloud-init-config
        timeout: 90
      - environment:
          CONTENTS: |
            network:
              config: disabled
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: disable-cloud-init-network-capabilities
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
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: add-tink-cloud-init-ds-config
        timeout: 90
      - environment:
          BLOCK_DEVICE: /dev/sda2
          FS_TYPE: ext4
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/kexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
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

Pay special attention to the `BOOTCONFIG_CONTENTS` environment section below if you wish to set up console redirection for the kernel and systemd.
If you are only using a direct attached monitor as your primary display device, no additional configuration is needed here.
However, if you need all boot output to be shown via a server’s serial console for example, extra configuration should be provided inside `BOOTCONFIG_CONTENTS`.

An empty `kernel {}` key is provided below in the example; inside this key is where you will specify your console devices.
You may specify multiple comma delimited console devices in quotes to a console key as such: `console = "tty0", "ttyS0,115200n8"`.
The order of the devices is significant; systemd will output to the last device specified.
The console key belongs inside the kernel key like so:
```
kernel {
    console = "tty0", "ttyS0,115200n8"
}
```

The above example will send all kernel output to both consoles, and systemd output to `ttyS0`.
Additional information about serial console setup can be found in the [Linux kernel documentation](https://www.kernel.org/doc/html/latest/admin-guide/serial-console.html).

```yaml
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
          IMG_URL: https://anywhere-assets.eks.amazonaws.com/releases/bundles/11/artifacts/raw/1-22/bottlerocket-v1.22.10-eks-d-1-22-8-eks-a-11-amd64.img.gz
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/image2disk:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: stream-image
        timeout: 360
      - environment:
          # An example console declaration that will send all kernel output to both consoles, and systemd output to ttyS0.
          # kernel {
          #     console = "tty0", "ttyS0,115200n8"
          # }
          BOOTCONFIG_CONTENTS: |
            kernel {}
          DEST_DISK: /dev/sda12
          DEST_PATH: /bootconfig.data
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
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
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-netconfig
        timeout: 90
      - environment:
          HEGEL_URLS: http://<hegel-ip>:50061
          DEST_DISK: /dev/sda12
          DEST_PATH: /user-data.toml
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-user-data
        timeout: 90
      - name: "reboot"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/reboot:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        timeout: 90
        volumes:
          - /worker:/worker
      name: my-cluster-name
      volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
      worker: '{{.device_1}}'
    version: "0.1"
```
## TinkerbellTemplateConfig Fields

The values in the `TinkerbellTemplateConfig` fields are created from the contents of the CSV file used to generate a configuration.
The template contains actions that are performed on a Bare Metal machine when it first boots up to be provisioned.
For advanced users, you can add these fields to your cluster configuration file if you have special needs to do so.

While there are fields that apply to all provisioned operating systems, actions are specific to each operating system.
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
* environment.DEST_DISK: The hard disk on which the operating system is deployed. The default is the first SCSI disk (/dev/sda), but can be changed for other disk types.
* environment.IMG_URL: The operating system tarball (ubuntu or other) to stream to the machine you are configuring.
* image: Container image needed to perform the steps needed by this action.
* timeout: Sets the amount of time (in seconds) that Tinkerbell has to stream the image, uncompress it, and write it to disk before timing out. Consider increasing this limit from the default 600 to a higher limit if this action is timing out.

## Ubuntu-specific actions

### template.tasks.actions.name.write-netplan (Ubuntu)
The `write-netplan` action writes Ubuntu network configuration information to the machine (see [Netplan](https://netplan.io/)) for details. It identifies:

* environment.CONTENTS.network.version: Identifies the network version.
* environment.CONTENTS.network.renderer: Defines the service to manage networking. By default, the `networkd` systemd service is used.
* environment.CONTENTS.network.ethernets: Network interface to external network (eno1, by default) and whether or not to use dhcp4 (true, by default).
* environment.DEST_DISK: Destination block storage device partition where the operating system is copied. By default, /dev/sda2 is used (sda1 is the EFI partition).
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

* environment.CONTENTS.datasource: Identifies Ec2 (Ec2.metadata_urls) as the data source and sets `Ec2.strict_id: false` to prevent cloud-init from producing warnings about this datasource.
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

There are known issues related to drivers with some hardware that may make it necessary to replace the kexec-image action with a full reboot.
If you require a full reboot, you can change the kexec-image setting as follows:

```yaml
actions:
- name: "reboot"
  image: public.ecr.aws/l0g8r8j6/tinkerbell/hub/reboot-action:latest
  timeout: 90
  volumes:
  - /worker:/worker
```

## Bottlerocket-specific actions

### template.tasks.actions.write-bootconfig (Bottlerocket)
The write-bootconfig action identifies the location on the machine to put content needed to boot the system from disk.

* environment.BOOTCONFIG_CONTENTS.kernel: Add kernel parameters that are passed to the kernel when the system boots.
* environment.DEST_DISK: Identifies the block storage device that holds the boot partition.
* environment.DEST_PATH: Identifies the file holding boot configuration data (`/bootconfig.data` in this example).
* environment.DIRMODE: The Linux permissions assigned to the boot directory.
* environment.FS_TYPE: The filesystem type associated with the boot partition.
* environment.GID: The group ID associated with files and directories created on the boot partition.
* environment.MODE: The Linux permissions assigned to files in the boot partition.
* environment.UID: The user ID associated with files and directories created on the boot partition. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.write-netconfig (Bottlerocket)
The write-netconfig action configures networking for the system.

* environment.CONTENTS: Add network values, including: `version = 1` (version number), `[eno1]` (external network interface), `dhcp4 = true` (turns on dhcp4), and `primary = true` (identifies this interface as the primary interface used by kubelet).
* environment.DEST_DISK: Identifies the block storage device that holds the network configuration information.
* environment.DEST_PATH: Identifies the file holding network configuration data (`/net.toml` in this example).
* environment.DIRMODE: The Linux permissions assigned to the directory holding network configuration settings.
* environment.FS_TYPE: The filesystem type associated with the partition holding network configuration settings.
* environment.GID: The group ID associated with files and directories created on the partition. GID 0 is the root group.
* environment.MODE: The Linux permissions assigned to files in the partition.
* environment.UID: The user ID associated with files and directories created on the partition. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.

### template.tasks.actions.write-user-data (Bottlerocket)
The write-user-data action configures the Tinkerbell Hegel service, which provides the metadata store for Tinkerbell.

* environment.HEGEL_URLS: The IP address and port number of the Tinkerbell [Hegel](https://docs.tinkerbell.org/services/hegel/) service.
* environment.DEST_DISK: Identifies the block storage device that holds the network configuration information.
* environment.DEST_PATH: Identifies the file holding network configuration data (`/net.toml` in this example).
* environment.DIRMODE: The Linux permissions assigned to the directory holding network configuration settings.
* environment.FS_TYPE: The filesystem type associated with the partition holding network configuration settings.
* environment.GID: The group ID associated with files and directories created on the partition. GID 0 is the root group.
* environment.MODE: The Linux permissions assigned to files in the partition.
* environment.UID: The user ID associated with files and directories created on the partition. UID 0 is the root user.
* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.

### template.tasks.actions.reboot (Bottlerocket)
The reboot action defines how the system restarts to bring up the installed system.

* image: Container image used to perform the steps needed by this action.
* timeout: Time needed to complete the action, in seconds.
* volumes: The volume (directory) to mount into the container from the installed system.
### version

Matches the current version of the Tinkerbell template.

## Custom Tinkerbell action examples
By creating your own custom Tinkerbell actions, you can add to or modify the installed operating system so those changes take effect when the installed system first starts (from a reboot or pivot).
The following example shows how to add a .deb package (`openssl`) to an Ubuntu installation:

```yaml
      - environment:
          BLOCK_DEVICE: /dev/sda1
          CHROOT: "y"
          CMD_LINE: apt -y update && apt -y install openssl
          DEFAULT_INTERPRETER: /bin/sh -c
          FS_TYPE: ext4
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/cexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: install-openssl
        timeout: 90
```
The following shows an example of adding a new user (`tinkerbell`) to an installed Ubuntu system:

```yaml
      - environment:
          BLOCK_DEVICE: <block device path> # E.g. /dev/sda1
          FS_TYPE: ext4
          CHROOT: y
          DEFAULT_INTERPRETER: "/bin/sh -c"
          CMD_LINE: "useradd --password $(openssl passwd -1 tinkerbell) --shell /bin/bash --create-home --groups sudo tinkerbell"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/cexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: "create-user"
        timeout: 90
```

Look for more examples as they are added to the [Tinkerbell examples](https://github.com/aws/eks-anywhere/tree/main/examples/tinkerbell) page.
