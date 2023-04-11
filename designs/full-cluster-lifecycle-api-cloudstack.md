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

* We currently only supports k8s version < 1.24 on CloudStack
* We only support Redhat, not on Ubuntu or Bottlerocket yet.

## Overview of Solution

The EKS-A controller running in management cluster reconciles EKS-A CRDs, which fully manages workload clusters. [Detailed lifecycle diagram](https://github.com/aws/eks-anywhere/blob/4d307c6fe75075adadae38537d0f211c9142003e/designs/full-cluster-lifecycle-api.md#L55)

### CloudStack defaults

We already have [SetDefaults](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/api/v1alpha1/cloudstackdatacenterconfig_types.go#L216) for datacenter, we can utilize it in its mutation webhook (need to create one). Cloudstack machine config doesn’t have machine config [default](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/providers/cloudstack/cloudstack.go#L162) values at this stage.

### Static/Data validation

CloudStack already has webhook for datacenter and machine config. We’ll reorganize validations, so keep static/data validations in webhook and keep validations need cmk in validator. The validator will be used in CloudStack provider, cluster reconciler and datacenter reconciler.

* Extract static/data validation from [validateDatacenterConfig](https://github.com/aws/eks-anywhere/blob/main/pkg/providers/cloudstack/validator.go#L60) and ValidateClusterMachineConfigs (https://github.com/aws/eks-anywhere/blob/main/pkg/providers/cloudstack/validator.go#L127) to datacenterConfig/machineConfig types, and called in webhook. The runtime validations stay in validator and will be called during reconcile.
* We have immutable fields validation in both webhook ([validateImmutableFieldsCloudStackMachineConfig](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/api/v1alpha1/cloudstackmachineconfig_webhook.go#L86) and provider ([validateMachineConfigImmutability](https://github.com/aws/eks-anywhere/blob/01cd1e7c3da0c6d87b2d85c4ac6e61f409091e9d/pkg/providers/cloudstack/cloudstack.go#L162)). There’re some gaps between two places so we’ll need to decide the final validation.
  * affinity group
  * [disk offering](https://github.com/aws/eks-anywhere/issues/5319)

### Runtime Validation

#### Validator Initialization

* **How to read credential**

  The validator requires cloudstack credential to build cmk. Our current CLI reads credential from env, and has a validation that accept new credential but not allowing existing credential update.

    1. (preferred) Read credentials from existing secrets. Customers would need to update k8s secrets object manually. Our [current mechanism](https://github.com/aws/eks-anywhere/blob/main/designs/cloudstack-multiple-endpoints.md?plain=1#L187)) doesn't allow users to update cluster and secrets at the same time as a safeguard. But for controller, users naturally can't perform those two operations at the same time, so it won't be a problem.
    2. Read credentials from env. This the credentials are call captured as a single env var so there's limitation to only store one and hard to rotate.

* **Where to build validator**

    We can use the [validator factory](https://github.com/aws/eks-anywhere/blob/3c1fd0ff732641ed02137213863942403f59c320/pkg/providers/cloudstack/validator_registry.go#L25) to pass validatorRegistry to controller, so the validator is built in every reconcile loop.

* **Datacenter Reconciler**

  We’ll have CloudStackDatacenterReconciler to validate and update corresponding status and failure message.

* **Cluster Reconciler**
```
func (r *CloudstackReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
    log = log.WithValues("provider", "cloudstack")
    clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
    if err != nil {
        return controller.Result{}, err
    }
    
    return controller.NewPhaseRunner().Register(
        r.ValidateAndSetEnv,
        r.ipValidator.ValidateControlPlaneIP,      
        r.ValidateDatacenterConfig,  
        r.ValidateMachineConfig,         
        clusters.CleanupStatusAfterValidate,      
        r.ReconcileControlPlane,      
        r.CheckControlPlaneReady,      
        r.ReconcileCNI,      
        r.ReconcileWorkers,
    ).Run(ctx, log, clusterSpec)
}
```

* validateAndSetEnv: setting [eksa license](https://github.com/aws/eks-anywhere/blob/3c1fd0ff732641ed02137213863942403f59c320/pkg/providers/cloudstack/cloudstack.go#L395), [validate k8s version](https://github.com/aws/eks-anywhere/blob/ed4425dadb19600b4eb446d29b81f5c2441c16f6/pkg/providers/cloudstack/cloudstack.go#L1371),
  [setDefaultAndValidateControlPlaneHostPort](https://github.com/aws/eks-anywhere/blob/3c1fd0ff732641ed02137213863942403f59c320/pkg/providers/cloudstack/validator.go#L211)
* ValidateDatacenterConfig checks status and failure message from CloudStackDatacenterReconciler
* After successful datacenter config validation, ValidateMachineConfig runs machine config validations via cmk for each availability zone in datacenter.
