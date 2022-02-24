## Using Terraform to manage an EKS Anywhere Cluster
This is a minimal guide that lays out how you can use Terraform to manage and modify an EKS Anywhere cluster. The guide is meant for illustrative purposes only and is not meant as a definitive approach to building production systems with Terraform and EKS Anywhere.

At its heart, EKS Anywhere is a set of Kubernetes CRDs which define an EKS Anywhere cluster, 
and a controller which moves the cluster to match these definitions. 
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
   ```bash
   export KUBECONFIG=/path/to/my/kubeconfig.kubeconfig
   cat << EOF >> ./provider.tf
   provider "kubernetes" {
     config_path    = "${KUBECONFIG}"
   }
   EOF
   ```

2. Get  `tfk8s` and use it to convert your EKS Anywhere cluster Kubernetes manifest into Terraform HCL:
   - Install [tfk8s](https://github.com/jrhouston/tfk8s#install)
   - Convert the manifest into Terraform HCL:
   ```bash
   export MY_EKSA_CLUSTER="myClusterName"
   kubectl get cluster ${MY_EKSA_CLUSTER} -o yaml | tfk8s --strip -o ${MY_EKSA_CLUSTER}.tf
   ``` 

3. Configure the Terraform cluster resource definition generated in step 2
   - Add the `namespace` `default` to the `metadata` of the cluster
   - Remove the `generation` field from the `metadata` of the cluster

   - Your cluster manifest metadata should look like this (`generation` may be different):
    ```bash
    manifest = {
      "apiVersion" = "anywhere.eks.amazonaws.com/v1alpha1"
      "kind" = "Cluster"
      "metadata" = {
        "name" = "MyClusterName"
        "namespace" = "default"
    }
    ```

   Set `metadata.generation` as a [computed field](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest#computed-fields)
   - Add the following to your cluster resource configuration
   computed_fields = ["metadata.generated"]

   Configure the field manager to [force reconcile managed resources](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest#field_manager)
   - Add the following configuration block to your cluster resource:
   ```bash
   field_manager {
     force_conflicts = true
   }
   ```

4. Import your EKS Anywhere cluster into terraform state:
   ```bash
   export MY_EKSA_CLUSTER="myClusterName"
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

### Appendix
Terraform K8s Provider https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs

tfk8s https://github.com/jrhouston/tfk8s
