---
title: "Create production cluster"
weight: 40
---

EKS Anywhere supports a vSphere provider for production grade EKS Anywhere deployments.
EKS Anywhere allows you to provision and manage Amazon EKS on your own infrastructure.

This document walks you through setting up EKS Anywhere in a way that:

* Deploys an initial cluster on your vSphere environment. That cluster can be used as a self-managed cluster (to run workloads) or a management cluster (to create and manage other clusters)
* Deploys zero or more workload clusters from the management cluster

If your initial cluster is a management cluster, it is intended to stay in place so you can use it later to modify, upgrade, and delete workload clusters.
Using a management cluster makes it faster to provision and delete workload clusters.
Also it lets you keep vSphere credentials for a set of clusters in one place: on the management cluster.
The alternative is to simply use your initial cluster to run workloads.

{{% alert title="Important" color="warning" %}}

Creating an EKS Anywhere management cluster is the recommended model.
Separating management features into a separate, persistent management cluster
provides a cleaner model for managing the lifecycle of workload clusters (to create, upgrade, and delete clusters), while workload clusters run user applications.
This approach also reduces provider permissions for workload clusters.

{{% /alert %}}

## Prerequisite Checklist

EKS Anywhere needs to be run on an administrative machine that has certain [machine
requirements]({{< relref "../install" >}}).
An EKS Anywhere deployment will also require the availability of certain
[resources from your VMware vSphere deployment]({{< relref "/docs/reference/vsphere/vsphere-prereq/_index.md" >}}).

## Steps

The following steps are divided into two sections:

* Create an initial cluster (used as a management or self-managed cluster)
* Create zero or more workload clusters from the management cluster

### Create an initial cluster

Follow these steps to create an EKS Anywhere cluster that can be used either as a management cluster or as a self-managed cluster (for running workloads itself).

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate an initial cluster config (named `mgmt` for this example):
   ```bash
   CLUSTER_NAME=mgmt
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider vsphere > eksa-mgmt-cluster.yaml
   ```

1. Modify the initial cluster config (`eksa-mgmt-cluster.yaml`) as follows:

   * Refer to [vsphere configuration]({{< relref "../../reference/clusterspec/vsphere" >}}) for information on configuring this cluster config for a vSphere provider.
   * Create at least two control plane nodes, three worker nodes, and three etcd nodes for a production cluster, to provide high availability and rolling upgrades.
   * Optionally, configure the cluster for [OIDC]({{< relref "/docs/reference/clusterspec/oidc" >}}), [etcd]({{< relref "/docs/reference/clusterspec/etcd" >}}), [proxy]({{< relref "/docs/reference/clusterspec/proxy" >}}), [gitops]({{< relref "/docs/reference/clusterspec/gitops" >}}) and/or [a container registry mirror]({{< relref "/docs/reference/clusterspec/registrymirror" >}}).

1. Set Credential Environment Variables

   Before you create the initial cluster, you will need to set and export these environment variables for your vSphere user name and password.
Make sure you use single quotes around the values so that your shell does not interpret the values:
   
   ```bash
   export EKSA_VSPHERE_USERNAME='billy'
   export EKSA_VSPHERE_PASSWORD='t0p$ecret'
   ```

1. Set License Environment Variable

   If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

   After you have created your `eksa-mgmt-cluster.yaml` and set your credential environment variables, you will be ready to create the cluster.


1. Create initial cluster: Create your initial cluster either with or without curated packages:
   - Cluster creation  without curated packages installation
      ```bash
      # Create a cluster without curated packages installation
      eksctl anywhere create cluster -f eksa-mgmt-cluster.yaml
      ```

   - Cluster creation with optional curated packages

     {{% alert title="Note" color="primary" %}}
   * It is *optional* to install the curated packages as part of the cluster creation.
   * `eksctl anywhere version` version should be `v0.9.0` or later.
   * If including curated packages during cluster creation, please set the environment variable: `export CURATED_PACKAGES_SUPPORT=true`
   * Post-creation installation and detailed package configurations can be found [here.]({{< relref "../../tasks/packages" >}})
   * The EKS Anywhere package controller and the EKS Anywhere Curated Packages (referred to as “features”) are provided as “preview features” subject to the AWS Service Terms, (including Section 2 (Betas and Previews)) of the same. During the EKS Anywhere Curated Packages Public Preview, the AWS Service Terms are extended to provide customers access to these features free of charge. These features will be subject to a service charge and fee structure at ”General Availability“ of the features.
     {{% /alert %}}

      * Discover curated packages to install
         ```bash
         eksctl anywhere list packages --source registry --kube-version 1.21
         ```
         Example command output
         ```                 
         Package                 Version(s)                                       
         -------                 ----------                                       
         harbor                  2.5.0-4324383d8c5383bded5f7378efb98b4d50af827b
         ```
      * Generate a curated-packages config

         The example shows how to install the `harbor` package from the [curated package list]({{< relref "../../reference/packagespec" >}}).
         ```bash
         eksctl anywhere generate package harbor --source registry --kube-version 1.21 > packages.yaml
         ```

      * Create the initial cluster

         ```bash
         # Create a cluster with curated packages installation
         eksctl anywhere create cluster -f eksa-mgmt-cluster.yaml --install-packages packages.yaml
         ```

1. Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory:

   ```bash
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   ```
1. Check the cluster nodes:

   To check that the cluster completed, list the machines to see the control plane, etcd, and worker nodes:

   ```bash
   kubectl get machines -A
   ```

   Example command output
   ```
   NAMESPACE   NAME                PROVIDERID        PHASE    VERSION
   eksa-system mgmt-b2xyz          vsphere:/xxxxx    Running  v1.21.2-eks-1-21-5
   eksa-system mgmt-etcd-r9b42     vsphere:/xxxxx    Running  
   eksa-system mgmt-md-8-6xr-rnr   vsphere:/xxxxx    Running  v1.21.2-eks-1-21-5
   ...
   ```

   The etcd machine doesn't show the Kubernetes version because it doesn't run the kubelet service.

1. Check the initial cluster's CRD:

   To ensure you are looking at the initial cluster, list the CRD to see that the name of its management cluster is itself:

   ```bash
   kubectl get clusters mgmt -o yaml
   ```

   Example command output
   ```
   ...
   kubernetesVersion: "1.21"
   managementCluster:
     name: mgmt
   workerNodeGroupConfigurations:
   ...
   ```

   {{% alert title="Note" color="primary" %}}
   The initial cluster is now ready to deploy workload clusters.
   However, if you just want to use it to run workloads, you can deploy pod workloads directly on the initial cluster without deploying a separate workload cluster and skip the section on running separate workload clusters.
   {{% /alert %}}

### Create separate workload clusters

Follow these steps if you want to use your initial cluster to create and manage separate workload clusters.

1. Generate a workload cluster config:
   ```bash
   CLUSTER_NAME=w01
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider vsphere > eksa-w01-cluster.yaml
   ```

   Refer to the initial config described earlier for the required and optional settings.
   The main differences are that you must have a new cluster name and cannot use the same vSphere resources.

1. Create a workload cluster

   To create a new workload cluster from your management cluster run this command, identifying:

   * The workload cluster YAML file
   * The initial cluster's credentials (this causes the workload cluster to be managed from the management cluster)

   ```bash
   # Create a cluster without curated packages installation
   eksctl anywhere create cluster \
       -f eksa-w01-cluster.yaml  \
       --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
   ```

   As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

1. Check the workload cluster:

   You can now use the workload cluster as you would any Kubernetes cluster.
   Change your credentials to point to the new workload cluster (for example, `mgmt-w01`), then run the test application with:

   ```bash
   export CLUSTER_NAME=mgmt-w01
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section]({{< relref "../../tasks/workload/test-app" >}}).

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `eksa-w02-cluster.yaml`), modifying resource names, and running the create cluster command again.

## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
