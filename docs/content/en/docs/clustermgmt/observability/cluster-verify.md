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
To verify that a cluster control plane is up and running, use the `kubectl` command to show that the control plane pods are all running.

```
kubectl get po -A -l control-plane=controller-manager
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-57b99f579f-sd85g       2/2     Running   0          47m
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-79cdf98fb8-ll498   2/2     Running   0          47m
capi-system                         capi-controller-manager-59f4547955-2ks8t                         2/2     Running   0          47m
capi-webhook-system                 capi-controller-manager-bb4dc9878-2j8mg                          2/2     Running   0          47m
capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-6b4cb6f656-qfppd       2/2     Running   0          47m
capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-bf7878ffc-rgsm8    2/2     Running   0          47m
capi-webhook-system                 capv-controller-manager-5668dbcd5-v5szb                          2/2     Running   0          47m
capv-system                         capv-controller-manager-584886b7bd-f66hs                         2/2     Running   0          47m

```

You may also check the status of the cluster control plane resource directly. 
This can be especially useful to verify clusters with multiple control plane nodes after an upgrade.
```
kubectl get kubeadmcontrolplanes.controlplane.cluster.x-k8s.io
NAME                       INITIALIZED   API SERVER AVAILABLE   VERSION              REPLICAS   READY   UPDATED   UNAVAILABLE
supportbundletestcluster   true          true                   v1.20.7-eks-1-20-6   1          1       1
```

To verify that the expected number of cluster worker nodes are up and running, use the `kubectl` command to show that nodes are `Ready`.
This will confirm that the expected number of worker nodes are present.
Worker nodes are named using the cluster name followed by the worker node group name (example: my-cluster-md-0)
```
kubectl get nodes
NAME                                           STATUS   ROLES                  AGE    VERSION
supportbundletestcluster-md-0-55bb5ccd-mrcf9   Ready    <none>                 4m   v1.20.7-eks-1-20-6
supportbundletestcluster-md-0-55bb5ccd-zrh97   Ready    <none>                 4m   v1.20.7-eks-1-20-6
supportbundletestcluster-mdrwf                 Ready    control-plane,master   5m   v1.20.7-eks-1-20-6
```


To test a workload in your cluster you can try deploying the [hello-eks-anywhere]({{< relref "../../workloadmgmt/test-app" >}}).
