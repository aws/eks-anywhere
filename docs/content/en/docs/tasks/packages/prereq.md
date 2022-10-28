---
title: "Package Prerequisites"
linkTitle: "Package Prerequisites"
weight: 10
description: >
  Prerequisites for using curated packages
---

## Prerequisites
Before installing any curated packages for EKS Anywhere, do the following:

* Check that the cluster `Kubernetes` version is `v1.21` or above. For example, you could run `kubectl get cluster -o yaml <cluster-name> | grep -i kubernetesVersion`
* Check that the version of `eksctl anywhere` is `v0.11.0` or above with the `eksctl anywhere version` command.
* It is recommended that the package controller is only installed on the management cluster.
* Check the existence of package controller:
    ```bash
    kubectl get pods -n eksa-packages | grep "eks-anywhere-packages"
    ```
    If the returned result is empty, you need to install the package controller.

* Install the package controller if it is not installed:
    Install the package controller
     
     *Note* This command is temporarily provided to ease integration with curated packages. This command will be deprecated in the future
 
     ```bash
     eksctl anywhere install packagecontroller -f $CLUSTER_NAME.yaml
     ```
