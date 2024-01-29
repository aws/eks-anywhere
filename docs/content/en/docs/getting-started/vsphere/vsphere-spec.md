---
title: "Configure for vSphere"
linkTitle: "Configuration"
weight: 20
aliases:
    /docs/reference/clusterspec/vsphere/
description: >
  Full EKS Anywhere configuration reference for a VMware vSphere cluster.
---

This is a generic template with detailed descriptions below for reference.

Key: <span style="color:green">Resources are in green</span> ; <span style="color:blue">Links to field descriptions are in blue</span>

<pre>
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name             <a href="#name-required"># Name of the cluster (required)</a>
spec:
   clusterNetwork:                   <a href="#clusternetwork-required"># Cluster network configuration (required)</a>
      cniConfig:                     <a href="#clusternetworkcniconfig-required"># Cluster CNI plugin - default: cilium (required)</a>
         cilium: {}
      pods:
         cidrBlocks:                 <a href="#clusternetworkpodscidrblocks0-required"># Subnet CIDR notation for pods (required)</a>
            - 192.168.0.0/16
      services:
         cidrBlocks:                 <a href="#clusternetworkservicescidrblocks0-required"># Subnet CIDR notation for services (required)</a>
            - 10.96.0.0/12
   controlPlaneConfiguration:        <a href="#controlplaneconfiguration-required"># Specific cluster control plane config (required)</a>
      count: <span style="color:green">2</span>                       <a href="#controlplaneconfigurationcount-required"># Number of control plane nodes (required)</a>
      endpoint:                      <a href="#controlplaneconfigurationendpointhost-required"># IP for control plane endpoint (required)</a>
         host: <span>"192.168.0.10"</span>
      machineGroupRef:               <a href="#controlplaneconfigurationmachinegroupref-required"># vSphere-specific Kubernetes node config (required)</a>
        kind: VSphereMachineConfig
        name: my-cluster-machines
      taints:                        <a href="#controlplaneconfigurationtaints"># Taints applied to control plane nodes </a>
      - key: <span>"key1"</span>
        value: <span>"value1"</span>
        effect: <span>"NoSchedule"</span>
      labels:                        <a href="#controlplaneconfigurationlabels"># Labels applied to control plane nodes </a>
        <span>"key1"</span>: <span>"value1"</span>
        <span>"key2"</span>: <span>"value2"</span>
   datacenterRef:                    <a href="#datacenterref"># Kubernetes object with vSphere-specific config </a>
      kind: VSphereDatacenterConfig
      name: my-cluster-datacenter
   externalEtcdConfiguration:
     count: <span style="color:green">3</span>                        <a href="#externaletcdconfigurationcount"># Number of etcd members </a>
     machineGroupRef:                <a href="#externaletcdconfigurationmachinegroupref"># vSphere-specific Kubernetes etcd config</a>
        kind: VSphereMachineConfig
        name: my-cluster-machines
   kubernetesVersion: <span>"1.25"</span>         <a href="#kubernetesversion-required"># Kubernetes version to use for the cluster (required)</a>
   workerNodeGroupConfigurations:    <a href="#workernodegroupconfigurations-required"># List of node groups you can define for workers (required) </a>
   - count: <span style="color:green">2</span>                        <a href="#workernodegroupconfigurationscount"># Number of worker nodes </a>
     machineGroupRef:                <a href="#workernodegroupconfigurationsmachinegroupref-required"># vSphere-specific Kubernetes node objects (required) </a>
       kind: VSphereMachineConfig
       name: my-cluster-machines
     name: md-0                      <a href="#workernodegroupconfigurationsname-required"># Name of the worker nodegroup (required) </a>
     taints:                         <a href="#workernodegroupconfigurationstaints"># Taints to apply to worker node group nodes </a>
     - key: <span>"key1"</span>
       value: <span>"value1"</span>
       effect: <span>"NoSchedule"</span>
     labels:                         <a href="#workernodegroupconfigurationslabels"># Labels to apply to worker node group nodes </a>
       <span>"key1"</span>: <span>"value1"</span>
       <span">"key2"</span>: <span>"value2"</span>
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereDatacenterConfig
metadata:
   name: my-cluster-datacenter
spec:
  datacenter: <span>"datacenter1"</span>          <a href="#datacenter-required"># vSphere datacenter name on which to deploy EKS Anywhere (required) </a>
  server: <span>"myvsphere.local"</span>          <a href="#server-required"># FQDN or IP address of vCenter server (required) </a>
  network: <span>"network1"</span>                <a href="#network-required"># Path to the VM network on which to deploy EKS Anywhere (required) </a>
  insecure: false                    <a href="#insecure-optional"># Set to true if vCenter does not have a valid certificate </a>
  thumbprint: <span>"1E:3B:A1:4C:B2:..."</span>   <a href="#thumbprint-required-if-insecurefalse"># SHA1 thumprint of vCenter server certificate (required if insecure=false)</a>

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
   name: my-cluster-machines
spec:
  diskGiB:  <span style="color:green">25</span>                         <a href="#diskgib-optional"># Size of disk on VMs, if no snapshots</a>
  datastore: <span>"datastore1"</span>              <a href="#datastore-required"># Path to vSphere datastore to deploy EKS Anywhere on (required)</a>
  folder: <span>"folder1"</span>                    <a href="#folder-required"># Path to VM folder for EKS Anywhere cluster VMs (required)</a>
  numCPUs: <span style="color:green">2</span>                           <a href="#numcpus-optional"># Number of CPUs on virtual machines</a>
  memoryMiB: <span style="color:green">8192</span>                      <a href="#memorymib-optional"># Size of RAM on VMs</a>
  osFamily: <span>"bottlerocket"</span>             <a href="#osfamily-optional"># Operating system on VMs</a>
  resourcePool: <span>"resourcePool1"</span>        <a href="#resourcepool-required"># vSphere resource pool for EKS Anywhere VMs (required)</a>
  storagePolicyName: <span>"storagePolicy1"</span>  <a href="#storagepolicyname-optional"># Storage policy name associated with VMs</a>
  template: <span>"bottlerocket-kube-v1-25"</span>  <a href="#template-optional"># VM template for EKS Anywhere (required for RHEL/Ubuntu-based OVAs)</a>
  cloneMode: <span>"fullClone"</span>               <a href="#clonemode-optional"># Clone mode to use when cloning VMs from the template</a>
  users:                               <a href="#users-optional"># Add users to access VMs via SSH</a>
  - name: <span>"ec2-user"</span>                   <a href="#users0name-optional"># Name of each user set to access VMs</a>
    sshAuthorizedKeys:                 <a href="#users0sshauthorizedkeys-optional"># SSH keys for user needed to access VMs</a>
    - <span>"ssh-rsa AAAAB3NzaC1yc2E..."</span>
  tags:                                <a href="#tags-optional"># List of tags to attach to cluster VMs, in URN format</a>
  - <span>"urn:vmomi:InventoryServiceTag:5b3e951f-4e1d-4511-95b1-5ba1ea97245c:GLOBAL"</span>
  - <span>"urn:vmomi:InventoryServiceTag:cfee03d0-0189-4f27-8c65-fe75086a86cd:GLOBAL"</span>
</pre>

The following additional optional configuration can also be included:

* [CNI]({{< relref "../optional/cni.md" >}})
* [IAM Roles for Service Accounts]({{< relref "../optional/irsa.md" >}})
* [IAM Authenticator]({{< relref "../optional/iamauth.md" >}})
* [OIDC]({{< relref "../optional/oidc.md" >}})
* [Gitops]({{< relref "../optional/gitops.md" >}})
* [Proxy]({{< relref "../optional/proxy.md" >}})
* [Registry Mirror]({{< relref "../optional/registrymirror.md" >}})
* [Host OS Config]({{< relref "../optional/hostOSConfig.md" >}})
* [Machine Health Check Timeouts]({{< relref "../optional/healthchecks.md" >}})

## Cluster Fields

### name (required)
Name of your cluster `my-cluster-name` in this example

{{% include "../_configuration/cluster_clusterNetwork.html" %}}

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with vsphere specific configuration for your nodes.  See [VSphereMachineConfig Fields](#vspheremachineconfig-fields) below.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster
creation process are [here]({{< relref "../vsphere/vsphere-prereq/#prepare-a-vmware-vsphere-environment" >}})

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint. For k8s versions prior to 1.24, it replaces `node-role.kubernetes.io/master`. For k8s versions 1.24+, it replaces `node-role.kubernetes.io/control-plane`. The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint.
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

### workerNodeGroupConfigurations.count
Number of worker nodes. Optional if the [cluster autoscaler curated package]({{< relref "../../packages/cluster-autoscaler/addclauto" >}}) is installed and autoscalingConfiguration is used, in which case count will default to `autoscalingConfiguration.minCount`.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with vsphere specific configuration for your nodes. See [VSphereMachineConfig Fields](#vspheremachineconfig-fields) below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations.autoscalingConfiguration.minCount
Minimum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.autoscalingConfiguration.maxCount
Maximum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must **NOT** have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

### workerNodeGroupConfigurations.kubernetesVersion
The Kubernetes version you want to use for this worker node group. [Supported values]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}): `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

Must be less than or equal to the cluster `kubernetesVersion` defined at the root level of the cluster spec. The worker node kubernetesVersion must be no more than two minor Kubernetes versions lower than the cluster control plane's Kubernetes version. Removing `workerNodeGroupConfiguration.kubernetesVersion` will trigger an upgrade of the node group to the `kubernetesVersion` defined at the root level of the cluster spec.

### externalEtcdConfiguration.count
Number of etcd members

### externalEtcdConfiguration.machineGroupRef
Refers to the Kubernetes object with vsphere specific configuration for your etcd members.  See [VSphereMachineConfig Fields](#vspheremachineconfig-fields) below.

### datacenterRef
Refers to the Kubernetes object with vsphere environment specific configuration.  See [VSphereDatacenterConfig Fields](#vspheredatacenterconfig-fields) below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. [Supported values]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}): `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

## VSphereDatacenterConfig Fields

### datacenter (required)
The name of the vSphere datacenter to deploy the EKS Anywhere cluster on. For example `SDDC-Datacenter`.

### network (required)
The path to the VM network to deploy your EKS Anywhere cluster on. For example, `/<DATACENTER>/network/<NETWORK_NAME>`.
Use `govc find -type n` to see a list of networks.

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
Operating System on virtual machines. Permitted values: bottlerocket, ubuntu, redhat (Default: bottlerocket)

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
[imported the OVA file into vSphere]({{< relref "../vsphere/customize/vsphere-ovas.md" >}}).
This is a required field if you are using Ubuntu-based or RHEL-based OVAs.
The `template` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.24, `template` must include 1.24, 1_24, 1-24 or 124.

### cloneMode (optional)
`cloneMode` defines the clone mode to use when creating the cluster VMs from the template. Allowed values are:
- `fullClone`: With full clone, the cloned VM is a separate independent copy of the template. This makes provisioning the VMs a bit slower at the cost of better customization and performance.
- `linkedClone`: With linked clone, the cloned VM shares the parent template's virtual disk. This makes provisioning the VMs faster while also saving the disk space. Linked clone does **not** allow customizing the disk size.
The template should meet the following properties to use `linkedClone`:
  - The template needs to have a snapshot
  - The template's disk size must match the VSphereMachineConfig's diskGiB

If this field is not specified, EKS Anywhere tries to determine the clone mode based on the following criteria:
- It uses linkedClone if the template has snapshots and the template diskSize matches the machineConfig DiskGiB.
- Otherwise, it uses use full clone.

### datastore (required)
The path to the vSphere [datastore](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.storage.doc/GUID-3CC7078E-9C30-402C-B2E1-2542BEE67E8F.html)
to deploy your EKS Anywhere cluster on, for example `/<DATACENTER>/datastore/<DATASTORE_NAME>`.
Use `govc find -type s` to get a list of datastores.

### folder (required)
The path to a VM folder for your EKS Anywhere cluster VMs. This allows you to organize your VMs. If the folder does not exist,
it will be created for you. If the folder is blank, the VMs will go in the root folder.
For example `/<DATACENTER>/vm/<FOLDER_NAME>/...`.
Use `govc find -type f` to get a list of existing folders.


### resourcePool (required)
The vSphere [Resource pools](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.resmgmt.doc/GUID-60077B40-66FF-4625-934A-641703ED7601.html)
for your VMs in the EKS Anywhere cluster. Examples of resource pool values include:

* If there is no resource pool: `/<datacenter>/host/<cluster-name>/Resources`
* If there is a resource pool:  `/<datacenter>/host/<cluster-name>/Resources/<resource-pool-name>`
* The wild card option `*/Resources` also often works.

Use `govc find -type p` to get a list of available resource pools.

### storagePolicyName (optional)
The storage policy name associated with your VMs. Generally this can be left blank.
Use `govc storage.policy.ls` to get a list of available storage policies.

### tags (optional)
Optional list of tags to attach to your cluster VMs in the URN format.

Example:
```
  tags:
  - urn:vmomi:InventoryServiceTag:8e0ce079-0675-47d6-8665-16ada4e6dabd:GLOBAL
```

### hostOSConfig (optional)
Optional host OS configurations for the EKS Anywhere Kubernetes nodes.
More information in the [Host OS Configuration]({{< relref "../optional/hostOSConfig.md" >}}) section.

## Optional VSphere Credentials
Use the following environment variables to configure the Cloud Provider with different credentials.

### EKSA_VSPHERE_CP_USERNAME
Username for Cloud Provider (Default: $EKSA_VSPHERE_USERNAME).

### EKSA_VSPHERE_CP_PASSWORD
Password for Cloud Provider (Default: $EKSA_VSPHERE_PASSWORD).
