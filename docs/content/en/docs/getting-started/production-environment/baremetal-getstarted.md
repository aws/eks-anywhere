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
See [Cluster topologies]({{< relref "../../concepts/cluster-topologies" >}}) for details.

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


1. Configure Curated Packages

   The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/). Cluster creation will succeed if authentication is not set up, but some warnings may be generated. Detailed package configurations can be found [here]({{< relref "../../tasks/packages" >}}).

   If you are going to use packages, set up authentication. These credentials should have [limited capabilities]({{< relref "../../tasks/packages/#setup-authentication-to-use-curated-packages" >}}):
   ```bash
   export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
   export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"  
   ```
     
1. Create the cluster, using the `hardware.csv` file you made in [Bare Metal preparation]({{< relref "/docs/reference/baremetal/bare-preparation.md" >}}):
   ```bash
   # Create a cluster without curated packages installation
   eksctl anywhere create cluster --hardware-csv hardware.csv -f eksa-mgmt-cluster.yaml
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
   NAMESPACE     NAME                        CLUSTER   NODENAME        PROVIDERID                              PHASE     AGE   VERSION
   eksa-system   mgmt-47zj8                  mgmt      eksa-node01     tinkerbell://eksa-system/eksa-node01    Running   1h    v1.23.7-eks-1-23-4
   eksa-system   mgmt-md-0-7f79df46f-wlp7w   mgmt      eksa-node02     tinkerbell://eksa-system/eksa-node02    Running   1h    v1.23.7-eks-1-23-4
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
