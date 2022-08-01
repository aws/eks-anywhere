# Embed EKS-D resources in EKS-A cluster

## Introduction

**Problem:** Currently, the EKS Anywhere controller pulls the EKS Distro manifest from the release URL provided in the versions bundle, which requires access to the internet.
It would be ideal to have this manifest and its CRDs embedded in the cluster to allow the controller to operate without the requirement to make that external call.
This will eventually be needed for air-gapped support.

### Goals and Objectives

As an EKS Anywhere user:

* I want the EKSA controller to be able to operate without internet access
* I want EKSD components to be supported during upgrade

## Overview of Solution

With this feature, the EKSD manifest & its CRD will be applied directly to the cluster.
This will allow the controller to grab the EKSD release resources from within the cluster instead of reading in the EKSD manifest from the internet when fetching the applied spec.
We will do this similarly to how we apply the EKSA components to the cluster, introducing a new task to the create workflow called `InstallEksdComponentsTask`.
In the upgrade workflow, we will also add a step to apply the EKSD resources as we update other cluster resources.

#### EKS-D bundle changes

We currently don’t pull in the EKSD release CRD in the bundle, but we need to apply this CRD before applying the release manifest.
I propose introducing a new field `components` in the EKSD bundle to refer to the file in S3 that stores this CRD.

Example bundles manifest with proposed changes:

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  creationTimestamp: "2022-02-01T16:41:19Z"
spec:
  cliMaxVersion: v0.0.0
  cliMinVersion: v0.0.0
  number: 1
  versionsBundles:
    eksD:
      channel: 1-21
      components: https://distro.eks.amazonaws.com/crds/releases.distro.eks.amazonaws.com-v1alpha1.yaml
      gitCommit: |
        fcddf59a516102e20f75b6c672e3aa74cf2da877
      kindNode:
        arch:
          ...
```

From here, we can use `kubectl apply` to apply the content from the EKSD release CRD & manifest to the cluster.

#### Cluster spec changes

We will need a way to refer to the EKSD release object on the cluster when calling it from the controller.
One way to achieve this is by grabbing the EKSD bundle from the bundles object.
We already apply the bundles to the cluster, and we fetch the bundles object in the controller.
However, we depend on the cluster name to fetch this bundle.
A more accurate way to fetch this bundle is by adding a field `bundlesRef` to the cluster spec.
This will serve as our source of truth for the bundles object on the cluster.


```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: test_cluster
spec:
  controlPlaneConfiguration:
    ...
  bundlesRef:
    apiGroup: anywhere.eks.amazonaws.com/v1alpha1
    kind: Bundles
    name: "bundle-name"
    namespace: eksa-system 
```

This bundles ref will allow us to accurately fetch the bundle from the cluster, access the EKSD bundle, and grab the correct release manifest in the controller.

### Upgrading EKSD components

Now that the EKSD components are embedded in the cluster, we have to ensure that those components are supported during upgrade.
We will update the upgrader in cluster manager to not only check for updates to EKSA, but it will also check if updates are needed for the EKSD objects on the cluster.
We will have a method `EksdChangeDiff` to check the EKSD version (based on the `name` in the bundle) to decide if we need to update the EKSD manifest objects that are applied to the cluster.


