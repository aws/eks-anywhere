---
title: "EKS Anywhere curated package management"
linkTitle: "Package management"
date: 2022-04-12
weight: 20
description: >
  Common tasks for managing curated packages.
---

The main goal of EKS Anywhere curated packages is to make it easy to install, configure and maintain operational components in an EKS Anywhere cluster. EKS Anywhere curated packages offers to run secure and tested operational components on EKS Anywhere clusters. Please check out [EKS Anywhere curated packages]({{< relref "../../concepts/packages" >}}) for more details.

### Check the existence of package controller
```bash
kubectl get pods -n eksa-packages | grep "eks-anywhere-packages"
```
Skip the following installation steps if the returned result is not empty.

{{% alert title="Important" color="warning" %}}

* To install EKS Anywhere, create an EKS Anywhere cluster or review the EKS Anywhere system requirements. See the [Getting started]({{< relref "../../getting-started" >}}) guide for details.

* Check if the version of `eksctl anywhere` is `v0.9.0` or above with the following commands:
    ```bash
    eksctl anywhere version
    ```
* Make sure cert-manager is up and running in the cluster.

{{% /alert %}}

### Install package controller

1. Install the package controller
    ```bash
    eksctl anywhere install packagecontroller --kube-version 1.21
    ```

1. Check the package controller
    ```bash
    kubectl get pods -n eksa-packages
    ```

    Example command output
    ```
    NAME                                       READY   STATUS     RESTARTS   AGE
    eks-anywhere-packages-57778bc88f-587tq     2/2     Running    0          16h
    ```
### Curated package list
See [packages]({{< relref "../../reference/packagespec" >}}) for the complete curated package list.