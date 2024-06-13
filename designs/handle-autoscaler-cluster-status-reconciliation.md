# Handle cluster status reconciliation in EKS Anywhere controller for Autoscaling Configuration

## Problem Statement

When a customer configures the [autoscaling](https://anywhere.eks.amazonaws.com/docs/getting-started/optional/autoscaling/) configuration for any of their worker node group, the number of worker nodes created in the cluster will be handled by the [cluster-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) in order to ensure that all pods have a place to run and there are no unneeded nodes. The cluster autoscaler also updates the replicas in the corresponding machine deployment object to match the actual number of machines provisioned. When the EKS-A controller reconciles the cluster status, it sees that the expected count of worker nodes does not match the observed count and marks the `WorkersReady` condition to False with the following message `Scaling down worker nodes, 1 expected (10 actual)`  This is because it gets the expected count from the worker node groups count in the cluster spec which is set by the customer during cluster creation or upgrade whereas the actual replicas are handled by the autoscaler. This doc discusses various options to fix this issue with the EKS Anywhere controller cluster status reconciliation for worker node groups with autoscaling configured.

## Overview of Solution

#### Handling cluster status reconciliation

Update the [totalExpected](https://github.com/aws/eks-anywhere/blob/a2a19920f4b7b54f6bc21f608ee5ecd5c6f0c45b/pkg/controller/clusters/status.go#L202) count of worker nodes to be equal to the count of worker nodes specified in the cluster spec *only* for worker node groups without autoscaling configured. For worker node groups which are configured with autoscaling, we include another validation for the `workersReady` condition which checks that the number of replicas lies between the minCount and maxCount specified in the autoscaler configuration. This validation will be done only after all the existing validations are done for worker node groups without autoscaling configured.

#### Handling cluster spec updates

When cluster spec is applied during cluster create/upgrade, we will not set the replicas in the md template for the worker node groups which have autoscaling configured. It will be defaulted to the minCount specified in the autoscaling configuration for new md objects during cluster creation whereas for cluster upgrades, it will be the same as the old md object’s replicas field value.

**Pros:**

* Removing the dependency on worker node group count for cluster creation too
* Worker node count is ignored which is what we want because autoscaler should handle it

**Cons:**

* Source of truth for worker nodes count would be md replicas which is not coming from an object that we own

#### Testing

E2E test will be added to test the worker node groups configured with autoscaler for cluster upgrades

#### Documentation

We need to explicitly document that the count will be ignored for all the worker node groups configuration which have autoscaling configured in the cluster spec for both cluster creation as well as upgrade.

## Alternate Solutions Considered

Here, option number corresponds to options for cluster status reconciliation (2 options)
Here, option letter corresponds to options for cluster spec updates (3 options)

### **Option 1a**

#### Handling cluster status reconciliation

For each worker node group, if the [count](https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-spec/#workernodegroupconfigurationscount-required) in the worker node group configuration for the cluster object is not equal to the replicas field in the machine deployment object, update the count to match it to the number of md replicas. This will be implemented in the MachineDeploymentReconciler in the EKS Anywhere controller.

#### Handling cluster spec updates

When cluster spec is applied during cluster create/upgrade, we will not set the replicas in the md template for the worker node groups which have autoscaling configured. It will be defaulted to the minCount specified in the autoscaling configuration for new md objects during cluster creation whereas for cluster upgrades, it will be the same as the old md object’s replicas field value.

### **Option 1b:**

#### Handling cluster status reconciliation

For each worker node group, if the [count](https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-spec/#workernodegroupconfigurationscount-required) in the worker node group configuration for the cluster object is not equal to the replicas field in the machine deployment object, update the count to match it to the number of md replicas. This will be implemented in the MachineDeploymentReconciler in the EKS Anywhere controller.

#### Handling cluster spec updates

We will deny any updates to the worker node count in the webhook if the autoscaling configuration is set. This will ensure that the md object is not re-applied by the controller to avoid updating the replicas field which should be handled by the autoscaler only.

### **Option 1c:**

#### Handling cluster status reconciliation

For each worker node group, if the [count](https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-spec/#workernodegroupconfigurationscount-required) in the worker node group configuration for the cluster object is not equal to the replicas field in the machine deployment object, update the count to match it to the number of md replicas. This will be implemented in the _MachineDeploymentReconciler_ in the EKS Anywhere controller.

#### Handling cluster spec updates

For each worker node group, if the [count](https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-spec/#workernodegroupconfigurationscount-required) in the worker node group configuration for the cluster object is not equal to the replicas field in the machine deployment object, update the count to match it to the number of md replicas. This will be implemented in the _ClusterReconciler_ in the EKS Anywhere controller.

### Option 2a:

#### Handling cluster status reconciliation

Update the [totalExpected](https://github.com/aws/eks-anywhere/blob/a2a19920f4b7b54f6bc21f608ee5ecd5c6f0c45b/pkg/controller/clusters/status.go#L202) count of worker nodes to be equal to the count of worker nodes specified in the cluster spec *only* for worker node groups without autoscaling configured. For worker node groups which are configured with autoscaling, we include another validation for the `workersReady` condition which checks that the number of replicas lies between the minCount and maxCount specified in the autoscaler configuration. This validation will be done only after all the existing validations are done for worker node groups without autoscaling configured.

#### Handling cluster spec updates

We will deny any updates to the worker node count in the webhook if the autoscaling configuration is set. This will ensure that the md object is not re-applied by the controller to avoid updating the replicas field which should be handled by the autoscaler only.

Proposed solution is better than this option because it does not force the user to remove their autoscaling configuration if they decide to make an update to the worker nodes count and not rely on the autoscaler anymore.

### Option 2b:

#### Handling cluster status reconciliation

Update the [totalExpected](https://github.com/aws/eks-anywhere/blob/a2a19920f4b7b54f6bc21f608ee5ecd5c6f0c45b/pkg/controller/clusters/status.go#L202) count of worker nodes to be equal to the count of worker nodes specified in the cluster spec *only* for worker node groups without autoscaling configured. For worker node groups which are configured with autoscaling, we include another validation for the `workersReady` condition which checks that the number of replicas lies between the minCount and maxCount specified in the autoscaler configuration. This validation will be done only after all the existing validations are done for worker node groups without autoscaling configured.

#### Handling cluster spec updates

For each worker node group, if the [count](https://anywhere.eks.amazonaws.com/docs/getting-started/vsphere/vsphere-spec/#workernodegroupconfigurationscount-required) in the worker node group configuration for the cluster object is not equal to the replicas field in the machine deployment object, update the count to match it to the number of md replicas. This will be implemented in the _ClusterReconciler_ in the EKS Anywhere controller.

Option 1c is better than this option because it can use the same function for updating the count in both MachineDeploymentReconciler as well as ClusterReconciler and also it does not have to change any logic for the cluster status reconciliation.
