---
title: "Create production cluster"
weight: 40
---

EKS Anywhere supports a vSphere provider for production grade EKS Anywhere deployments.
The purpose of this doc is to walk you through getting set-up with EKS Anywhere (EKS-A).
EKS-A allows you to provision and manage Amazon EKS on your own infrastructure.

## Prerequisite Checklist

EKS Anywhere needs to be run on an administrative machine that has certain [machine
requirements]({{< relref "../install" >}}).
An EKS Anywhere deployment will also require the availability of certain
[resources from your VMware vSphere deployment]({{< relref "../../reference/clusterspec/vsphere" >}}).

To run EKS Anywhere, you will need the following:

## Steps

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate a cluster config
   ```bash
   CLUSTER_NAME=prod
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider vsphere > eksa-cluster.yaml
   ```

    A production grade EKS-A cluster should be made with at least three control plane nodes and three worker nodes
    for high availability and rolling upgrades.:
    ```
      controlPlaneConfiguration:
        count: 3
        endpoint:
          host: 198.18.100.79
        machineGroupRef:
          kind: VSphereMachineConfig
          name: prod-control-plane
      workerNodeGroupConfigurations:
      - count: 3
        machineGroupRef:
          kind: VSphereMachineConfig
          name: prod-data-plane
    ```

    Further information about the values in the `eksa-cluster.yaml` can be found in the [cluster specification
    reference]({{< relref "../../reference/clusterspec/vsphere.md" >}})

1. Set Credential Environment Variables

   Before you create a cluster, you will need to set and export these environment variables for your vSphere user
   name and password. Make sure you use single quotes around the values so that your shell does not interpret the values:
   
   ```bash
   export EKSA_VSPHERE_USERNAME='billy'
   export EKSA_VSPHERE_PASSWORD='t0p$ecret'
   ```

   EKS Anywhere clusters function as both workload and management clusters.
   Management clusters are responsible for the lifecycle of workload clusters (i.e. create, upgrade, and delete clusters), while workload clusters run user applications.

   Future versions of EKS Anywhere will enable users to create a dedicated management cluster that will govern multiple workload clusters allowing for segmentation of different cluster types.

1. Create a cluster

   After you have created your `eks-cluster.yaml` and set your credential environment variables, you will be ready
   to create a cluster:
   ```bash
   eksctl anywhere create cluster -f eksa-cluster.yaml
   ```
   Example command output
   ```
    Performing setup and validations
    ✅ Connected to server
    ✅ Authenticated to vSphere
    ✅ datacenter validated
    ✅ datastore validated
    ✅ folder validated
    ✅ resource pool validated
    ✅ network validated
    ✅ Template validated
    ✅ validation succeeded	{"validation": "vsphere Provider setup is valid"}
    Creating new bootstrap cluster
    Installing cluster-api providers on bootstrap cluster
    Provider specific setup
    Creating new workload cluster
    Installing networking on workload cluster
    Installing storage class on workload cluster
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
    capi-kubeadm-bootstrap-system       Active   4m26s
    capi-kubeadm-control-plane-system   Active   4m19s
    capi-system                         Active   4m36s
    capi-webhook-system                 Active   4m46s
    capv-system                         Active   4m4s
    cert-manager                        Active   5m40s
    default                             Active   12m
    eksa-system                         Active   3m7s
    kube-node-lease                     Active   12m
    kube-public                         Active   12m
    kube-system                         Active   12m
   ```

   You can now use the cluster like you would any Kubernetes cluster.
   Deploy the test application with:

   ```bash
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section]({{< relref "../../tasks/workload/test-app" >}}).
   See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section with more information on common operational tasks like scaling and deleting the cluster.
