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
* [Machine Health Checks]({{< relref "../optional/healthchecks.md" >}})
* [API Server Extra Args]({{< relref "../optional/api-server-extra-args.md" >}})

To generate your own cluster configuration, follow instructions from the [Create Bare Metal cluster]({{< relref "./baremetal-getstarted" >}}) section and modify it using descriptions below.
For information on how to add cluster configuration settings to this file for advanced node configuration, see [Advanced Bare Metal cluster configuration]({{< relref "#advanced-bare-metal-cluster-configuration" >}}).

>**_NOTE_**: Bare Metal cluster creation with RHEL 9 raw OS images requires advanced cluster configurations to be set. To create Bare Metal RHEL 9 clusters, modify the cluster configurations using descriptions below and follow [Advanced Bare Metal cluster configuration]({{< relref "#advanced-bare-metal-cluster-configuration" >}}).
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
  kubernetesVersion: "1.31"
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

### controlPlaneConfiguration.taints (optional)
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint (For k8s versions prior to 1.24, `node-role.kubernetes.io/master`. For k8s versions 1.24+, `node-role.kubernetes.io/control-plane`). The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
>

### controlPlaneConfiguration.labels (optional)
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

#### controlPlaneConfiguration.upgradeRolloutStrategy (optional)
Configuration parameters for upgrade strategy.

#### controlPlaneConfiguration.upgradeRolloutStrategy.type (optional)
Default: `RollingUpdate`

Type of rollout strategy. Supported values: `RollingUpdate`,`InPlace`.

>**_NOTE:_** The upgrade rollout strategy type must be the same for all control plane and worker nodes.

#### controlPlaneConfiguration.upgradeRolloutStrategy.rollingUpdate (optional)
Configuration parameters for customizing rolling upgrade behavior.

>**_NOTE:_** The rolling update parameters can only be configured if `upgradeRolloutStrategy.type` is `RollingUpdate`.

#### controlPlaneConfiguration.upgradeRolloutStrategy.rollingUpdate.maxSurge (optional)
Default: 1

This can not be 0 if maxUnavailable is 0.

The maximum number of machines that can be scheduled above the desired number of machines.

Example: When this is set to n, the new worker node group can be scaled up immediately by n when the rolling upgrade starts. Total number of machines in the cluster (old + new) never exceeds (desired number of machines + n). Once scale down happens and old machines are brought down, the new worker node group can be scaled up further ensuring that the total number of machines running at any time does not exceed the desired number of machines + n.

### controlPlaneConfiguration.skipLoadBalancerDeployment (optional)
Optional field to skip deploying the control plane load balancer. Make sure your infrastructure can handle control plane load balancing when you set this field to true. In most cases, you should not set this field to true.

### datacenterRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration. See `TinkerbellDatacenterConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.31`, `1.30`, `1.29`, `1.28`, `1.27`

### managementCluster (required)
Identifies the name of the management cluster.
If your cluster spec is for a standalone or management cluster, this value is the same as the cluster name.

### workerNodeGroupConfigurations (optional)
This takes in a list of node groups that you can define for your workers.

You can omit `workerNodeGroupConfigurations` when creating Bare Metal clusters. If you omit `workerNodeGroupConfigurations`, control plane nodes will not be tainted and all pods will run on the control plane nodes. This mechanism can be used to deploy Bare Metal clusters on a single server. You can also run multi-node Bare Metal clusters without `workerNodeGroupConfigurations`.

>**_NOTE:_** Empty `workerNodeGroupConfigurations` is not supported when Kubernetes version <= 1.21.

### workerNodeGroupConfigurations[*].count (optional)
Number of worker nodes. (default: `1`) It will be ignored if the [cluster autoscaler curated package]({{< relref "../../packages/cluster-autoscaler/addclauto" >}}) is installed and `autoscalingConfiguration` is used to specify the desired range of replicas.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations[*].machineGroupRef (required)
Refers to the Kubernetes object with Tinkerbell-specific configuration for your nodes. See `TinkerbellMachineConfig Fields` below.

### workerNodeGroupConfigurations[*].name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations[*].autoscalingConfiguration (optional)
Configuration parameters for Cluster Autoscaler.

>**_NOTE:_** Autoscaling configuration is not supported when using the `InPlace` upgrade rollout strategy.

### workerNodeGroupConfigurations[*].autoscalingConfiguration.minCount (optional)
Minimum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations[*].autoscalingConfiguration.maxCount (optional)
Maximum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations[*].taints (optional)
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations[*].labels (optional)
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

### workerNodeGroupConfigurations[*].kubernetesVersion (optional)
The Kubernetes version you want to use for this worker node group. [Supported values]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}): `1.31`, `1.30`, `1.29`, `1.28`, `1.27`

Must be less than or equal to the cluster `kubernetesVersion` defined at the root level of the cluster spec. The worker node kubernetesVersion must be no more than two minor Kubernetes versions lower than the cluster control plane's Kubernetes version. Removing `workerNodeGroupConfiguration.kubernetesVersion` will trigger an upgrade of the node group to the `kubernetesVersion` defined at the root level of the cluster spec.

#### workerNodeGroupConfigurations[*].upgradeRolloutStrategy (optional)
Configuration parameters for upgrade strategy.

#### workerNodeGroupConfigurations[*].upgradeRolloutStrategy.type (optional)
Default: `RollingUpdate`

Type of rollout strategy. Supported values: `RollingUpdate`,`InPlace`.

>**_NOTE:_** The upgrade rollout strategy type must be the same for all control plane and worker nodes.

#### workerNodeGroupConfigurations[*].upgradeRolloutStrategy.rollingUpdate (optional)
Configuration parameters for customizing rolling upgrade behavior.

>**_NOTE:_** The rolling update parameters can only be configured if `upgradeRolloutStrategy.type` is `RollingUpdate`.

#### workerNodeGroupConfigurations[*].upgradeRolloutStrategy.rollingUpdate.maxSurge (optional)
Default: 1

This can not be 0 if maxUnavailable is 0.

The maximum number of machines that can be scheduled above the desired number of machines.

Example: When this is set to n, the new worker node group can be scaled up immediately by n when the rolling upgrade starts. Total number of machines in the cluster (old + new) never exceeds (desired number of machines + n). Once scale down happens and old machines are brought down, the new worker node group can be scaled up further ensuring that the total number of machines running at any time does not exceed the desired number of machines + n.

#### workerNodeGroupConfigurations[*].upgradeRolloutStrategy.rollingUpdate.maxUnavailable (optional)
Default: 0

This can not be 0 if MaxSurge is 0.

The maximum number of machines that can be unavailable during the upgrade.

Example: When this is set to n, the old worker node group can be scaled down by n machines immediately when the rolling upgrade starts. Once new machines are ready, old worker node group can be scaled down further, followed by scaling up the new worker node group, ensuring that the total number of machines unavailable at all times during the upgrade never falls below n.

## TinkerbellDatacenterConfig Fields

### tinkerbellIP (required)
Required field to identify the IP address of the Tinkerbell service.
This IP address must be a unique IP in the network range that does not conflict with other IPs.
Once the Tinkerbell services move from the Admin machine to run on the target cluster, this IP address makes it possible for the stack to be used for future provisioning needs.
When separate management and workload clusters are supported in Bare Metal, the IP address becomes a necessity.

### osImageURL (required)
Required field to set the operating system. In order to use Ubuntu or RHEL see [building baremetal node images]({{< relref "../../osmgmt/artifacts/#build-bare-metal-node-images" >}}). This field is also useful if you want to provide a customized operating system image or simply host the standard image locally. To upgrade a node or group of nodes to a new operating system version (ie. RHEL 8.7 to RHEL 8.8), modify this field to point to the new operating system image URL and run [upgrade cluster command]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/#upgrade-cluster-command" >}}).
The `osImageURL` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.31, the `osImageURL` name should include 1.31, 1_31, 1-31 or 131.

>**_NOTE:_** osImageURL field cannot be set both in the `TinkerbellDatacenterConfig` and `TinkerbellMachineConfig` objects. If this value is set for `TinkerbellDatacenterConfig`, osImageURL has to be set to empty string `""` for all the `TinkerbellMachineConfigs`.

### hookImagesURLPath (optional)
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

### skipLoadBalancerDeployment (optional)
Optional field to skip deploying the default load balancer for Tinkerbell stack.

EKS Anywhere for Bare Metal uses `kube-vip` load balancer by default to expose the Tinkerbell stack externally.
You can disable this feature by setting this field to `true`.
>**_NOTE:_** If you skip load balancer deployment, you will have to ensure that the Tinkerbell stack is available at [tinkerbellIP]({{< relref "#tinkerbellip-required" >}}) once the cluster creation is finished. One way to achieve this is by using the [MetalLB]({{< relref "../../packages/metallb" >}}) package.

### loadBalancerInterface (optional)
Optional field to configure a custom load balancer interface for Tinkerbell stack.

## TinkerbellMachineConfig Fields
In the example, there are `TinkerbellMachineConfig` sections for control plane (`my-cluster-name-cp`) and worker (`my-cluster-name`) machine groups.
The following fields identify information needed to configure the nodes in each of those groups.
>**_NOTE:_** Currently, you can only have one machine group for all machines in the control plane, although you can have multiple machine groups for the workers.
>
### hardwareSelector (required)
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

### osImageURL (required)
Required field to set the operating system. In order to use Ubuntu or RHEL see [building baremetal node images]({{< relref "../../osmgmt/artifacts/#build-bare-metal-node-images" >}}). This field is also useful if you want to provide a customized operating system image or simply host the standard image locally. To upgrade a node or group of nodes to a new operating system version (ie. RHEL 8.7 to RHEL 8.8), modify this field to point to the new operating system image URL and run [upgrade cluster command]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/#upgrade-cluster-command" >}}). The `osImageURL` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.31, the `osImageURL` name should include 1.31, 1_31, 1-31 or 131.

>**_NOTE:_** If this value is set for a single `TinkerbellMachineConfig`, osImageURL has to be set for all the `TinkerbellMachineConfigs`. osImageURL field cannot be set both in the `TinkerbellDatacenterConfig` and `TinkerbellMachineConfig` objects. If set for `TinkerbellMachineConfig`, the value must be set to empty string `""` for `TinkerbellDatacenterConfig`

### templateRef (optional)
Identifies the template that defines the actions that will be applied to the TinkerbellMachineConfig.
See TinkerbellTemplateConfig fields below.
EKS Anywhere will generate default templates based on `osFamily` during the `create` command.
You can override this default template by providing your own template here.

### users (optional)
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

When you generate a Bare Metal cluster configuration, by default, the `TinkerbellTemplateConfig` is not shown in the generated configuration file. The `TinkerbellTemplateConfig` defines the actions to provision each node, such as getting installation media, configure networking, add users, and otherwise configure the node. Internally, EKS Anywhere generates a default `TinkerbellTemplateConfig` based on the operating system family you choose. The default `TinkerbellTemplateConfig` is sufficient for most use cases.

If your use case necessitates that the operating system have additional configuration, you can add a `TinkerbellTemplateConfig` to your cluster configuration file with your own customizations. To do this, start with the default `TinkerbellTemplateConfig` generated by EKS Anywhere and modify it as needed. To generate the default `TinkerbellTemplateConfig`, use the following command:

```bash
eksctl anywhere generate tinkerbelltemplateconfig -f eksa-mgmt-cluster.yaml
```

Now you can add your own Actions for configuring nodes. We highly recommend that you do not modify the first and the last Actions in the default `TinkerbellTemplateConfig`. The first Action streams the OS image to the disk, and the last Action reboots the node. See the upstream Tinkerbell documentation for more information on [Templates](https://tinkerbell.org/docs/concepts/templates/) and [Actions](https://tinkerbell.org/docs/concepts/templates/#actions).

The following shows the default `TinkerbellTemplateConfig` generated by `eksctl anywhere generate tinkerbelltemplateconfig`.

### Ubuntu

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
          DEST_DISK: '{{ index .Hardware.Disks 0 }}'
          IMG_URL: https://my-file-server/ubuntu-2204-kube-v1.30.gz
        image: 127.0.0.1/embedded/image2disk
        name: stream image to disk
        timeout: 600
      - environment:
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}'
          DEST_PATH: /etc/netplan/config.yaml
          DIRMODE: "0755"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          STATIC_NETPLAN: "true"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: write netplan config
        pid: host
        timeout: 90
      - environment:
          CONTENTS: 'network: {config: disabled}'
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}'
          DEST_PATH: /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: disable cloud-init network capabilities
        timeout: 90
      - environment:
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: [http://<admin-machine-ip>:50061,http://<tinkerbellIP-from-cluster-config>:50061]
                strict_id: false
            manage_etc_hosts: localhost
            warnings:
              dsid_missing_source: off
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}'
          DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: add cloud-init config
        timeout: 90
      - environment:
          CONTENTS: |
            datasource: Ec2
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}'
          DEST_PATH: /etc/cloud/ds-identify.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: add cloud-init ds config
        timeout: 90
      - image: 127.0.0.1/embedded/reboot
        name: reboot
        pid: host
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

### Redhat

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
          DEST_DISK: '{{ index .Hardware.Disks 0 }}'
          IMG_URL: https://my-file-server/rhel-9-kube-v1.30.0.gz
        image: 127.0.0.1/embedded/image2disk
        name: stream image to disk
        timeout: 600
      - environment:
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}'
          DEST_PATH: /etc/netplan/config.yaml
          DIRMODE: "0755"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          STATIC_NETPLAN: "true"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: write netplan config
        pid: host
        timeout: 90
      - environment:
          CONTENTS: 'network: {config: disabled}'
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}'
          DEST_PATH: /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: disable cloud-init network capabilities
        timeout: 90
      - environment:
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: ['http://<admin-machine-ip>:50061','http://<tinkerbellIP-from-cluster-config>:50061']
                strict_id: false
            manage_etc_hosts: localhost
            warnings:
              dsid_missing_source: off
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}'
          DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: add cloud-init config
        timeout: 90
      - environment:
          CONTENTS: |
            datasource: Ec2
          DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}'
          DEST_PATH: /etc/cloud/ds-identify.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: 127.0.0.1/embedded/writefile
        name: add cloud-init ds config
        timeout: 90
      - image: 127.0.0.1/embedded/reboot
        name: reboot
        pid: host
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

## TinkerbellTemplateConfig fields

Each Action has its own set of environment variables that are used to configure the Action. EKS Anywhere embeds a set of images into HookOS. This means that when these images are used in an Action they will not be pulled from any external registry. No external network communication is needed to use them. These Actions are prefixed with `127.0.0.1/embedded`. 

>**_NOTE:_**
Actions are user-defined and can be any container image that is available to HookOS. You can create your own Action images if the embedded Action images do not meet your requirements.
>

The following describes the embedded Action images that are available in HookOS and their configuration options.

### cexec

`127.0.0.1/embedded/cexec`

The `cexec` Action performs execution either within a chroot environment or within the HookOS filesystem. The primary use-case is being able to provision files/an Operating System to disk and then being able to execute something that resides within that filesystem.

All options can be set either via environment variables or CLI flags.
CLI flags take precedence over environment variables, which take precedence over default values.

| Env variable | Flag | Type | Default Value | Required | Description |
|--------------|------|------|---------------|----------|-------------|
| `BLOCK_DEVICE` | `--block-device` | string | "" | yes | The block device to mount. |
| `FS_TYPE` | `--fs-type` | string | "" | yes | The filesystem type of the block device. |
| `CHROOT` | `--chroot` | string | "" | no | If set to `y` (or a non empty string), the Action will execute the given command within a chroot environment. This option is DEPRECATED. Future versions will always chroot. |
| `CMD_LINE` | `--cmd-line` | string | "" | yes | The command to execute. |
| `DEFAULT_INTERPRETER` | `--default-interpreter` | string | "" | no | The default interpreter to use when executing commands. This is useful when you need to execute multiple commands. |
| `UPDATE_RESOLV_CONF` | `--update-resolv-conf` | boolean | false | no | If set to `true`, the cexec Action will update the `/etc/resolv.conf` file within the chroot environment with the `/etc/resolv.conf` from the host. |
| `JSON_OUTPUT` | `--json-output` | boolean | true | no | If set to `true`, the cexec Action will log output in JSON format. The defaults to `true`. If set to `false`, the cexec Action will log output in plain text format. |

Any environment variables you set on the Action will be available to the command you execute.
For example, if you set `DEBIAN_FRONTEND: noninteractive` as an environment variable, it will be available to the command you execute.

### writefile

`127.0.0.1/embedded/writefile`

The `writefile` Action will mount a block device and write a file to a destination path on its filesystem.

| Env variable | Type | Required | Description |
|--------------|------|----------|-------------|
| `DEST_DISK` | string | yes | The block device to mount. |
| `FS_TYPE` | string | yes | The filesystem type of the block device. |
| `DEST_PATH` | string | yes | The path to write the file to. |
| `CONTENTS` | string | yes | The contents of the file to write. |
| `UID` | string | yes | The user ID to set on the file. |
| `GID` | string | yes | The group ID to set on the file. |
| `MODE` | string | yes | The permission bits to set on the file. |
| `DIRMODE` | string | yes | The permission bits to set on the directory. |

The follow field must be set on the Action.

```yaml
pid: host
```

### image2disk

`127.0.0.1/embedded/image2disk`

The `image2disk` Action will stream either a compressed or not compressed remote image (raw) to a block device.

| Env variable | Type | Default Value | Required | Description |
|--------------|------|---------------|----------|-------------|
| IMG_URL | string | "" | yes | URL of the image to be streamed |
| DEST_DISK | string | "" | yes | Block device to which to write the image |
| COMPRESSED | bool | false | no | Decompress the image before writing it to the disk |
| RETRY_ENABLED | bool | false | no | Retry the Action, using exponential backoff, for the duration specified in `RETRY_DURATION_MINUTES` before failing |
| RETRY_DURATION_MINUTES | int | 10 | no | Duration for which the Action will retry before failing |
| PROGRESS_INTERVAL_SECONDS | int | 3 | no | Interval at which the progress of the image transfer will be logged |
| TEXT_LOGGING | bool | false | no | Output from the Action will be logged in a more human friendly text format, JSON format is used by default |

### oci2disk

`127.0.0.1/embedded/oci2disk`

The `oci2disk` Action provides the capability of streaming a raw (compressed) disk image from an OCI compliant registry to a local block device.

| Env variable | Type | Default Value | Required | Description |
|--------------|------|---------------|----------|-------------|
| IMG_URL | string | "" | yes | URL of the image to be streamed |
| DEST_DISK | string | "" | yes | Block device to which to write the image |
| COMPRESSED | bool | false | no | Decompress the image before writing it to the disk |

### kexec

 `127.0.0.1/embedded/kexec`

The `kexec` Action makes use of the [Linux kexec function](https://en.wikipedia.org/wiki/Kexec) to boot directly into a kernel.

BLOCK_DEVICE: /dev/sda3
      FS_TYPE: ext4
      KERNEL_PATH: /boot/vmlinuz
      INITRD_PATH: /boot/initrd
      CMD_LINE: "root=/dev/sda3 ro"

| Env variable | Type | Default Value | Required | Description |
|--------------|------|---------------|----------|-------------|
| BLOCK_DEVICE | string | "" | yes | The block device to mount. |
| FS_TYPE | string | "" | yes | The filesystem type of the block device. |
| KERNEL_PATH | string | "" | no | The path to the kernel to boot. When not supplied the kernel path will discovered from `/boot/grub/grub.cfg` on the `BLOCK_DEVICE`. |
| INITRD_PATH | string | "" | yes | The path to the initial ramdisk to boot. |
| CMD_LINE | string | "" | no | The command line to pass to the kernel. When not supplied the kernel cmdline parameters will discovered from `/boot/grub/grub.cfg` on the `BLOCK_DEVICE`. |

>**_NOTE:_**
There are known issues related to drivers with some hardware that may make it necessary to replace the `kexec` Action image with the `reboot` Action image.
>

### reboot

`127.0.0.1/embedded/reboot`

The `reboot` Action will reboot the machine.

The follow fields must be set on the Action.

```yaml
pid: host
volumes:
  - /worker:/worker
```

## Custom Tinkerbell action examples

The following example shows how to add a .deb package (`openssl`) to a Redhat installation:

```yaml
- environment:
    BLOCK_DEVICE: '{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}' # /dev/sda1
    CHROOT: "y"
    CMD_LINE: apt -y update && apt -y install openssl
    DEFAULT_INTERPRETER: /bin/sh -c
    FS_TYPE: ext4
  image: 127.0.0.1/embedded/cexec
  name: install openssl
  timeout: 90
```

The following shows an example of adding a new user (`tinkerbell`) to an Ubuntu installation:

```yaml
- environment:
    BLOCK_DEVICE: '{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}' # /dev/sda2
    FS_TYPE: ext4
    CHROOT: y
    DEFAULT_INTERPRETER: "/bin/sh -c"
    CMD_LINE: "useradd --password $(openssl passwd -1 tinkerbell) --shell /bin/bash --create-home --groups sudo tinkerbell"
  image: 127.0.0.1/embedded/cexec
  name: create a user
  timeout: 90
```

Look for more examples as they are added to the [Tinkerbell examples](https://github.com/aws/eks-anywhere/tree/main/examples/tinkerbell) page.
