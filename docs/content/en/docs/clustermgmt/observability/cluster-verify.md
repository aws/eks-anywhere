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

**Verify an EKS Anywhere cluster from the target cluster**
First, set the `KUBECONFIG` environment variable is set up for the target cluster (management or workload cluster) of choice:

```
export KUBECONFIG=<path-to-cluster-kubeconfig>
```

To verify that the cluster components are up and running, use the `kubectl` command to show that the pods are all running.
There will be more or less pods depending on whether a management or workload cluster is targeted.

```
kubectl get pods -A
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS      AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-8665b88c65-v982t       1/1     Running   0             3d22h
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-67595c55d8-z7627   1/1     Running   0             3d22h
capi-system                         capi-controller-manager-88bdd56b4-wnk66                          1/1     Running   0             3d22h
capv-system                         capv-controller-manager-644d9864dc-hbrcz                         1/1     Running   1 (16h ago)   3d22h
cert-manager                        cert-manager-548579646f-4tgb2                                    1/1     Running   0             3d22h
cert-manager                        cert-manager-cainjector-cbb6df554-w5fjx                          1/1     Running   0             3d22h
cert-manager                        cert-manager-webhook-54f748c89b-qnfr2                            1/1     Running   0             3d22h
eksa-packages                       ecr-credential-provider-package-4c7mk                            1/1     Running   0             3d22h
eksa-packages                       ecr-credential-provider-package-nvlkb                            1/1     Running   0             3d22h
eksa-packages                       eks-anywhere-packages-784c6fc8b9-2t5nr                           1/1     Running   0             3d22h
eksa-system                         eksa-controller-manager-76f484bd5b-x6qld                         1/1     Running   0             3d22h
etcdadm-bootstrap-provider-system   etcdadm-bootstrap-provider-controller-manager-6bcdd4f5d7-wvqw8   1/1     Running   0             3d22h
etcdadm-controller-system           etcdadm-controller-controller-manager-6f96f5d594-kqnfw           1/1     Running   0             3d22h
kube-system                         cilium-lbqdt                                                     1/1     Running   0             3d22h
kube-system                         cilium-operator-55c4778776-jvrnh                                 1/1     Running   0             3d22h
kube-system                         cilium-operator-55c4778776-wjjrk                                 1/1     Running   0             3d22h
kube-system                         cilium-psqm2                                                     1/1     Running   0             3d22h
kube-system                         coredns-69797695c4-kdtjc                                         1/1     Running   0             3d22h
kube-system                         coredns-69797695c4-r25vv                                         1/1     Running   0             3d22h
kube-system                         etcd-mgmt-clrt4                                                  1/1     Running   0             3d22h
kube-system                         kube-apiserver-mgmt-clrt4                                        1/1     Running   0             3d22h
kube-system                         kube-controller-manager-mgmt-clrt4                               1/1     Running   0             3d22h
kube-system                         kube-proxy-588gj                                                 1/1     Running   0             3d22h
kube-system                         kube-proxy-hrksw                                                 1/1     Running   0             3d22h
kube-system                         kube-scheduler-mgmt-clrt4                                        1/1     Running   0             3d22h
kube-system                         kube-vip-mgmt-clrt4                                              1/1     Running   0             3d22h
kube-system                         vsphere-cloud-controller-manager-7vzjx                           1/1     Running   0             3d22h
kube-system                         vsphere-cloud-controller-manager-cqfs5                           1/1     Running   0             3d22h
kube-system                         vsphere-csi-controller-79fb7fbb6b-9vvr7                          5/5     Running   0             3d22h
kube-system                         vsphere-csi-node-bths4                                           3/3     Running   0             3d22h
kube-system                         vsphere-csi-node-r8lwp                                           3/3     Running   0             3d22h
```

To verify that the expected number of cluster nodes are up and running, use the `kubectl` command to show that nodes are `Ready`.

This will confirm that the expected number of nodes are present.
Worker nodes are named using the cluster name followed by the worker node group name (example: mgmt-md-0), the others are control plane nodes.

```
kubectl get nodes
NAME                              STATUS   ROLES           AGE   VERSION
mgmt-clrt4                        Ready    control-plane   3d22h   v1.27.1-eks-61789d8
mgmt-md-0-5557f7c7bxsjkdg-l2kpt   Ready    <none>          3d22h   v1.27.1-eks-61789d8
```

**Verify an EKS Anywhere cluster from the management cluster**

Set up `CLUSTER_NAME` and `KUBECONFIG` environment variable for the management cluster:
```
export CLUSTER_NAME=mgmt
export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

You may verify the control plane is up and running by filtering the pods by the `control-plane=controller-manager` label.
```
kubectl get pod -A -l control-plane=controller-manager
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS      AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-8665b88c65-v982t       1/1     Running   0             3d21h
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-67595c55d8-z7627   1/1     Running   0             3d21h
capi-system                         capi-controller-manager-88bdd56b4-wnk66                          1/1     Running   0             3d21h
capv-system                         capv-controller-manager-644d9864dc-hbrcz                         1/1     Running   1 (15h ago)   3d21h
eksa-packages                       eks-anywhere-packages-784c6fc8b9-2t5nr                           1/1     Running   0             3d21h
etcdadm-bootstrap-provider-system   etcdadm-bootstrap-provider-controller-manager-6bcdd4f5d7-wvqw8   1/1     Running   0             3d21h
etcdadm-controller-system           etcdadm-controller-controller-manager-6f96f5d594-kqnfw           1/1     Running   0             3d21h
```

You may also check the status of the cluster control plane resource directly. 
This can be especially useful to verify clusters with multiple control plane nodes after an upgrade.

```
kubectl get kubeadmcontrolplanes.controlplane.cluster.x-k8s.io -n eksa-system
NAME   CLUSTER   INITIALIZED   API SERVER AVAILABLE   REPLICAS   READY   UPDATED   UNAVAILABLE   AGE     VERSION
mgmt   mgmt      true          true                   1          1       1                       3d22h   v1.27.1-eks-1-27-4
w01    w01       true          true                   1          1       1         0             16h     v1.27.1-eks-1-27-4
```

You may check the status of the cluster resource directly.
```
kubectl get clusters.cluster.x-k8s.io -A -o=custom-columns=NAME:.metadata.name,CONTROLPLANE-READY:.status.controlPlaneReady,INFRASTRUCTURE-READY:.status.infrastructureReady,MANAGED-EXTERNAL-ETCD-INITIALIZED:.status.managedExternalEtcdInitialized,MANAGED-EXTERNAL-ETCD-READY:.status.managedExternalEtcdReady

NAME   CONTROLPLANE-READY   INFRASTRUCTURE-READY   MANAGED-EXTERNAL-ETCD-INITIALIZED   MANAGED-EXTERNAL-ETCD-READY
mgmt   true                 true                   <none>                              <none>
w01    true                 true                   true                                true
```

To verify that the expected number of cluster machines are up and running, use the `kubectl` command to show that the machines are `Running`.

This will confirm that the expected number of provider machines with the correct version have been provisioned and running.
The machine objects are named using the cluster name as a prefix and there should be one created for each node in your cluster.

```
kubectl get machines -A
NAMESPACE     NAME                              CLUSTER   NODENAME                          PROVIDERID                                       PHASE     AGE     VERSION
eksa-system   mgmt-clrt4                        mgmt      mgmt-clrt4                        vsphere://421a801c-ac46-f47e-de1f-f070ef990c4d   Running   3d22h   v1.27.1-eks-1-27-4
eksa-system   mgmt-md-0-5557f7c7bxsjkdg-l2kpt   mgmt      mgmt-md-0-5557f7c7bxsjkdg-l2kpt   vsphere://421a4b9b-c457-fc4d-458a-d5092f981c5d   Running   3d22h   v1.27.1-eks-1-27-4
eksa-system   w01-7hzfh                         w01       w01-7hzfh                         vsphere://421a642b-f4ef-5764-47f9-5b56efcf8a4b   Running   15h     v1.27.1-eks-1-27-4
eksa-system   w01-etcd-z2ggk                    w01                                         vsphere://421ac003-3a1a-7dd9-ac83-bd0c75370cc4   Running   15h     
eksa-system   w01-md-0-799ffd7946x5gz8w-p94mt   w01       w01-md-0-799ffd7946x5gz8w-p94mt   vsphere://421a7b77-ca57-dc78-18bf-f361081a2c5e   Running   15h     v1.27.1-eks-1-27-4
```

To test a workload in your cluster you can try deploying the [hello-eks-anywhere]({{< relref "../../workloadmgmt/test-app" >}}).
