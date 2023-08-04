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
Docker is available for non-production environments.
We step through the cluster creation workflow for Bare Metal, vSphere, and Nutanix providers below.


## Management and workload clusters

EKS Anywhere offers two cluster deployment topology options:

* **Standalone cluster**: If you want only a single EKS Anywhere cluster, you can deploy a self-managed, standalone cluster.
This type of cluster contains all Cluster API (CAPI) management components needed to manage itself, including managing its own upgrades.
It can also run workloads.

* **Management cluster with workload clusters**: If you plan to deploy multiple clusters, you should first deploy a _management cluster_.
The management cluster can then be used to deploy, upgrade, delete, and otherwise manage a fleet of _workload clusters_.

For further details about the different cluster topologies, see [Architecture.]({{< relref "architecture.md" >}})

## Before cluster creation

Some assets need to be in place before you can create an EKS Anywhere cluster.
You need to have an Administrative machine that includes the tools required to create the cluster.
Next, you need get the software tools and artifacts used to build the cluster.
Then you also need to prepare the provider, such as a vCenter environment, a Prism Central environment, or a set of Bare Metal machines, on which to create the resulting cluster. 

### Administrative machine

The Administrative machine is needed to provide:

* A place to run the commands to create and manage the target cluster.
* A Docker container runtime to run a temporary, local bootstrap cluster that creates the resulting target cluster.
* A place to hold the `kubeconfig` file needed to perform administrative actions using `kubectl`.
(The `kubeconfig` file is stored in the root of the folder created during cluster creation.)

See the [Install EKS Anywhere]({{< relref "../getting-started/install" >}}) guide for Administrative machine requirements.

### EKS Anywhere software

To obtain EKS Anywhere software, you need Internet access to the repositories holding that software.
EKS Anywhere software is divided into two types of components:
The CLI for managing clusters and the cluster components and controllers used to run workloads and configure clusters.
The software you need to obtain includes:

* **Command line tools**: Binaries to [install on the Administrative machine]({{< relref "../getting-started/install" >}}) include `eksctl`, `eksctl-anywhere`, `kubectl`, and `aws-iam-authenticator`.
* **Cluster components and controllers**: These components are listed on the [artifacts]({{< relref "../osmgmt/artifacts" >}}) page for each provider.

If you are operating behind a firewall that limits access to the Internet, you can configure EKS Anywhere to use a [proxy service]({{< relref "../getting-started/optional/proxy" >}}) to connect to the Internet.

For more information on the software used in EKS Distro, which includes the Kubernetes release and related software in EKS Anywhere, see the [EKS Distro Releases](https://distro.eks.amazonaws.com/#releases) page.

### Providers

EKS Anywhere uses an infrastructure provider model for creating, upgrading, and managing Kubernetes clusters that leverages the [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) project.

Like Cluster API, EKS Anywhere runs a [kind](https://kind.sigs.k8s.io/) cluster on the local Administrative machine to act as a bootstrap cluster.
However, instead of using CAPI directly with the `clusterctl` command to manage the workload cluster, you use the `eksctl anywhere` command which abstracts that process for you, including calling `clusterctl` under the covers.

With your Administrative machine in place, you need to prepare your [provider]({{< relref "../getting-started/chooseprovider/" >}}) for EKS Anywhere.
The following sections describe how to create a Bare Metal, vSphere or Nutanix cluster.

### Cluster Network
EKS Anywhere clusters use the `clusterNetwork` field in the cluster spec to allocate pod and service IPs. Once the cluster is created, the `pods.cidrBlocks`, `services.cidrBlocks` and `nodes.cidrMaskSize` fields are immutable. As a result, extra care should be taken to ensure that there are sufficient IPs and IP blocks available when provisioning large clusters.
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
    services:
      cidrBlocks:
      - 10.96.0.0/12
```

The cluster `pods.cidrBlocks` is subdivided between nodes with a default block of size `/24` per node, which can also be [configured via]({{< relref "../getting-started/optional/cni/#node-ips-configuration-option" >}}) the  `nodes.cidrMaskSize` field. This node CIDR block is then used to assign pod IPs on the node.

{{% alert title="Warning" color="warning" %}}
The maximum number of nodes will be limited to the number of subnets of size `/24` (or `nodes.cidrMaskSize` if configured) that can fit in the cluster `pods.cidrBlocks`.

The maximum number of pods per node is also limited by the size of the node CIDR block. For example with the default `/24` node CIDR mask size, there are a maximum of 256 IPs available for pods. Kubernetes recommends [no more than 110 pods per node.](https://kubernetes.io/docs/setup/best-practices/cluster-large/)
{{% /alert %}}
