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

1. Generate a curated-packages config
   {{% alert title="Note" color="primary" %}}
   * It is *optional* to install the curated packages as part of the cluster creation and `eksctl anywhere` version should be `v0.9.0` or later.
   * Post-creation installation and detailed package configurations can be found [here.]({{< relref "../../tasks/packages" >}})
   {{% /alert %}}
   Example shows how to install two packages `flux` and `harbor` from the [curated package list]({{< relref "../../reference/packagespec" >}}).
   ```bash
   eksctl anywhere generate package flux harbor -d .
   ```

1. Create a cluster

   ```bash
   # Create a cluster without curated packages installation
   eksctl anywhere create cluster -f $CLUSTER_NAME.yaml
   # Create a cluster with curated packages installation
   eksctl anywhere create cluster -f $CLUSTER_NAME.yaml --install-packages ./curated-packages/
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
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section with more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section with more information on post-creation curated packages installation.
