---
title: "Configure for Nutanix"
linkTitle: "Configuration"
weight: 40
aliases:
    /docs/reference/clusterspec/nutanix/
description: >
  Full EKS Anywhere configuration reference for a Nutanix cluster
---

This is a generic template with detailed descriptions below for reference.

The following additional optional configuration can also be included:

* [CNI]({{< relref "../optional/cni.md" >}})
* [IAM Roles for Service Accounts]({{< relref "../optional/irsa.md" >}})
* [IAM Authenticator]({{< relref "../optional/iamauth.md" >}})
* [OIDC]({{< relref "../optional/oidc.md" >}})
* [Registry Mirror]({{< relref "../optional/registrymirror.md" >}})
* [Proxy]({{< relref "../optional/proxy.md" >}})
* [Gitops]({{< relref "../optional/gitops.md" >}})
* [Machine Health Check Timeouts]({{< relref "../optional/healthchecks.md" >}})

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
 name: mgmt
 namespace: default
spec:
 clusterNetwork:
   cniConfig:
     cilium: {}
   pods:
     cidrBlocks:
       - 192.168.0.0/16
   services:
     cidrBlocks:
       - 10.96.0.0/16
 controlPlaneConfiguration:
   count: 3
   endpoint:
     host: ""
   machineGroupRef:
     kind: NutanixMachineConfig
     name: mgmt-cp-machine
 datacenterRef:
   kind: NutanixDatacenterConfig
   name: nutanix-cluster
 kubernetesVersion: "1.28"
 workerNodeGroupConfigurations:
   - count: 1
     machineGroupRef:
       kind: NutanixMachineConfig
       name: mgmt-machine
     name: md-0
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixDatacenterConfig
metadata:
 name: nutanix-cluster
 namespace: default
spec:
 endpoint: pc01.cloud.internal
 port: 9440
 credentialRef:
   kind: Secret
   name: nutanix-credentials
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixMachineConfig
metadata:
 annotations:
   anywhere.eks.amazonaws.com/control-plane: "true"
 name: mgmt-cp-machine
 namespace: default
spec:
 cluster:
   name: nx-cluster-01
   type: name
 image:
   name: eksa-ubuntu-2004-kube-v1.28
   type: name
 memorySize: 4Gi
 osFamily: ubuntu
 subnet:
   name: vm-network
   type: name
 systemDiskSize: 40Gi
 project:
   type: name
   name: my-project
 users:
   - name: eksa
     sshAuthorizedKeys:
       - ssh-rsa AAAA…
 vcpuSockets: 2
 vcpusPerSocket: 1
 additionalCategories:
   - key: my-category
     value: my-category-value
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixMachineConfig
metadata:
 name: mgmt-machine
 namespace: default
spec:
 cluster:
   name: nx-cluster-01
   type: name
 image:
   name: eksa-ubuntu-2004-kube-v1.28
   type: name
 memorySize: 4Gi
 osFamily: ubuntu
 subnet:
   name: vm-network
   type: name
 systemDiskSize: 40Gi
 project:
   type: name
   name: my-project
 users:
   - name: eksa
     sshAuthorizedKeys:
       - ssh-rsa AAAA…
 vcpuSockets: 2
 vcpusPerSocket: 1
---
```

## Cluster Fields

### name (required)
Name of your cluster `mgmt` in this example.

{{% include "../_configuration/cluster_clusterNetwork.html" %}}

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with Nutanix specific configuration for your nodes. See `NutanixMachineConfig` fields below.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster
creation process are [here]({{< relref "./nutanix-prereq/#prepare-a-nutanix-environment" >}}).

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers. You may define one or more worker node groups.

### workerNodeGroupConfigurations.count
Number of worker nodes. Optional if `autoscalingConfiguration` is used, in which case count will default to `autoscalingConfiguration.minCount`.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with Nutanix specific configuration for your nodes. See `NutanixMachineConfig` fields below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: `md-0`)

### workerNodeGroupConfigurations.autoscalingConfiguration.minCount
Minimum number of nodes for this node group’s autoscaling configuration.

### workerNodeGroupConfigurations.autoscalingConfiguration.maxCount
Maximum number of nodes for this node group’s autoscaling configuration.

### workerNodeGroupConfigurations.kubernetesVersion
The Kubernetes version you want to use for this worker node group. Supported values: 1.28, 1.27, 1.26, 1.25, 1.24

### datacenterRef
Refers to the Kubernetes object with Nutanix environment specific configuration. See `NutanixDatacenterConfig` fields below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

## NutanixDatacenterConfig Fields

### endpoint (required)
The Prism Central server fully qualified domain name or IP address. If the server IP is used, the PC SSL certificate must have an IP SAN configured.

### port (required)
The Prism Central server port. (Default: `9440`)

### credentialRef (required)
Reference to the Kubernetes secret that contains the Prism Central credentials.

### insecure (optional)
Set insecure to `true` if the Prism Central server does not have a valid certificate. This is not recommended for production use cases. (Default: `false`)

### additionalTrustBundle (optional; required if using a self-signed PC SSL certificate)
The PEM encoded CA trust bundle.

The `additionalTrustBundle` needs to be populated with the PEM-encoded x509 certificate of the Root CA that issued the certificate for Prism Central. Suggestions on how to obtain this certificate are [here]({{< relref "./nutanix-prereq/#prepare-a-nutanix-environment" >}}).

__Example__:</br>
```yaml
 additionalTrustBundle: |
    -----BEGIN CERTIFICATE-----
    <certificate string>
    -----END CERTIFICATE-----
    -----BEGIN CERTIFICATE-----
    <certificate string>
    -----END CERTIFICATE-----
```

## NutanixMachineConfig Fields

### cluster
Reference to the Prism Element cluster.

### cluster.type
Type to identify the Prism Element cluster. (Permitted values: `name` or `uuid`)

### cluster.name
Name of the Prism Element cluster.

### cluster.uuid
UUID of the Prism Element cluster.

### image
Reference to the OS image used for the system disk.

### image.type
Type to identify the OS image. (Permitted values: `name` or `uuid`)

### image.name (`name` or `UUID` required)
Name of the image
The `image.name` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.24, `image.name` must include 1.24, 1_24, 1-24 or 124.

### image.uuid (`name` or `UUID` required)
UUID of the image
The name of the image associated with the `uuid` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.24, the name associated with `image.uuid` field must include 1.24, 1_24, 1-24 or 124.

### memorySize
Size of RAM on virtual machines (Default: `4Gi`)

### osFamily (optional)
Operating System on virtual machines. Permitted values: `ubuntu` and `redhat`. (Default: `ubuntu`)

### subnet
Reference to the subnet to be assigned to the VMs.

### subnet.name (`name` or `UUID` required)
Name of the subnet.

### subnet.type
Type to identify the subnet. (Permitted values: `name` or `uuid`)

### subnet.uuid (`name` or `UUID` required)
UUID of the subnet.

### systemDiskSize
Amount of storage assigned to the system disk. (Default: `40Gi`)

### vcpuSockets
Amount of vCPU sockets. (Default: `2`)

### vcpusPerSocket
Amount of vCPUs per socket. (Default: `1`)

### project	(optional)
Reference to an existing project used for the virtual machines.

### project.type
Type to identify the project. (Permitted values: `name` or `uuid`)

### project.name (`name` or `UUID` required)
Name of the project

### project.uuid (`name` or `UUID` required)
UUID of the project

### additionalCategories (optional)
Reference to a list of existing [Nutanix Categories](https://portal.nutanix.com/page/documents/details?targetId=Prism-Central-Guide:ssp-ssp-categories-manage-pc-c.html) to be assigned to virtual machines.

### additionalCategories[0].key
Nutanix Category to add to the virtual machine.

### additionalCategories[0].value
Value of the Nutanix Category to add to the virtual machine

### users (optional)
The users you want to configure to access your virtual machines. Only one is permitted at this time.

### users[0].name (optional)
The name of the user you want to configure to access your virtual machines through ssh.

The default is `eksa` if `osFamily=ubuntu`

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your virtual machines through ssh (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster VMs so you can ssh into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<VM-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value
