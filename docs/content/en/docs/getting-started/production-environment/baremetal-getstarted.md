---
title: "Create Bare Metal production cluster"
linkTitle: "Bare Metal cluster" 
weight: 40
description: >
  Create a production-quality cluster on Bare Metal
---

EKS Anywhere supports a Bare Metal provider for production grade EKS Anywhere deployments.
EKS Anywhere allows you to provision and manage Kubernetes clusters based on Amazon EKS software on your own infrastructure.

This document walks you through setting up EKS Anywhere as a self-managed cluster.
It does not yet support the concept of a separate management cluster for managing one or more workload clusters.

## Prerequisite checklist

EKS Anywhere needs:

* To be run on an Admin machine that has certain [machine requirements]({{< relref "../install" >}}).
* To meet certain [Bare Metal requirements]({{< relref "/docs/reference/baremetal/bare-prereq.md" >}}) for hardware and network configuration.
* To have some [Bare Metal preparation]({{< relref "/docs/reference/baremetal/bare-preparation.md" >}}) be in place before creating an EKS Anywhere cluster.

Also, see the [Ports and protocols]({{< relref "/docs/reference/ports.md" >}}) page for information on ports that need to be accessible from control plane, worker, and Admin machines.

## Steps

The following steps are needed to create a self-managed Bare Metal EKS Anywhere cluster.

### Create the cluster

Follow these steps to create an EKS Anywhere cluster.

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Set an environment variables for your cluster name
   
   ```bash
   export CLUSTER_NAME=mgmt
   ```
1. Generate a cluster config file for your Bare Metal provider (using tinkerbell as the provider type).
   ```bash
   eksctl anywhere generate clusterconfig $CLUSTER_NAME --provider tinkerbell > eksa-mgmt-cluster.yaml
   ```

1. Modify the cluster config (`eksa-mgmt-cluster.yaml`) by referring to the [Bare Metal configuration]({{< relref "../../reference/clusterspec/baremetal" >}}) reference documentation.

1. Set License Environment Variable

   If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

   After you have created your `eksa-mgmt-cluster.yaml` and set your credential environment variables, you will be ready to create the cluster.

1. Create the cluster, using the `hardware.csv` file you made in [Bare Metal preparation]({{< relref "/docs/reference/baremetal/bare-preparation.md" >}}),
   either with or without curated packages:
   - Cluster creation without curated packages installation
      ```bash
      # Create a cluster without curated packages installation
      eksctl anywhere create cluster --hardware-csv hardware.csv -f eksa-mgmt-cluster.yaml
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
         Example command output:
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
         eksctl anywhere create cluster -f eksa-mgmt-cluster.yaml \
            --hardware-csv hardware.csv --install-packages packages.yaml
         ```

1. Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory:

   ```bash
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   ```
1. Check the cluster nodes:

   To check that the cluster completed, list the machines to see the control plane and worker nodes:

   ```bash
   kubectl get machines -A
   ```

   Example command output:
   ```
   NAMESPACE   NAME                PROVIDERID         PHASE    VERSION
   eksa-system mgmt-b2xyz          tinkerbell:/xxxxx  Running  v1.21.2-eks-1-21-5
   eksa-system mgmt-md-8-6xr-rnr   tinkerbell:/xxxxx  Running  v1.21.2-eks-1-21-5
   ...
   ```

1. Check the cluster:

   You can now use the cluster as you would any Kubernetes cluster.
   To try it out, run the test application with:

   ```bash
   export CLUSTER_NAME=mgmt
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in [Deploy test workload]({{< relref "../../tasks/workload/test-app" >}}).

## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
