
---
title: "Curated Packages Troubleshooting"
linkTitle: "Curated Packages Troubleshooting"
weight: 50
description: >
  Troubleshooting specific to curated packages
aliases:
   - /docs/tasks/troubleshoot/_packages
---


The major component of Curated Packages is the package controller. If the container is not running or not running correctly, packages will not be installed. Generally it should be debugged like any other Kubernetes application. The first step is to check that the pod is running.
```bash
kubectl get pods -n eksa-packages
```

You should see at least two pods with running and one or more refresher completed.
```bash
NAME                                     READY   STATUS      RESTARTS   AGE
eks-anywhere-packages-69d7bb9dd9-9d47l   1/1     Running     0          14s
eksa-auth-refresher-w82nm                0/1     Completed   0          10s
```

The describe command might help to get more detail on why there is a problem
```bash
kubectl describe pods -n eksa-packages
```

Logs of the controller can be seen in a normal Kubernetes fashion
```bash
kubectl logs deploy/eks-anywhere-packages -n eksa-packages controller
```

To get the general state of the package controller, run the following command:
```bash
kubectl get packages,packagebundles,packagebundlecontrollers -A
```

You should see an active packagebundlecontroller and an available bundle. The packagebundlecontroller should indicate the active bundle. It may take a few minutes to download and active the latest bundle. Thest state of the package in this example is installing and there is an error downloading the chart.
```bash
NAMESPACE       NAME                                         PACKAGE   AGE     STATE        CURRENTVERSION   TARGETVERSION                                                   DETAIL
eksa-packages   package.packages.eks.amazonaws.com/my-test   Test      2m33s   installing                    v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm (latest)   loading helm chart my-test: locating helm chart oci://public.ecr.aws/eks-anywhere/hello-eks-anywhere tag sha256:64ea03b119d2421f9206252ff4af4bf7cdc2823c343420763e0e6fc20bf03b68: failed to download "oci://public.ecr.aws/eks-anywhere/hello-eks-anywhere" at version "v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm"

NAMESPACE       NAME                                                   STATE
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-23-68    available

NAMESPACE       NAME                                                      ACTIVEBUNDLE   STATE    DETAIL
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/prod   v1-23-68       active   
```

### Error: this command is currently not supported

```
Error: this command is currently not supported
```

Curated packages became generally available with version `v0.11.0`. Use the version command to make sure you are running version v0.11.0 or later:

```bash
eksctl anywhere version
```

### Package controller not running
If you do not see a pod or various resources for the package controller, it may be that it is not installed.

```
No resources found in eksa-packages namespace.
```

Most likely the cluster was created with an older version of the EKS Anywhere CLI. Curated packages became generally available with `v0.11.0`. Use the `eksctl anywhere version` command to verify you are running a new enough release and you can use the `eksctl anywhere install packagecontroller` command to install the package controller on an older release.

### Errors during cluster creation

During cluster creation, you should see messages after the cluster is created when the package controller and any packages are installed.

```
🎉 Cluster created!
----------------------------------------------------------------------------------------------------------------
The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription.
----------------------------------------------------------------------------------------------------------------
Installing helm chart on cluster	{"chart": "eks-anywhere-packages", "version": "0.2.0-eks-a-v0.0.0-dev-build.3842"}
Warning: No AWS key/license provided. Please be aware this will prevent the package controller from installing curated packages.
```

If the `No AWS key/license provided` message appears during package controller installation, make sure you set and export the `EKSA_AWS_ACCESS_KEY_ID` and `EKSA_AWS_SECRET_ACCESS_KEY` varialbles to the access key and secret key of your AWS account. This will allow you to get access to container images in private ECR. A subscription is required to access the packages.

### ImagePullBackOff on Package or Package Controller

If a package or the package controller fails to start with ImagePullBackOff

```
NAME                                     READY   STATUS             RESTARTS   AGE
eks-anywhere-packages-6589449669-q7rjr   0/1     ImagePullBackOff   0          13h
```

This is most like because the machine running kubelet in your Kubernetes cluster cannot access the registry with the images or those images do not exist on that registry. Log into the machine and see if it has access to the images:

```bash
ctr image pull public.ecr.aws/eks-anywhere/eks-anywhere-packages@sha256:whateveritis
```

### Error: cert-manager is not present in the cluster
```
Error: curated packages cannot be installed as cert-manager is not present in the cluster
```
This is most likely caused by an action to install curated packages at a workload cluster. Note curated packages installation at workload cluster creation is currently not supported. In order to use packages on workload clusters, you can do a post-creation curated packages installation:
- Install cert-manager, refer to cert-manager [installation guide](https://cert-manager.io/docs/installation/) for more details.
- Install package controller using following command:
  ```bash
  eksctl anywhere install packagecontroller -f $CLUSTER_NAME.yaml
  ```
- Install packages, refer to [package management]({{< relref "../../tasks/packages" >}}) for more details.
