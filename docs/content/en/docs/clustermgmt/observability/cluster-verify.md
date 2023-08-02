---
title: "Verify cluster"
linkTitle: "Verify cluster"
weight: 11
aliases:
    /docs/tasks/cluster/cluster-verify/
date: 2017-01-05
description: >
  How to verify an EKS Anywhere cluster is running properly
---

**Verify an EKS Anywhere cluster**
First, set the `KUBECONFIG` environment variable for your cluster:

```
export KUBECONFIG=<path-to-cluster-kubeconfig>
```

To verify that the cluster components are up and running, use the `kubectl` command to show that the pods are all running. Below is the example but output may change as per the deployment. Please ensure that pods should be in running STATUS:

```
kubectl get pods -n kube-system | head -n 5
NAME                                     READY   STATUS    RESTARTS      AGE
kube-apiserver-eksapoc-lkfnr             1/1     Running   0             10d
kube-apiserver-eksapoc-wjq2c             1/1     Running   0             10d
kube-apiserver-eksapoc-zqsv9             1/1     Running   0             10d
kube-controller-manager-eksapoc-lkfnr    1/1     Running   0             10d
kube-controller-manager-eksapoc-wjq2c    1/1     Running   0             10d
```

To verify that the expected number of cluster nodes are up and running, use the `kubectl` command to show that nodes are `Ready`.

```
kubectl get nodes | grep -i control-plane
NAME                                 STATUS   ROLES           AGE   VERSION
eksapoc-lkfnr                        Ready    control-plane   10d   v1.27.1-eks-75a5dcc
eksapoc-wjq2c                        Ready    control-plane   10d   v1.27.1-eks-75a5dcc
eksapoc-zqsv9                        Ready    control-plane   10d   v1.27.1-eks-75a5dcc
```

**Deploy Test Application**

To test a workload in your cluster you can try deploying the [hello-eks-anywhere]({{< relref "../../workloadmgmt/test-app" >}}).
