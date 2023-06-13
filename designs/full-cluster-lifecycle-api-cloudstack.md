# Cluster Lifecycle API CloudStack

## Introduction

### Problem

Currently the only way to manage EKS Anywhere CloudStack cluster is by using CLI. We need to support native K8s experience on CloudStack cluster lifecycle.

### Goals and Objectives

* Create/Upgrade/Delete CloudStack workload clusters using: kubectl apply -f cluster.yaml
* Create/Update/Upgrade/Delete CloudStack workload clusters using GitOps/Terraform

#### Not in Scope

* Create/Upgrade/Delete self-managed clusters with API

#### Limitation

* We currently only supports k8s version < 1.24 on CloudStack due to current CAPC version.

## Overview of Solution

The EKS-A controller running in management cluster fully manages workload clusters, including reconciling EKS-A CRDs and installing the CNI in the workload cluster. Though with different machine provider, the CloudStack cluster reconciling process watches and reconciles the same CAPI objects and EKS-A cluster object as vSphere, the reconciliation flow is similar to [vSphere reconciler](images/cluster_reconcile.png) which watches resources and use an event handler to enqueue reconcile requests in response to those events.

In order to maintain same level of validations logic we currently run in CLI, we'll import validations into data validation and run time validation.
* Data validation: Kubernetes offers validation webhook for CRDs which runs data validation before accepted by kube-api server. Data validations should be light and fast.
* Run time validation: some validations require calling the CloudStack API, like if provided availability zone/account exits, which is too heavy for the webhook. We will implement those validations in a separate datacenter reconciler, and block the reconciliation process on failure, which is similar to [vSphere datacenter reconciler](https://github.com/aws/eks-anywhere/blob/main/designs/full-cluster-lifecycle-api.md?plain=1#L82)

### Defaults

As part of the CLI logic, we set default values to some spec if they're missing from what customers provide, like [datacenter availability zone](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/api/v1alpha1/cloudstackdatacenterconfig_types.go#L216), which can be leveraged in mutation webhook. 

We don't want to modify the spec that customers have already specified. [Control Plane Host Port](https://github.com/aws/eks-anywhere/blob/3c1fd0ff732641ed02137213863942403f59c320/pkg/providers/cloudstack/validator.go#L211) could be missing from what customers provide in control plane host, and the reconciler would set a default port when generating CAPI objects. 

CloudStack machine config doesn’t have default values at this stage.

### Data validation

In order to better sort out data validation and run time validation, we’ll reorganize validations in CLI.

* Extract data validation from ValidateClusterMachineConfigs (https://github.com/aws/eks-anywhere/blob/main/pkg/providers/cloudstack/validator.go#L127) and called from webhook.
* We have immutable fields validation in both webhook ([validateImmutableFieldsCloudStackMachineConfig](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/api/v1alpha1/cloudstackmachineconfig_webhook.go#L86) and provider [validateMachineConfigImmutability](https://github.com/aws/eks-anywhere/blob/01cd1e7c3da0c6d87b2d85c4ac6e61f409091e9d/pkg/providers/cloudstack/cloudstack.go#L162)). There’re some gaps between two places so we’ll need to decide the final immutable fields.
  * affinity group
  * [disk offering](https://github.com/aws/eks-anywhere/issues/5319) (can be modified from CAPI v1.4)
* We have [k8s version](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/providers/cloudstack/cloudstack.go#L1371) limitation due to capc version. This lightweight validation can be done in cluster webhook.

### Runtime Validation

#### Datacenter Reconciler

We’ll have CloudStackDatacenterReconciler to validate provided datacenter spec and update corresponding status and failure message. Any validation failure in this step would stop further reconciliation like cni/control plane node/worker node until customer can provide valid datacenter information.

#### How to read credential

  The datacenter reconciler needs a validator with CloudStack credential to build cmk to talk to CloudStack API. The reconciler will retrieve the CloudStack credentials from secrets. These secrets are created by the CLI during the management cluster creation or added by customers. CloudStack supports multiple secrets by referring secret name and credential profile name.

  CLI doesn't allow customers to [rotate credentials](https://github.com/aws/eks-anywhere/blob/main/designs/cloudstack-multiple-endpoints.md?plain=1#L187). Since reconciler reads credentials from secrets dynamically, customers are allowed to rotate secrets.

#### Where to build validator

  We can use the [validator factory](https://github.com/aws/eks-anywhere/blob/3c1fd0ff732641ed02137213863942403f59c320/pkg/providers/cloudstack/validator_registry.go#L25) to pass validatorRegistry to controller, so the validator is built in every reconcile loop.

### Cluster Reconciler skeleton
```
func (r *CloudstackReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
    log = log.WithValues("provider", "cloudstack")
    clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
    if err != nil {
        return controller.Result{}, err
    }
    
    return controller.NewPhaseRunner().Register(
        r.ipValidator.ValidateControlPlaneIP,      // checks whether the control plane ip is used by another cluster
        r.ValidateDatacenterConfig,                // checks no failre from datacenter reconciler 
        r.ValidateMachineConfig,                   // Once datacenter is validated, generate availability zone from datacenter to perform run time validation on machine config
        clusters.CleanupStatusAfterValidate,       // removes errors from the cluster status after all validation phases have been executed
        r.ReconcileControlPlane,
        r.CheckControlPlaneReady,      
        r.ReconcileCNI,      
        r.ReconcileWorkers,
    ).Run(ctx, log, clusterSpec)
}
```

