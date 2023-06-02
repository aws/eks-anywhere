---
title: "Overview"
linkTitle: "Overview"
weight: 1
aliases:
    /docs/concepts/clusterworkflow
date: 2017-01-05
description: >
  Explanation of the process of creating an EKS Anywhere cluster
---

Each EKS Anywhere cluster is built from a cluster specification file, with the structure of the configuration file based on the target provider for the cluster.
Currently, Bare Metal, CloudStack, Nutanix, Snow, and VMware vSphere are the recommended providers for supported EKS Anywhere clusters.
Docker is available as an unsupported provider.
We step through the cluster creation workflow for Bare Metal and vSphere providers here.


## Management and workload clusters

EKS Anywhere offers two cluster deployment topology options:

* **Standalone cluster**: If want only a single EKS Anywhere cluster, you can deploy a self-managed, standalone cluster.
This type of cluster contains all Cluster API (CAPI) management components needed to manage itself, including managing its own upgrades.
It can also run workloads.

* **Management cluster with workload clusters**: If you plan to deploy multiple clusters, the project recommends you first deploy a _management cluster_.
The management cluster can then be used to deploy, upgrade, delete, and otherwise manage a fleet of _workload clusters_.

For further details about the different cluster topologies, see [Architecture]({{< relref "architecture.md" >}})

## Before cluster creation

Some assets need to be in place before you can create an EKS Anywhere cluster.
You need to have an Administrative machine that includes the tools required to create the cluster.
Next, you need get the software tools and artifacts used to build the cluster.
Then you also need to prepare the provider, such as a vCenter environment or a set of Bare Metal machines, on which to create the resulting cluster. 

### Administrative machine

The Administrative machine is needed to provide:

* A place to run the commands to create and manage the target cluster.
* A Docker container runtime to run a temporary, local bootstrap cluster that creates the resulting target cluster on the vSphere provider.
* A place to hold the `kubeconfig` file needed to perform administrative actions using `kubectl`.
(The `kubeconfig` file is stored in the root of the folder created during cluster creation.)

See the [Install EKS Anywhere]({{< relref "../getting-started/install" >}}) guide for Administrative machine requirements.

### EKS Anywhere software

To obtain EKS Anywhere software, you need Internet access to the repositories holding that software.
EKS Anywhere software is divided into two types of components:
The CLI interface for managing clusters and the cluster components and controllers used to run workloads and configure clusters.
The software you need to obtain includes:

* **Command line tools**: Binaries to [install on the administrative machine]({{< relref "../getting-started/install" >}}), include `eksctl`, `eksctl-anywhere`, `kubectl`, and `aws-iam-authenticator`.
* **Cluster components and controllers**: These components are listed on the [artifacts]({{< relref "../osmgmt/artifacts" >}}) page for each provider.

If you are operating behind a firewall that limits access to the Internet, you can configure EKS Anywhere to identify the location of the [proxy service]({{< relref "../getting-started/optional/proxy" >}}) you choose to connect to the Internet.

For more information on the software used in EKS Distro, which includes the Kubernetes release and related software in EKS Anywhere, see the [EKS Distro Releases](https://distro.eks.amazonaws.com/#releases) GitHub page.

### Providers

EKS Anywhere uses an infrastructure provider model for creating, upgrading, and managing Kubernetes clusters that leverages the [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) project.

Like Cluster API, EKS Anywhere runs a [kind](https://kind.sigs.k8s.io/) cluster on the local Administrative machine to act as a bootstrap cluster.
However, instead of using CAPI directly with the `clusterctl` command to manage the workload cluster, you use the `eksctl anywhere` command which abstracts that process for you, including calling `clusterctl` under the covers.

With your Administrative machine in place, you need to prepare your [provider]({{< relref "../getting-started/chooseprovider/" >}}) for EKS Anywhere.
The following sections describe how to create a Bare Metal or vSphere cluster.

## Creating a Bare Metal cluster
The following diagram illustrates what happens when you start the cluster creation process for a Bare Metal provider, as described in the [Bare Metal Getting started]({{< relref "../getting-started/baremetal" >}}) guide.

### Start creating a Bare Metal cluster

![Start creating EKS Anywhere Bare Metal cluster](/images/eksa-baremetal-start.png)

#### 1. Generate a config file for Bare Metal
Identify the provider (`--provider tinkerbell`) and the cluster name to the `eksctl anywhere create clusterconfig` command and direct the output into a cluster config `.yaml` file.

#### 2. Modify the config file and hardware CSV file
Modify the generated cluster config file to suit your situation.
Details about this config file are contained on the [Bare Metal Config]({{< relref "../getting-started/baremetal/bare-spec/" >}}) page.
Create a hardware configuration file (`hardware.csv`) as described in [Prepare hardware inventory]({{< relref "../getting-started/baremetal/bare-preparation/#prepare-hardware-inventory" >}}).

#### 3. Launch the cluster creation

Run the `eksctl anywhere cluster create` command, providing the cluster config and hardware CSV files.
To see details on the cluster creation process, increase verbosity (`-v=9` provides maximum verbosity).

#### 4. Create bootstrap cluster and provision hardware

The cluster creation process starts by creating a temporary Kubernetes bootstrap cluster on the Administrative machine.
Containerized components of the Tinkerbell provisioner run either as pods on the bootstrap cluster (Hegel, Rufio, and Tink) or directly as containers on Docker (Boots).
Those Tinkerbell components drive the provisioning of the operating systems and Kubernetes components on each of the physical computers.

With the information gathered from the cluster specification and the hardware CSV file, three custom resource definitions (CRDs) are created.
These include:

* Hardware custom resources: Which store hardware information for each machine
* Template custom resources: Which store the tasks and actions
* Workflow custom resources: Which put together the complete hardware and template information for each machine. There are different workflows for control plane and worker nodes.

As the bootstrap cluster comes up and Tinkerbell components are started, you should see messages like the following:

```bash
$ eksctl anywhere create cluster --hardware-csv hardware.csv -f eksa-mgmt-cluster.yaml
Performing setup and validations
 Tinkerbell Provider setup is valid
 Validate certificate for registry mirror
 Create preflight validations pass
Creating new bootstrap cluster
Provider specific pre-capi-install-setup on bootstrap cluster
Installing cluster-api providers on bootstrap cluster
Provider specific post-setup
Creating new workload cluster
```
At this point, Tinkerbell will try to boot up the machines in the target cluster.

### Continuing cluster creation

Tinkerbell takes over the activities for creating provisioning the Bare Metal machines to become the new target cluster.
See [Overview of Tinkerbell in EKS Anywhere]({{< relref "../getting-started/baremetal/overview/#overview-of-tinkerbell-in-eks-anywhere" >}}) for examples of commands you can run to watch over this process.

![Continue creating EKS Anywhere Bare Metal cluster](/images/eksa-baremetal-continue.png)

#### 1. Tinkerbell network boots and configures nodes

* Rufio uses BMC information to set the power state for the first control plane node it wants to provision.
* When the node boots from its NIC, it talks to the Boots DHCP server, which fetches the kernel and initramfs (HookOS) needed to network boot the machine.
* With HookOS running on the node, the operating system identified by `IMG_URL` in the cluster specification is copied to the identified `DEST_DISK` on the machine.
* The Hegel components provides data stores that contain information used by services such as cloud-init to configure each system.
* Next, the workflow is run on the first control plane node, followed by network booting and running the workflow for each subsequent control plane node.
* Once the control plane is up, worker nodes are network booted and workflows are run to deploy each node.

#### 2. Tinkerbell components move to the target cluster

Once all the defined nodes are added to the cluster, the Tinkerbell components and associated data are moved to run as pods on worker nodes in the new workload cluster.

### Deleting Tinkerbell from Admin machine

All Tinkerbell-related pods and containers are then deleted from the Admin machine.
Further management of tinkerbell and related information can be done using from the new cluster, using tools such as `kubectl`.

![Delete Tinkerbell pods and container](/images/eksa-baremetal-delete.png)

## Creating a vSphere cluster

The following diagram illustrates what happens when you start the cluster creation process, as described in the [vSphere Getting started]({{< relref "../getting-started/vsphere" >}}) guide.

### Start creating a vSphere cluster 
![Start creating EKS Anywhere cluster](/images/eksa-start.png)

#### 1. Generate a config file for vSphere

To this command, you identify the name of the provider (`-p vsphere`) and a cluster name and redirect the output to a file.
The result is a config file template that you need to modify for the specific instance of your provider. 


#### 2. Modify the config file

Using the generated cluster config file, make modifications to suit your situation.
Details about this config file are contained on the [vSphere Config]({{< relref "../getting-started/vsphere/" >}}) page.

#### 3. Launch the cluster creation

Once you have modified the cluster configuration file, use `eksctl anywhere cluster create -f $CLUSTER_NAME.yaml` starts the cluster creation process.
To see details on the cluster creation process, increase verbosity (`-v=9` provides maximum verbosity).

#### 4. Authenticate and create bootstrap cluster

After authenticating to vSphere and validating the assets there, the cluster creation process starts off creating a temporary Kubernetes bootstrap cluster on the Administrative machine.
To begin, the cluster creation process runs a series of [govc](https://github.com/vmware/govmomi/tree/master/govc) commands to check on the vSphere environment:

* Checks that the vSphere environment is available.

* Using the URL and credentials provided in the cluster spec files, authenticates to the vSphere provider.

* Validates the datacenter and the datacenter network exists:

* Validates that the identified datastore (to store your EKS Anywhere cluster) exists, that the folder holding your EKS Anywhere cluster VMs exists, and that the resource pools containing compute resources exist.
If you have multiple `VSphereMachineConfig` objects in your config file, will see these validations repeated:

* Validates the virtual machine templates to be used for the control plane and worker nodes (such as `ubuntu-2004-kube-v1.20.7`):


If all validations pass, you will see this message:

```
✅ Vsphere Provider setup is valid
```

Next, the process runs the [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) command to build a single-node Kubernetes bootstrap cluster on the Administrative machine.
This includes pulling the kind node image, preparing the node, writing the configuration, starting the control-plane, installing CNI, and installing the StorageClass. You will see:

After this point the bootstrap cluster is installed, but not yet fully configured.

### Continuing cluster creation

The following diagram illustrates the activities that occur next:

![Continue creating EKS Anywhere cluster](/images/eksa-continue.png)

#### 1. Add CAPI management

Cluster API (CAPI) management is added to the bootstrap cluster to direct the creation of the target cluster.

#### 2. Set up cluster

Configure the control plane and worker nodes.

#### 3. Add Cilium networking

Add Cilium as the CNI plugin to use for networking between the cluster services and pods.

#### 4. Add storage

Add the default storage class to the cluster

#### 5. Add CAPI to target cluster

Add the CAPI service to the target cluster in preparation for it to take over management of the cluster after the cluster creation is completed and the bootstrap cluster is deleted.
The bootstrap cluster can then begin moving the CAPI objects over to the target cluster, so it can take over the management of itself.

With the bootstrap cluster running and configured on the Administrative machine, the creation of the target cluster begins.
It uses `kubectl` to apply a target cluster configuration as follows:

* Once etcd, the control plane, and the worker nodes are ready, it applies the networking configuration to the target cluster.

* The default storage class is installed on the target cluster.

* CAPI providers are configured on the target cluster, in preparation for the target cluster to take over responsibilities for running the components needed to manage the itself.

* With CAPI running on the target cluster, CAPI objects for the target cluster are moved from the bootstrap cluster to the target cluster’s CAPI service (done internally with the `clusterctl` command):

* Add Kubernetes CRDs and other addons that are specific to EKS Anywhere.

* The cluster configuration is saved:

Once etcd, the control plane, and the worker nodes are ready, it applies the networking configuration to the workload cluster:

```
Installing networking on workload cluster
```

Next, the default storage class is installed on the workload cluster:

```
Installing storage class on workload cluster
```

After that, the CAPI providers are configured on the workload cluster, in preparation for the workload cluster to take over responsibilities for running the components needed to manage the itself.

```
Installing cluster-api providers on workload cluster
```

With CAPI running on the workload cluster, CAPI objects for the workload cluster are moved from the bootstrap cluster to the workload cluster’s CAPI service (done internally with the `clusterctl` command):

```
Moving cluster management from bootstrap to workload cluster
```

At this point, the cluster creation process will add Kubernetes CRDs and other addons that are specific to EKS Anywhere.
That configuration is applied directly to the cluster:

```
Installing EKS-A custom components (CRD and controller) on workload cluster
Creating EKS-A CRDs instances on workload cluster
Installing GitOps Toolkit on workload cluster

```
If you did not specify GitOps support, starting the flux service is skipped:

```
GitOps field not specified, bootstrap flux skipped

```
The cluster configuration is saved:

```
Writing cluster config file
```

With the cluster up, and the CAPI service running on the new cluster, the bootstrap cluster is no longer needed and is deleted:

![Delete EKS Anywhere bootstrap cluster](/images/eksa-delete.png)

At this point, cluster creation is complete.
You can now use your target cluster as either:

* A standalone cluster (to run workloads) or
* A management cluster (to optionally create one or more workload clusters)


### Creating workload clusters (optional)

As described in [Create separate workload clusters]({{< relref "./vsphere/vsphere-getstarted#create-separate-workload-clusters" >}}), you can use the cluster you just created as a management cluster to create and manage one or more workload clusters on the same vSphere provider as follows:

* Use `eksctl` to generate a cluster config file for the new workload cluster.
* Modify the cluster config with a new cluster name and different vSphere resources.
* Use `eksctl` to create the new workload cluster from the new cluster config file and credentials from the initial management cluster.
