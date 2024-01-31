---
title: "Create Bare Metal cluster"
linkTitle: "3. Create cluster" 
weight: 20
description: >
  Create a cluster on Bare Metal
---

EKS Anywhere supports a Bare Metal provider for EKS Anywhere deployments.
EKS Anywhere allows you to provision and manage Kubernetes clusters based on Amazon EKS software on your own infrastructure.

This document walks you through setting up EKS Anywhere on Bare Metal as a standalone, self-managed cluster or combined set of management/workload clusters.
See [Cluster topologies]({{< relref "../../concepts/architecture" >}}) for details.

**Note:** Before you create your cluster, you have the option of validating the EKS Anywhere bundle manifest container images by following instructions in the [Verify Cluster Images]({{< relref "../../clustermgmt/verify-cluster-image.md" >}}) page.

## Prerequisite checklist

EKS Anywhere needs:

* To be run on an Admin machine that has certain [machine requirements]({{< relref "../install" >}}).
* To run a cluster in an [airgapped environment]({{< relref "../airgapped" >}}) (optional)
* To meet [networking]({{< relref "../ports/" >}}) requirements
* To meet certain [Bare Metal requirements]({{< relref "./bare-prereq/" >}}) for hardware and network configuration.
* To have some [Bare Metal preparation]({{< relref "./bare-preparation/" >}}) be in place before creating an EKS Anywhere cluster.

## Steps

The following steps are divided into two sections:

* Create an initial cluster (used as a management or self-managed cluster)
* Create zero or more workload clusters from the management cluster

### Create an initial cluster

Follow these steps to create an EKS Anywhere cluster that can be used either as a management cluster or as a self-managed cluster (for running workloads itself).

<!-- this content needs to be indented so the numbers are automatically incremented -->

0. Optional Configuration

   **Set License Environment Variable**

      Add a license to any cluster for which you want to receive paid support. If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/clustermgmt/support/cluster-license" >}}) if you are licensing an existing cluster):

      ```bash
      export EKSA_LICENSE='my-license-here'
      ```

      After you have created your `eksa-mgmt-cluster.yaml` and set your credential environment variables, you will be ready to create the cluster.

   **Configure Curated Packages**

      The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/). Cluster creation will succeed if authentication is not set up, but some warnings may be generated.  Detailed package configurations can be found [here.]({{< relref "../../packages" >}})

      If you are going to use packages, set up authentication. These credentials should have [limited capabilities:]({{< relref "../../packages/prereq#setup-authentication-to-use-curated-packages" >}})
      ```bash
      export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
      export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
      export EKSA_AWS_REGION="us-west-2"  
      ```

1. Set an environment variable for your cluster name:
   
   ```bash
   export CLUSTER_NAME=mgmt
   ```
1. Generate a cluster config file for your Bare Metal provider (using tinkerbell as the provider type).
   ```bash
   eksctl anywhere generate clusterconfig $CLUSTER_NAME --provider tinkerbell > eksa-mgmt-cluster.yaml
   ```

1. Modify the cluster config (`eksa-mgmt-cluster.yaml`) by referring to the [Bare Metal configuration]({{< relref "./bare-spec/" >}}) reference documentation.
     
1. Create the cluster, using the `hardware.csv` file you made in [Bare Metal preparation]({{< relref "./bare-preparation/" >}}).

   For a regular cluster create (with internet access), type the following:

   ```bash
   eksctl anywhere create cluster \
      --hardware-csv hardware.csv \
      -f eksa-mgmt-cluster.yaml \
      # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
   ```
   
   For an airgapped cluster create, follow [Preparation for airgapped deployments]({{< relref "../install#prepare-for-airgapped-deployments-optional" >}}) instructions, then type the following:

   ```bash
   eksctl anywhere create cluster
      --hardware-csv hardware.csv \
      -f $CLUSTER_NAME.yaml \
      --bundles-override ./eks-anywhere-downloads/bundle-release.yaml \
      # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
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

   Verify the test application in [Deploy test workload.]({{< relref "../../workloadmgmt/test-app" >}})

### Create separate workload clusters

Follow these steps if you want to use your initial cluster to create and manage separate workload clusters.

0. Set License Environment Variable (Optional)

   Add a license to any cluster for which you want to receive paid support. If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/clustermgmt/support/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

1. Generate a workload cluster config:
   ```bash
   CLUSTER_NAME=w01
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider tinkerbell > eksa-w01-cluster.yaml
   ```

   Refer to the initial config described earlier for the required and optional settings.
   Ensure workload cluster object names (`Cluster`, `TinkerbellDatacenterConfig`, `TinkerbellMachineConfig`, etc.) are distinct from management cluster object names. Keep the `tinkerbellIP` of workload cluster the same as `tinkerbellIP` of the management cluster.

1. Be sure to set the `managementCluster` field to identify the name of the management cluster.

   For example, the management cluster, _mgmt_ is defined for our workload cluster _w01_ as follows:

   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
     name: w01
   spec:
     managementCluster:
       name: mgmt
   ```

1. Create a workload cluster

   To create a new workload cluster from your management cluster run this command, identifying:

   * The workload cluster YAML file
   * The initial cluster's credentials (this causes the workload cluster to be managed from the management cluster)

   Create a workload cluster in one of the following ways:
   * **eksctl CLI**: To create a workload cluster with eksctl, run:

     ```bash
     eksctl anywhere create cluster \
         -f eksa-w01-cluster.yaml  \
         --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig \
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         # --hardware-csv <hardware.csv> \ # uncomment to add more hardware
         # --bundles-override ./eks-anywhere-downloads/bundle-release.yaml \ # uncomment for airgapped install
     ```
     As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

   * **kubectl CLI**: The cluster lifecycle feature lets you use kubectl to talks to the Kubernetes API to create a workload cluster. To use kubectl, run:
      ```bash
      kubectl apply -f eksa-w01-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
      ```

      To check the state of a cluster managed with the cluster lifecyle feature, use `kubectl` to show the cluster object with its status.
      
      The `status` field on the cluster object field holds information about the current state of the cluster.

      ```
      kubectl get clusters w01 -o yaml
      ```

      The cluster has been fully upgraded once the status of the `Ready` condition is marked `True`.
      See the [cluster status]({{< relref "../../clustermgmt/cluster-status" >}}) guide for more information.

   * **GitOps**: See [Manage separate workload clusters with GitOps]({{< relref "../../clustermgmt/cluster-flux.md#manage-separate-workload-clusters-using-gitops" >}})

   * **Terraform**: See [Manage separate workload clusters with Terraform]({{< relref "../../clustermgmt/cluster-terraform.md#manage-separate-workload-clusters-using-terraform" >}})

     >**NOTE**: For kubectl, GitOps and Terraform:
     > * The baremetal controller does not support scaling upgrades and Kubernetes version upgrades in the same request.
     > * While creating a new workload cluster if you need to add additional machines for the target workload cluster, run:
     >   ```
     >   eksctl anywhere generate hardware -z updated-hardware.csv > updated-hardware.yaml
     >   kubectl apply -f updated-hardware.yaml
     >   ```
     > * For creating multiple workload clusters, it is essential that the hardware labels and selectors defined for a given workload cluster are unique to that workload cluster. For instance, for an EKS Anywhere cluster named `eksa-workload1`, the hardware that is assigned for this cluster should have labels that are only going to be used for this cluster like `type=eksa-workload1-cp` and `type=eksa-workload1-worker`.
     Another workload cluster named `eksa-workload2` can have labels like `type=eksa-workload2-cp` and `type=eksa-workload2-worker`. Please note that even though labels can be arbitrary, they need to be unique for each workload cluster. Not specifying unique cluster labels can cause cluster creations to behave in unexpected ways which may lead to unsuccessful creations and unstable clusters.
     See the [hardware selectors]({{< relref "./bare-spec/#hardwareselector" >}}) section for more information

1. Check the workload cluster:

   You can now use the workload cluster as you would any Kubernetes cluster.
   Change your credentials to point to the new workload cluster (for example, `mgmt-w01`), then run the test application with:

   ```bash
   export CLUSTER_NAME=mgmt-w01
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section.]({{< relref "../../workloadmgmt/test-app" >}})

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `eksa-w02-cluster.yaml`), modifying resource names, and running the create cluster command again.


## Next steps:
* See the [Cluster management]({{< relref "../../clustermgmt" >}}) section for more information on common operational tasks like deleting the cluster.

* See the [Package management]({{< relref "../../packages" >}}) section for more information on post-creation curated packages installation.
