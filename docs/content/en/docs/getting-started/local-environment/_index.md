---
title: Create local cluster
weight: 20
---

## EKS Anywhere docker provider deployments

EKS Anywhere supports a Docker provider for *development and testing use cases only.* 
This allows you to try EKS Anywhere on your local system before deploying to a supported provider.

To install the EKS Anywhere binaries and see system requirements please follow the [installation guide]({{< relref "../install" >}}).

## Steps

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate a cluster config
   ```bash
   CLUSTER_NAME=dev-cluster
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > $CLUSTER_NAME.yaml
   ```

   The command above creates a file named eksa-cluster.yaml with the contents below in the path where it is executed.
   The configuration specification is divided into two sections:

   * Cluster
   * DockerDatacenterConfig

   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
   name: dev-cluster
   spec:
   clusterNetwork:
      cniConfig:
         cilium: {}
      pods:
         cidrBlocks:
         - 192.168.0.0/16
      services:
         cidrBlocks:
         - 10.96.0.0/12
   controlPlaneConfiguration:
      count: 1
   datacenterRef:
      kind: DockerDatacenterConfig
      name: dev-cluster
   externalEtcdConfiguration:
      count: 1
   kubernetesVersion: "1.21"
   managementCluster:
      name: dev-cluster
   workerNodeGroupConfigurations:
   - count: 1
      name: md-0
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: DockerDatacenterConfig
   metadata:
   name: dev-cluster
   spec: {}
   ```

   * Apart from the base configuration, you can add additional optional configuration to enable supported features:
      * [OIDC](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/oidc/) 
      * [etcd](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/etcd/)
      * [proxy](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/proxy/)
      * [gitops](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/gitops/)

1. Create Cluster: Create your cluster either with or without curated packages:

   - Cluster creation without curated packages installation
      ```bash
      eksctl anywhere create cluster -f $CLUSTER_NAME.yaml
      ```
      Example command output
      ```
      Performing setup and validations
      ✅ validation succeeded {"validation": "docker Provider setup is valid"}
      Creating new bootstrap cluster
      Installing cluster-api providers on bootstrap cluster
      Provider specific setup
      Creating new workload cluster
      Installing networking on workload cluster
      Installing cluster-api providers on workload cluster
      Moving cluster management from bootstrap to workload cluster
      Installing EKS-A custom components (CRD and controller) on workload cluster
      Creating EKS-A CRDs instances on workload cluster
      Installing AddonManager and GitOps Toolkit on workload cluster
      GitOps field not specified, bootstrap flux skipped
      Deleting bootstrap cluster
      🎉 Cluster created!
      ```
   - Cluster creation with optional curated packages

     {{% alert title="Note" color="primary" %}}
   * It is *optional* to install curated packages as part of the cluster creation.
   * `eksctl anywhere version` version should be later than `v0.9.0`.
   * If including curated packages during cluster creation, please set the environment variable: `export CURATED_PACKAGES_SUPPORT=true`
   * Post-creation installation and detailed package configurations can be found [here.]({{< relref "../../tasks/packages" >}})
     {{% /alert %}}

      * Discover curated-packages to install
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

      * Create a cluster

         ```bash
         # Create a cluster with curated packages installation
         eksctl anywhere create cluster -f $CLUSTER_NAME.yaml --install-packages packages.yaml
         ```
         Example command output
         ```
         Performing setup and validations
         ✅ validation succeeded {"validation": "docker Provider setup is valid"}
         Creating new bootstrap cluster
         Installing cluster-api providers on bootstrap cluster
         Provider specific setup
         Creating new workload cluster
         Installing networking on workload cluster
         Installing cluster-api providers on workload cluster
         Moving cluster management from bootstrap to workload cluster
         Installing EKS-A custom components (CRD and controller) on workload cluster
         Creating EKS-A CRDs instances on workload cluster
         Installing AddonManager and GitOps Toolkit on workload cluster
         GitOps field not specified, bootstrap flux skipped
         Deleting bootstrap cluster
         🎉 Cluster created!
         ----------------------------------------------------------------------------------------------------------------
         The EKS Anywhere package controller and the EKS Anywhere Curated Packages
         (referred to as “features”) are provided as “preview features” subject to the AWS Service Terms,
         (including Section 2 (Betas and Previews)) of the same. During the EKS Anywhere Curated Packages Public Preview,
         the AWS Service Terms are extended to provide customers access to these features free of charge.
         These features will be subject to a service charge and fee structure at ”General Availability“ of the features.
         ----------------------------------------------------------------------------------------------------------------
         Installing curated packages controller on workload cluster
         package.packages.eks.amazonaws.com/my-harbor created
         ```

1. Use the cluster

   Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory

   ```bash
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl get ns
   ```
   Example command output
   ```
   NAME                                STATUS   AGE
   capd-system                         Active   21m
   capi-kubeadm-bootstrap-system       Active   21m
   capi-kubeadm-control-plane-system   Active   21m
   capi-system                         Active   21m
   capi-webhook-system                 Active   21m
   cert-manager                        Active   22m
   default                             Active   23m
   eksa-system                         Active   20m
   kube-node-lease                     Active   23m
   kube-public                         Active   23m
   kube-system                         Active   23m
   ```

   You can now use the cluster like you would any Kubernetes cluster.
   Deploy the test application with:

   ```bash
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section]({{< relref "../../tasks/workload/test-app" >}}).

## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
