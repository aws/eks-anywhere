---
title: "vSphere configuration"
linkTitle: "vSphere"
weight: 10
description: >
  Full EKS Anywhere configuration reference for a VMware vSphere cluster.
---

This is a generic template with detailed descriptions below for reference

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   clusterNetwork:
      cni: "cilium"
      pods:
         cidrBlocks:
            - 192.168.0.0/16
      services:
         cidrBlocks:
            - 10.96.0.0/12
   controlPlaneConfiguration:
      count: 1
      endpoint:
         host: ""
      machineGroupRef:
        kind: VSphereMachineConfig
        name: my-cluster-machines
      taints:
      - key: ""
        value: ""
        effect: ""
      labels:
        "<key1>": ""
        "<key2>": "" 
   datacenterRef:
      kind: VSphereDatacenterConfig
      name: my-cluster-datacenter
   externalEtcdConfiguration:
     count: 3
     machineGroupRef:
        kind: VSphereMachineConfig
        name: my-cluster-machines
   kubernetesVersion: "1.21"
   workerNodeGroupConfigurations:
   - count: 1
     machineGroupRef:
       kind: VSphereMachineConfig
       name: my-cluster-machines
     name: md-0
     taints:
     - key: ""
       value: ""
       effect: ""
     labels:
       "<key1>": ""
       "<key2>": "" 
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereDatacenterConfig
metadata:
   name: my-cluster-datacenter
spec:
  datacenter: ""
  server: ""
  network: ""
  insecure:
  thumbprint: ""

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
   name: my-cluster-machines
spec:
  diskGiB:
  datastore: ""
  folder: ""
  numCPUs:
  memoryMiB:
  osFamily: ""
  resourcePool: ""
  storagePolicyName: ""
  template: ""
  users:
  - name: ""
    sshAuthorizedKeys:
    - ""
```

## Cluster Fields

The following additional optional configuration can also be included.

* [OIDC]({{< relref "oidc.md" >}})
* [etcd]({{< relref "etcd.md" >}})
* [proxy]({{< relref "proxy.md" >}})
* [gitops]({{< relref "gitops.md" >}})

### name (required)
Name of your cluster `my-cluster-name` in this example

### clusterNetwork (required)
Specific network configuration for your Kubernetes cluster.

### clusterNetwork.cni (required)
CNI plugin to be installed in the cluster. The only supported value at the moment is `cilium`.

### clusterNetwork.pods.cidrBlocks[0] (required)
Subnet used by pods in CIDR notation. Please note that only 1 custom pods CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the VMs.

### clusterNetwork.services.cidrBlocks[0] (required)
Subnet used by services in CIDR notation. Please note that only 1 custom services CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the VMs.

### clusterNetwork.dns.resolvConf.path (optional)
Path to the file with a custom DNS resolver configuration.

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with vsphere specific configuration for your nodes. See `VSphereMachineConfig Fields` below.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster 
creation process are [here]({{< relref "../vsphere/vsphere-prereq/#:~:text=Below%20are%20some,existent%20mac%20address." >}})

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

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers.
You may define one or more worker node groups.

### workerNodeGroupConfigurations.count (required)
Number of worker nodes

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with vsphere specific configuration for your nodes. See `VSphereMachineConfig Fields` below.

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

### externalEtcdConfiguration.count
Number of etcd members

### externalEtcdConfiguration.machineGroupRef
Refers to the Kubernetes object with vsphere specific configuration for your etcd members. See `VSphereMachineConfig Fields` below.

### datacenterRef
Refers to the Kubernetes object with vsphere environment specific configuration. See `VSphereDatacenterConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.20`, `1.21`

## VSphereDatacenterConfig Fields

### datacenter (required)
The vSphere datacenter to deploy the EKS Anywhere cluster on. For example `SDDC-Datacenter`.

### network (required)
The VM network to deploy your EKS Anywhere cluster on.

### server (required)
The vCenter server fully qualified domain name or IP address. If the server IP is used, the `thumbprint` must be set
or `insecure` must be set to true.

### insecure (optional)
Set insecure to `true` if the vCenter server does not have a valid certificate. (Default: false)

### thumbprint (required if insecure=false)
The SHA1 thumbprint of the vCenter server certificate which is only required if you have a self signed certificate.

There are several ways to obtain your vCenter thumbprint. The easiest way is if you have `govc` installed, you
can run:

```
govc about.cert -thumbprint -k
```

Another way is from the vCenter web UI, go to Administration/Certificate Management and click view details of the
machine certificate. The format of this thumbprint does not exactly match the format required though and you will
need to add `:` to separate each hexadecimal value.

Another way to get the thumbprint is use this command with your servers certificate in a file named `ca.crt`:

```
openssl x509 -sha1 -fingerprint -in ca.crt -noout
```

If you specify the wrong thumbprint, an error message will be printed with the expected thumbprint. If no valid
certificate is being used, `insecure` must be set to true.


## VSphereMachineConfig Fields

### memoryMiB (optional)
Size of RAM on virtual machines (Default: 8192)

### numCPUs (optional)
Number of CPUs on virtual machines (Default: 2)

### osFamily (optional)
Operating System on virtual machines. Permitted values: ubuntu, bottlerocket (Default: bottlerocket)

### diskGiB (optional)
Size of disk on virtual machines if snapshots aren't included (Default: 25)

### users (optional)
The users you want to configure to access your virtual machines. Only one is permitted at this time

### users[0].name (optional)
The name of the user you want to configure to access your virtual machines through ssh.

The default is `ec2-user` if `osFamily=bottlrocket` and `capv` if `osFamily=ubuntu`

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your virtual machines through ssh (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster VMs so you can ssh into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<VM-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value

### template (optional)
The VM template to use for your EKS Anywhere cluster. This template was created when you
[imported the OVA file into vSphere]({{< relref "../vsphere/vsphere-ovas.md" >}}).
This is a required field if you are using Bottlerocket OVAs.

### datastore (required)
The vSphere [datastore](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.storage.doc/GUID-3CC7078E-9C30-402C-B2E1-2542BEE67E8F.html)
to deploy your EKS Anywhere cluster on.

### folder (required)
The VM folder for your EKS anywhere cluster VMs. This allows you to organize your VMs. If the folder does not exist,
it will be created for you. If the folder is blank, the VMs will go in the root folder.

### resourcePool (required)
The vSphere [Resource pools](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.resmgmt.doc/GUID-60077B40-66FF-4625-934A-641703ED7601.html)
for your VMs in the EKS Anywhere cluster. Examples of resource pool values include:

* If there is no resource pool: `/<datacenter>/host/<cluster-name>/Resources`
* If there is a resource pool:  `/<datacenter>/host/<cluster-name>/Resources/<resource-pool-name>`
* The wild card option `*/Resources` also often works.

### storagePolicyName (optional)
The storage policy name associated with your VMs.
