---
title: "Overview"
linkTitle: "Overview"
weight: 1
aliases:
    /docs/concepts/clusterworkflow
date: 2023-08-11
description: >
  Overview of the EKS Anywhere cluster creation process
---

#### Overview

Kubernetes clusters require infrastructure capacity for the Kubernetes control plane, etcd, and worker nodes. EKS Anywhere provisions and manages this capacity on your behalf when you create EKS Anywhere clusters by interacting with the underlying infrastructure interfaces. Today, EKS Anywhere supports vSphere, bare metal, Snow, Apache CloudStack and Nutanix infrastructure providers. EKS Anywhere can also run on Docker for dev/test and non-production deployments only.

If you are creating your first EKS Anywhere cluster, you must first prepare an Administrative machine (Admin machine) where you install and run the EKS Anywhere CLI. The EKS Anywhere CLI (`eksctl anywhere`) is the primary tool you will use to create and manage your first cluster. 

Your interface for configuring EKS Anywhere clusters is the cluster specification yaml (cluster spec). This cluster spec is where you define cluster configuration including cluster name, network, Kubernetes version, control plane settings, worker node settings, and operating system. You also specify environment-specific configuration in the cluster spec for vSphere, bare metal, Snow, CloudStack, and Nutanix. When you perform cluster lifecycle operations, you modify the cluster spec, and then apply the cluster spec changes to your cluster in a declarative manner.

Before creating EKS Anywhere clusters, you must determine the operating system you will use. EKS Anywhere supports Bottlerocket, Ubuntu, and Red Hat Enterprise Linux (RHEL). All operating systems are not supported on each infrastructure provider. If you are using Ubuntu or RHEL, you must build your images before creating your cluster. For details reference the [Operating System Management documenation]({{< relref "../osmgmt" >}})

During initial cluster creation, the EKS Anywhere CLI performs the following high-level actions
- Confirms the target cluster environment is available
- Confirms authentication succeeds to the target environment
- Performs infrastructure provider-specific validations
- Creates a bootstrap cluster (Kind cluster) on the Admin machine
- Installs Cluster API (CAPI) and EKS-A core components on the bootstrap cluster
- Creates the EKS Anywhere cluster on the infrastructure provider
- Moves the Cluster API and EKS-A core components from the bootstrap cluster to the EKS Anywhere cluster
- Shuts down the bootstrap cluster

During initial cluster creation, you can observe the progress through the EKS Anywhere CLI output and by monitoring the CAPI and EKS-A controller manager logs on the bootstrap cluster. To access the bootstrap cluster, use the `kubeconfig` file in the `<cluster-name>/generated/<cluster-name>.kind.kubeconfig` file location.

After initial cluster creation, you can access your cluster using the `kubeconfig` file, which is located in the `<cluster-name>/<cluster-name>-eks-a-cluster.kubeconfig` file location. You can SSH to the nodes that EKS Anywhere created on your behalf with the keys in the `<cluster-name>/eks-a-id_rsa` location by default. 

While you do not need to maintain your Admin machine, you must save your `kubeconfig`, SSH keys, and EKS Anywhere cluster spec to a safe location if you intend to use a different Admin machine in the future. 

See the [Admin machine]({{< relref "./install" >}}) page for details and requirements to get started setting up your Admin machine.

#### Infrastructure Providers

EKS Anywhere uses an infrastructure provider model for creating, upgrading, and managing Kubernetes clusters that is based on the [Kubernetes Cluster API](https://cluster-api.sigs.k8s.io/) (CAPI) project.

Like CAPI, EKS Anywhere runs a [Kind](https://kind.sigs.k8s.io/) cluster on the Admin machine to act as a bootstrap cluster. However, instead of using CAPI directly with the `clusterctl` command to manage EKS Anywhere clusters, you use the `eksctl anywhere` command which simplifies that operation.

Before creating your first EKS Anywhere cluster, you must choose your infrastructure provider and ensure the requirements for that environment are met. Reference the infrastructure provider-specific sections below for more information.
- [VMWare vSphere]({{< relref "./vsphere" >}}) 
- [Bare Metal]({{< relref "./baremetal" >}}) 
- [Snow]({{< relref "./snow" >}}) 
- [CloudStack]({{< relref "./cloudstack" >}}) 
- [Nutanix]({{< relref "./nutanix" >}}) 

#### Deployment Architectures

EKS Anywhere supports two deployment architectures:

* **Standalone clusters**: If you want only a single EKS Anywhere cluster, you can deploy a standalone cluster.
This deployment type runs the CAPI and EKS-A management components on a single standalone cluster alongside the Kubernetes cluster that runs workloads. Standalone clusters must be managed with the EKS Anywhere CLI. A standalone cluster is effectively a management cluster, but in this deployment type, only manages itself.

* **Management cluster with separate workload clusters**: If you plan to deploy multiple EKS Anywhere clusters, you should deploy a management cluster with separate workload clusters. With this deployment type, the management cluster is used to perform cluster lifecycle operations on a fleet of workload clusters. The management cluster must be managed with the EKS Anywhere CLI, whereas workload clusters can be managed with the EKS Anywhere CLI, Kubernetes API-compatible tooling, or with Infrastructure as Code (IAC) tooling such as Terraform or GitOps.

For details on the EKS Anywhere architectures, see the [Architecture page.]({{< relref "../concepts/architecture.md" >}}) 

#### EKS Anywhere software

When setting up your Admin machine, you need Internet access to the repositories hosting the EKS Anywhere software.
EKS Anywhere software is divided into two types of components: The EKS Anywhere CLI for managing clusters and the cluster components and controllers used to run workloads and configure clusters.

* **Command line tools**: Binaries installed on the Admin machine include `eksctl`, `eksctl-anywhere`, `kubectl`, and `aws-iam-authenticator`.
* **Cluster components and controllers**: These components are listed on the [artifacts]({{< relref "../osmgmt/artifacts" >}}) page for each provider.

If you are operating behind a firewall that limits access to the Internet, you can configure EKS Anywhere to use a [proxy service]({{< relref "../getting-started/optional/proxy" >}}) to connect to the Internet.

For more information on the software used in EKS Distro, which includes the Kubernetes release and related software in EKS Anywhere, see the [EKS Distro Releases](https://distro.eks.amazonaws.com/#releases) page.
