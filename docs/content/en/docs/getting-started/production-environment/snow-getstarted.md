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

* Certain pre-steps to complete before interacting with a Snowball device. See [Actions to complete before ordering a Snowball Edge device for Amazon EKS Anywhere](https://docs.aws.amazon.com/snowball/latest/developer-guide/eksa-gettingstarted.html).
* EKS Anywhere enabled Snowball devices. See [Ordering a Snowball Edge device for use with Amazon EKS Anywhere](https://docs.aws.amazon.com/snowball/latest/developer-guide/order-sbe.html) for ordering experience through the AWS Snow Family console.
* To be run on an Admin instance in a Snowball Edge device. See [Configuring and starting Amazon EKS Anywhere on Snowball Edge devices](https://docs.aws.amazon.com/snowball/latest/developer-guide/eksa-configuration.html) for setting up the devices, launching the Admin instance, fetching and copying the device credentials to the Admin instance for `eksctl` CLI to consume.

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
   
   * [Configuring Amazon EKS Anywhere for disconnected operation](https://docs.aws.amazon.com/snowball/latest/developer-guide/configure-disconnected.html) shows AWS examples of selecting and building a private registry in a Snowball Edge device.
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

   >**NOTE**: Ensure workload cluster object names (`Cluster`, `SnowDatacenterConfig`, `SnowMachineConfig`, etc.) are distinct from management cluster object names.

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

1. Set License Environment Variable

   Add a license to any cluster for which you want to receive paid support. If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

1. Create a workload cluster in one of the following ways:
   
   * **GitOps**: See [Manage separate workload clusters with GitOps]({{< relref "../../tasks/cluster/cluster-flux.md#manage-separate-workload-clusters-using-gitops" >}})

   * **Terraform**: See [Manage separate workload clusters with Terraform]({{< relref "../../tasks/cluster/cluster-terraform.md#manage-separate-workload-clusters-using-terraform" >}})

     > **NOTE**: `snowDatacenterConfig.spec.identityRef` and a Snow bootstrap credentials secret need to be specified when provisioning a cluster through `GitOps` or `Terraform`, as EKS Anywhere Cluster Controller will not create a Snow bootstrap credentials secret like `eksctl CLI` does when field is empty.
     >
     > `snowMachineConfig.spec.sshKeyName` must be specified to SSH into your nodes when provisioning a cluster through `GitOps` or `Terraform`, as the EKS Anywhere Cluster Controller will not generate the keys like `eksctl CLI` does when the field is empty.

    * **eksctl CLI**: To create a workload cluster with `eksctl`, run:
      ```bash
      eksctl anywhere create cluster \
          -f eksa-w01-cluster.yaml  \
          --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
      ```
      As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

1. Check the workload cluster:

   You can now use the workload cluster as you would any Kubernetes cluster.
   * If your workload cluster was created with `eksctl`,
     change your credentials to point to the new workload cluster (for example, `w01`), then run the test application with:

      ```bash
      export CLUSTER_NAME=w01
      export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
      kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
      ```

   * If your workload cluster was created with GitOps or Terraform, the kubeconfig for your new cluster is stored as a secret on the management cluster.
     You can get credentials and run the test application as follows:
      ```bash
      kubectl get secret -n eksa-system w01-kubeconfig -o jsonpath=‘{.data.value}' | base64 —decode > w01.kubeconfig
      export KUBECONFIG=w01.kubeconfig
      kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
      ```

   Verify the test application in the [deploy test application section.]({{< relref "../../tasks/workload/test-app" >}})

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `eksa-w02-cluster.yaml`), modifying resource names, and running the create cluster command again.


## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
