---
title: "Package management"
date: 2022-04-12
weight: 20
description: >
  Common tasks for managing curated packages.
---

The main goal of EKS Anywhere curated packages is to make it easy to install, configure and maintain operational components in an EKS Anywhere cluster. EKS Anywhere curated packages offers to run secure and tested operational components on EKS Anywhere clusters. Please check out [package controller]({{< relref "../../concepts" >}}) for more details.

{{% alert title="Important" color="warning" %}}

To install the EKS Anywhere binaries, create a EKS Anywhere cluster and see system requirements, please follow the [getting-started guide.]({{< relref "../../getting-started" >}})

{{% /alert %}}

### Install package controller

1. Install package controller
    ```bash
    PACKAGE_REGISTRY=oci://public.ecr.aws/l0g8r8j6/eks-anywhere-packages
    PACKAGE_VERSION=0.1.6-eks-a-v0.0.0-dev-build.2150
    helm install eks-anywhere-packages ${PACKAGE_REGISTRY} --version ${PACKAGE_VERSION}
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