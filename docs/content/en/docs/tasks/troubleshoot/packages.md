
---
title: "Curated Packages Troubleshooting"
linkTitle: "Curated Packages Troubleshooting"
weight: 50
description: >
  Troubleshooting specific to curated packages
aliases:
   - /docs/tasks/troubleshoot/_packages
---


You must set and export the `CURATED_PACKAGES_SUPPORT` environment variable before running any commands for packages to activate the feature flag.

```bash
export CURATED_PACKAGES_SUPPORT=true
```

The major component of Curated Packages is the package controller. If the container is not running or not running correctly, packages will not be installed. Generally it should be debugged like any other Kubernetes application. The first step is to check that the pod is running.
```bash
kubectl get pods -n eksa-packages
```

You should see one pod running with two containers
```bash
NAME                                     READY   STATUS    RESTARTS   AGE
eks-anywhere-packages-6c7db8bc6f-xg6bq   2/2     Running   0          3m35s
```

The describe command might help to get more detail on why there is a problem
```bash
kubectl describe pods -n eksa-packages
```

Logs of the controller can be seen in a normal Kubernetes fashion
```bash
kubectl logs deploy/eks-anywhere-packages -n eksa-packages controller
```

The general state of the package can be seen through the custom resources
```bash
kubectl get packages,packagebundles,packagebundlecontrollers -A
```

This will generate output similar to this
```bash
NAMESPACE       NAME                                         PACKAGE   AGE     STATE        CURRENTVERSION   TARGETVERSION                                                   DETAIL
eksa-packages   package.packages.eks.amazonaws.com/my-test   Test      2m33s   installing                    v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm (latest)   loading helm chart my-test: locating helm chart oci://public.ecr.aws/l0g8r8j6/hello-eks-anywhere tag sha256:64ea03b119d2421f9206252ff4af4bf7cdc2823c343420763e0e6fc20bf03b68: failed to download "oci://public.ecr.aws/l0g8r8j6/hello-eks-anywhere" at version "v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm"

NAMESPACE       NAME                                                   STATE
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-21-1001    active

NAMESPACE       NAME                                                                                 STATE
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/eksa-packages-bundle-controller   active
```

Looking at the output, you can see the active packagebundlecontroller and packagebundle. The state of the package is "installing".

### Error: curated packages installation is not supported in this release

```
Error: curated packages installation is not supported in this release
```
Curated packages is supported behind a feature flag, you must set and export the `CURATED_PACKAGES_SUPPORT` environment variable before 

```bash
export CURATED_PACKAGES_SUPPORT=true
```

### Error: this command is currently not supported

```
Error: this command is currently not supported
```

Curated packages is supported behind a feature flag, you must set and export the `CURATED_PACKAGES_SUPPORT` environment variable.

```bash
export CURATED_PACKAGES_SUPPORT=true
```

### Package controller not running
If you do not see a pod or various resources for the package controller, it may be that it is not installed.

```
No resources found in eksa-packages namespace.
```

Most likely the cluster was created with an older version of the EKS Anywhere CLI or the feature flag was not enabled. If you run the version command, it should return `v0.9.0` or later release.

```bash
eksctl anywhere version
```
Curated packages is supported behind a feature flag, you must set and export the `CURATED_PACKAGES_SUPPORT` environment variable.

```bash
export CURATED_PACKAGES_SUPPORT=true
```

During cluster creation, you should see messages after the cluster is created when the package controller and any packages are installed.

```
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

### ImagePullBackOff on Package or Package Controller

If a package or the package controller fails to start with ImagePullBackOff

```
NAME                                     READY   STATUS             RESTARTS   AGE
eks-anywhere-packages-6589449669-q7rjr   0/2     ImagePullBackOff   0          13h
```

This is most like because the machine running kubelet in your Kubernetes cluster cannot access the registry with the images or those images do not exist on that registry. Log into the machine and see if it has access to the images:

```bash
ctr image pull public.ecr.aws/eks-anywhere/eks-anywhere-packages@sha256:whateveritis
```

