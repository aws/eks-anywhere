---
title: "Create Snow production cluster" 
linkTitle: "Snow cluster" 
weight: 40
description: >
  Create a production-quality cluster on AWS Snow
---

EKS Anywhere supports an AWS Snow provider for production grade EKS Anywhere deployments.

This document walks you through setting up EKS Anywhere on Snow as a standalone, self-managed cluster or combined set of management/workload clusters.
See [Cluster topologies]({{< relref "../../concepts/cluster-topologies" >}}) for details.

## Prerequisite checklist

EKS Anywhere on Snow needs:

* Certain pre-steps to complete before interacting with a Snowball device. See [Actions to complete before ordering a Snowball Edge device for Amazon EKS Anywhere](https://docs.aws.amazon.com/snowball/latest/developer-guide/whatisedge.html).
* EKS Anywhere enabled Snowball devices. See [Ordering a Snowball Edge device for use with Amazon EKS Anywhere](https://docs.aws.amazon.com/snowball/latest/developer-guide/whatisedge.html) for ordering experience through the AWS Snow Family console.
* To be run on an Admin instance in a Snowball Edge device. See [Configuring and starting Amazon EKS Anywhere on Snowball Edge devices](https://docs.aws.amazon.com/snowball/latest/developer-guide/whatisedge.html) for setting up the devices, launching the Admin instance, fetching and copying the device credentials to the Admin instance for `eksctl` CLI to consume.

Also, see the [Ports and protocols]({{< relref "/docs/reference/ports.md" >}}) page for information on ports that need to be accessible from control plane, worker, and Admin machines.

## Steps

The following steps are divided into two sections:

* Create an initial cluster (used as a management or standalone cluster)
* Create zero or more workload clusters from the management cluster

### Create an initial cluster

Follow these steps to create an EKS Anywhere cluster that can be used either as a management cluster or as a standalone cluster (for running workloads itself).

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Set an environment variables for your cluster name
   
   ```bash
   export CLUSTER_NAME=mgmt
   ```

1. Generate a cluster config file for your Snow provider

   ```bash
   eksctl anywhere generate clusterconfig $CLUSTER_NAME --provider snow > eksa-mgmt-cluster.yaml
   ```

1. Optionally import images to private registry
   
   This optional step imports EKS Anywhere artifacts and release bundle to a local registry. This is required for air-gapped installation.
   
   * [Configuring Amazon EKS Anywhere for disconnected operation](https://docs.aws.amazon.com/snowball/latest/developer-guide/whatisedge.html) shows AWS examples of selecting and building a private registry in a Snowball Edge device.
   * For air-gapped scenario, run the `import images` with `--input` and `--bundles` arguments pointing to the artifacts and bundle release files that pre-exist in the Admin instance.
   * Refer to the [Registry Mirror configuration]({{< relref "../../reference/clusterspec/optional/registrymirror" >}}) for more information about using private registry.

   ```bash
   eksctl anywhere import images \
      --input /usr/lib/eks-a/artifacts/artifacts.tar.gz \
      --bundles /usr/lib/eks-a/manifests/bundle-release.yaml \
      --registry $PRIVATE_REGISTRY_ENDPOINT \
      --insecure=true
   ```

1. Modify the cluster config (`eksa-mgmt-cluster.yaml`) as follows:
   * Refer to the [Snow configuration]({{< relref "../../reference/clusterspec/snow" >}}) for information on configuring this cluster config for a Snow provider.
   * Add [Optional]({{< relref "/docs/reference/clusterspec/optional/" >}}) configuration settings as needed.

1. Set License Environment Variable

   If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

1. Configure Curated Packages

   The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/). Cluster creation will succeed if authentication is not set up, but some warnings may be generated. Detailed package configurations can be found [here]({{< relref "../../tasks/packages" >}}).

   If you are going to use packages, set up authentication. These credentials should have [limited capabilities]({{< relref "../../tasks/packages/#setup-authentication-to-use-curated-packages" >}}):
   ```bash
   export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
   export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
   export EKSA_AWS_REGION="us-west-2" 
   ```

   *Curated packages are not yet supported on air-gapped installation.*

1. Set Credential Environment Variables

   Before you create the initial cluster, you will need to use the `credentials` and `ca-bundles` files that are in the Admin instance, and export these environment variables for your AWS Snowball device credentials.
Make sure you use single quotes around the values so that your shell does not interpret the values:
   
   ```bash
   export EKSA_AWS_CREDENTIALS_FILE='/PATH/TO/CREDENTIALS/FILE'
   export EKSA_AWS_CA_BUNDLES_FILE='/PATH/TO/CABUNDLES/FILE'
   ```

   After you have created your `eksa-mgmt-cluster.yaml` and set your credential environment variables, you will be ready to create the cluster.

1. Create cluster

   a. For none air-gapped environment
   ```bash
   eksctl anywhere create cluster \
      -f eksa-mgmt-cluster.yaml
   ```

   b. For air-gapped environment
   ```bash
   eksctl anywhere create cluster \
      -f eksa-mgmt-cluster.yaml \
      --bundles-override /usr/lib/eks-a/manifests/bundle-release.yaml
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
   NAMESPACE    NAME                        CLUSTER  NODENAME                    PROVIDERID                                       PHASE    AGE    VERSION
   eksa-system  mgmt-etcd-dsxb5             mgmt                                 aws-snow:///192.168.1.231/s.i-8b0b0631da3b8d9e4  Running  4m59s  
   eksa-system  mgmt-md-0-7b7c69cf94-99sll  mgmt     mgmt-md-0-1-58nng           aws-snow:///192.168.1.231/s.i-8ebf6b58a58e47531  Running  4m58s  v1.24.9-eks-1-24-7
   eksa-system  mgmt-srrt8                  mgmt     mgmt-control-plane-1-xs4t9  aws-snow:///192.168.1.231/s.i-8414c7fcabcf3d7c1  Running  4m58s  v1.24.9-eks-1-24-7
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

### Create separate workload clusters

Follow these steps if you want to use your initial cluster to create and manage separate workload clusters.

1. Generate a workload cluster config:
   ```bash
   CLUSTER_NAME=w01
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider snow > eksa-w01-cluster.yaml
   ```

   Refer to the initial config described earlier for the required and optional settings.

   Ensure workload cluster object names (`Cluster`, `SnowDatacenterConfig`, `SnowMachineConfig`, etc.) are distinct from management cluster object names. Be sure to set the `managementCluster` field to identify the name of the management cluster.

1. Set License Environment Variable

   Add a license to any cluster for which you want to receive paid support. If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

1. Create a workload cluster in one of the following ways:
   
   * **GitOps**: Recommended for more permanent cluster configurations.
     1. Clone your git repo and add the new cluster specification. Be sure to follow the directory structure defined on [Manage cluster with GitOps]({{< relref "/docs/tasks/cluster/cluster-flux" >}}):

      ```
      clusters/<management-cluster-name>/$CLUSTER_NAME/eksa-system/eksa-cluster.yaml
      ```

      2. Commit the file to your git repository
         ```bash
         git add clusters/<management-cluster-name>/$CLUSTER_NAME/eksa-system/eksa-cluster.yaml
         git commit -m 'Creating new workload cluster'
         git push origin main
         ```
         
      3. The flux controller will automatically make the required changes.

     > **NOTE**: Specify the `namespace` for all EKS Anywhere objects when you are using GitOps to create new workload clusters (even for the `default` namespace, use `namespace: default` on those objects).
     >
     > `snowDatacenterConfig.spec.identityRef` and a Snow bootstrap credentials secret need to be specified when provisionig a cluster through `GitOps`, as EKS Anywhere Cluster Controller will not create a Snow bootstrap credentials secret like `eksctl CLI` does when field is empty.
     >
     > Make sure there is a `kustomization.yaml` file under the namespace directory for the management cluster. Creating a Gitops enabled management cluster with `eksctl` should create the `kustomization.yaml` file automatically.
     
   See [Manage cluster with GitOps]({{< relref "/docs/tasks/cluster/cluster-flux" >}}) for more details.
   
   * **eksctl CLI**: Useful for temporary cluster configurations. To create a workload cluster with `eksctl`, run:
      ```bash
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

   Verify the test application in the [deploy test application section.]({{< relref "../../tasks/workload/test-app" >}})

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `eksa-w02-cluster.yaml`), modifying resource names, and running the create cluster command again.


## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
