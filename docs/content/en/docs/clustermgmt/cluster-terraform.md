---
title: "Manage cluster with Terraform"
linkTitle: "Manage with Terraform"
weight: 70
aliases:
    /docs/tasks/cluster/cluster-terraform/
date: 2017-01-05
description: >
  Use Terraform to manage EKS Anywhere Clusters
---

>**_NOTE_**: Support for using Terraform to manage and modify an EKS Anywhere cluster is available for vSphere, Snow, Bare metal, Nutanix and CloudStack clusters.
>

## Using Terraform to manage an EKS Anywhere Cluster (Optional)
This guide explains how you can use Terraform to manage and modify an EKS Anywhere cluster. 
The guide is meant for illustrative purposes and is not a definitive approach to building production systems with Terraform and EKS Anywhere.

At its heart, EKS Anywhere is a set of Kubernetes CRDs, which define an EKS Anywhere cluster,
and a controller, which moves the cluster state to match these definitions.
These CRDs, and the EKS-A controller, live on the management cluster or
on a self-managed cluster.
We can manage a subset of the fields in the EKS Anywhere CRDs with any tool that can interact with the Kubernetes API, like `kubectl` or, in this case, the Terraform Kubernetes provider.

In this guide, we'll show you how to import your EKS Anywhere cluster into Terraform state and
how to scale your EKS Anywhere worker nodes using the Terraform Kubernetes provider.

### Prerequisites
- An existing EKS Anywhere cluster

- the latest version of [Terraform](https://www.terraform.io/downloads)

- the latest version of [tfk8s](https://github.com/jrhouston/tfk8s), a tool for converting Kubernetes manifest files to Terraform HCL


### Guide
0. Create an EKS-A management cluster, or a self-managed stand-alone cluster.
- if you already have an existing EKS-A cluster, skip this step.
- if you don't already have an existing EKS-A cluster, follow [the official instructions to create one](https://anywhere.eks.amazonaws.com/docs/getting-started/install/)

1. Set up the Terraform Kubernetes provider
   Make sure your KUBECONFIG environment variable is set
   ```bash
   export KUBECONFIG=/path/to/my/kubeconfig.kubeconfig
   ```

   Set an environment variable with your cluster name:
   ```bash
   export MY_EKSA_CLUSTER="myClusterName"
   ```

   ```bash
   cat << EOF > ./provider.tf
   provider "kubernetes" {
     config_path    = "${KUBECONFIG}"
   }
   EOF
   ```

2. Get  `tfk8s` and use it to convert your EKS Anywhere cluster Kubernetes manifest into Terraform HCL:
    - Install [tfk8s](https://github.com/jrhouston/tfk8s#install)
    - Convert the manifest into Terraform HCL:
   ```bash
   kubectl get cluster ${MY_EKSA_CLUSTER} -o yaml | tfk8s --strip -o ${MY_EKSA_CLUSTER}.tf
   ``` 

3. Configure the Terraform cluster resource definition generated in step 2
    - Set `metadata.generation` as a [computed field](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest#computed-fields). Add the following to your cluster resource configuration
   ```bash
   computed_fields = ["metadata.generated"]
   ```
    - Configure the field manager to [force reconcile managed resources](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest#field_manager). Add the following configuration block to your cluster resource:
   ```bash
   field_manager {
     force_conflicts = true
   }
   ```
    - Add the `namespace` to the `metadata` of the cluster
    - Remove the `generation` field from the `metadata` of the cluster
    - Your Terraform cluster resource should look similar to this:
    ```bash
    computed_fields = ["metadata.generated"]
    field_manager {
      force_conflicts = true
    }
    manifest = {
      "apiVersion" = "anywhere.eks.amazonaws.com/v1alpha1"
      "kind" = "Cluster"
      "metadata" = {
        "name" = "MyClusterName"
        "namespace" = "default"
    }
    ```

4. Import your EKS Anywhere cluster into terraform state:
   ```bash
   terraform init
   terraform import kubernetes_manifest.cluster_${MY_EKSA_CLUSTER} "apiVersion=anywhere.eks.amazonaws.com/v1alpha1,kind=Cluster,namespace=default,name=${MY_EKSA_CLUSTER}"
   ```

   After you `import` your cluster, you will need to run `terraform apply` one time to ensure that the `manifest` field of your cluster resource is in-sync.
   This will not change the state of your cluster, but is a required step after the initial import.
   The `manifest` field stores the contents of the associated kubernetes manifest, while the `object` field stores the actual state of the resource.

5. Modify Your Cluster using Terraform
    - Modify the `count` value of one of your `workerNodeGroupConfigurations`, or another mutable field, in the configuration stored in `${MY_EKSA_CLUSTER}.tf` file.
    - Check the expected diff between your cluster state and the modified local state via `terraform plan`

   You should see in the output that the worker node group configuration count field (or whichever field you chose to modify) will be modified by Terraform.

6. Now, actually change your cluster to match the local configuration:
   ```bash
   terraform apply
   ```

7. Observe the change to your cluster. For example:
   ```bash
   kubectl get nodes
   ```

## Manage separate workload clusters using Terraform

Follow these steps if you want to use your initial cluster to create and manage separate workload clusters via Terraform.

> **NOTE**: If you choose to manage your cluster using Terraform, do not use `kubectl` to edit your cluster objects as this can lead to field manager conflicts.

### Prerequisites
- An existing EKS Anywhere cluster imported into Terraform state.
  If your existing cluster is not yet imported, see this [guide.]({{< relref "#guide" >}}).
- A cluster configuration file for your new workload cluster.
  
### Create cluster using Terraform

1. Create the new cluster configuration Terraform file. 
   ```bash
      tfk8s -f new-workload-cluster.yaml -o new-workload-cluster.tf
   ```

   > **NOTE**: Specify the `namespace` for all EKS Anywhere objects when you are using Terraform to manage your clusters (even for the `default` namespace, use `"namespace" = "default"` on those objects).
   > 
   > Ensure workload cluster object names are distinct from management cluster object names. Be sure to set the `managementCluster` field to identify the name of the management cluster.
   
2. Ensure that this new Terraform workload cluster configuration exists in the same directory as the management cluster Terraform files.
      ```
      my/terraform/config/path
      ├── management-cluster.tf
      ├── new-workload-cluster.tf
      ├── provider.tf
      ├──  ... 
      └──
      ```

3. Verify the changes to be applied:
   ```bash
   terraform plan
   ```

4. If the plan looks as expected, apply those changes to create the new cluster resources:
   ```bash
   terraform apply
   ```
   
5.  You can list the workload clusters managed by the management cluster.
      ```bash
      export KUBECONFIG=${PWD}/${MGMT_CLUSTER_NAME}/${MGMT_CLUSTER_NAME}-eks-a-cluster.kubeconfig
      kubectl get clusters
      ```

6. Check the state of a cluster using `kubectl` to show the cluster object with its status.
   
   The `status` field on the cluster object field holds information about the current state of the cluster.

   ```
   kubectl get clusters w01 -o yaml
   ```

   The cluster has been fully upgraded once the status of the `Ready` condition is marked `True`.
   See the [cluster status]({{< relref "./cluster-status" >}}) guide for more information.
   
7. The kubeconfig for your new cluster is stored as a secret on the management cluster.
   You can get the workload cluster credentials and run the test application on your new workload cluster as follows:
   ```bash
   kubectl get secret -n eksa-system w01-kubeconfig -o jsonpath='{.data.value}' | base64 --decode > w01.kubeconfig
   export KUBECONFIG=w01.kubeconfig
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```
   
### Upgrade cluster using Terraform
   
1. To upgrade a workload cluster using Terraform, modify the desired fields in the Terraform resource file.
   As an example, to upgrade a cluster with version 1.24 to 1.25 you would modify your Terraform cluster resource:
   ```bash
    manifest = {
      "apiVersion" = "anywhere.eks.amazonaws.com/v1alpha1"
      "kind" = "Cluster"
      "metadata" = {
        "name" = "MyClusterName"
        "namespace" = "default"
      }
      "spec" = {
        "kubernetesVersion" = "1.25"
         ...
         ...
      }
    ```
   >**_NOTE:_** If you have a custom machine image for your nodes you may also need to update your MachineConfig with a new `template`.
   
2. Apply those changes:
   ```bash
   terraform apply
   ```

For a comprehensive list of upgradeable fields for VSphere, Snow, and Nutanix, see the [upgradeable attributes section]({{< relref "./cluster-upgrades/vsphere-and-cloudstack-upgrades.md#upgradeable-cluster-attributes" >}}).
 
### Delete cluster using Terraform

1. To delete a workload cluster using Terraform, you will need the name of the Terraform cluster resource.
   This can be found on the first line of your cluster resource definition.
   ```bash
   terraform destroy --target kubernetes_manifest.cluster_w01
   ```

### Appendix
Terraform K8s Provider https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs

tfk8s https://github.com/jrhouston/tfk8s
