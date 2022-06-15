# `Cluster` reference to `Bundles`

## Introduction

**Problem:** we rely on naming and a default namespace to find the `Bundles` object being used by an EKS-A `Cluster`.

Currently, when we create a cluster, we rename the Bundles object we are using to have the same name as the Cluster object. Then, we use that same name and same namespace to retrieve it from the cluster and rebuild a cluster.Spec in the CLI/controller.

This is not a very clean pattern since it forces us to duplicate `Bundles` for multiple clusters. Also, in order to point a cluster to a new `Bundles` we need to update the existing one (or delete and recreate). Ideally, `Bundles` should be immutable in the cluster in the same way they are in real life.

### Goals and Objectives
As an EKS-A cluster administrator I want to:

* Not have duplicated `Bundles` objects
* Not have my `default` namespace polluted with `Bundles` objects since these are internal details of the EKS-A implementation
* Upgrade a `Cluster`'s dependencies through the controller by pointing it to a new `Bundles` object

## Solution Details

Add a new field to the `Cluster` object:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
  namespace: default
spec:
  bundlesRef:
    name: bundles-15
    namespace: eksa-system
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
...
```

A few notes:

* This new field should be updatable for managed clusters through.
* For self-managed clusters (aka management clusters) all upgrades should go through the CLI, so we should let it update this field after an upgrade. However, we can potentially let users use this field to upgrade to a particular `Bundles` instead of always to the last one. This decision is not in scope for this doc.
* When creating a cluster, this field should be optional. For management clusters, the CLI will set the reference to the latest available `Bundles`. For workload clusters, the default will be the `Bundles` being referenced by the management cluster.
* `Bundles` should be installed in the `eksa-system` by default.
