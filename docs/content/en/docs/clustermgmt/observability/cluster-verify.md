---
title: "Verify EKS Anywhere cluster status"
linkTitle: "Verify cluster status"
weight: 10
aliases:
    /docs/tasks/cluster/cluster-verify/
date: 2017-01-05
description: >
  Verify the status of EKS Anywhere clusters
---

{{% alert title="Note" color="primary" %}}
- To check the status of a single cluster, configure `kubectl` to communicate with the cluster by setting the `KUBECONFIG` environment variable to point to your cluster's `kubeconfig` file. 
- To check the status of workload clusters from a management cluster, configure `kubectl` with the `kubeconfig` of the management cluster.
{{% /alert %}}

#### Check cluster nodes
To verify the expected number of cluster nodes are present and running, use the `kubectl` command to show that nodes are `Ready`.

Worker nodes are named using the cluster name followed by the worker node group name. In the example below, the cluster name is `mgmt` and the worker node group name is `md-0`. The other nodes shown in the response are control plane or etcd nodes.

```bash
kubectl get nodes
```
```
NAME                              STATUS   ROLES           AGE   VERSION
mgmt-clrt4                        Ready    control-plane   3d22h   v1.27.1-eks-61789d8
mgmt-md-0-5557f7c7bxsjkdg-l2kpt   Ready    <none>          3d22h   v1.27.1-eks-61789d8
```

#### Check cluster machines

To verify that the expected number of cluster machines are present and running, use the `kubectl` command to show that the machines are `Running`.

The machine objects are named using the cluster name as a prefix and there should be one created for each node in your cluster. In the example below, the command was run against a management cluster with a single attached workload cluster. When the command is run against a management cluster, all machines for the management cluster and attached workload clusters are shown.

```bash
kubectl get machines -A
```
```
NAMESPACE     NAME                              CLUSTER   NODENAME                          PROVIDERID                                       PHASE     AGE     VERSION
eksa-system   mgmt-clrt4                        mgmt      mgmt-clrt4                        vsphere://421a801c-ac46-f47e-de1f-f070ef990c4d   Running   3d22h   v1.27.1-eks-1-27-4
eksa-system   mgmt-md-0-5557f7c7bxsjkdg-l2kpt   mgmt      mgmt-md-0-5557f7c7bxsjkdg-l2kpt   vsphere://421a4b9b-c457-fc4d-458a-d5092f981c5d   Running   3d22h   v1.27.1-eks-1-27-4
eksa-system   w01-7hzfh                         w01       w01-7hzfh                         vsphere://421a642b-f4ef-5764-47f9-5b56efcf8a4b   Running   15h     v1.27.1-eks-1-27-4
eksa-system   w01-etcd-z2ggk                    w01                                         vsphere://421ac003-3a1a-7dd9-ac83-bd0c75370cc4   Running   15h     
eksa-system   w01-md-0-799ffd7946x5gz8w-p94mt   w01       w01-md-0-799ffd7946x5gz8w-p94mt   vsphere://421a7b77-ca57-dc78-18bf-f361081a2c5e   Running   15h     v1.27.1-eks-1-27-4
```

#### Check cluster components
To verify cluster components are present and running, use the `kubectl` command to show that the system Pods are `Running`. The number of Pods may vary based on the infrastructure provider (vSphere, bare metal, Snow, Nutanix, CloudStack), and whether the cluster is a workload cluster or a management cluster.

```bash
kubectl get pods -A
```
```
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
```

#### Check control plane components

You can verify the control plane is present and running by filtering Pods by the `control-plane=controller-manager` label.
```bash
kubectl get pod -A -l control-plane=controller-manager
```
```
NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS      AGE
capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-8665b88c65-v982t       1/1     Running   0             3d21h
capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-67595c55d8-z7627   1/1     Running   0             3d21h
capi-system                         capi-controller-manager-88bdd56b4-wnk66                          1/1     Running   0             3d21h
capv-system                         capv-controller-manager-644d9864dc-hbrcz                         1/1     Running   1 (15h ago)   3d21h
eksa-packages                       eks-anywhere-packages-784c6fc8b9-2t5nr                           1/1     Running   0             3d21h
etcdadm-bootstrap-provider-system   etcdadm-bootstrap-provider-controller-manager-6bcdd4f5d7-wvqw8   1/1     Running   0             3d21h
etcdadm-controller-system           etcdadm-controller-controller-manager-6f96f5d594-kqnfw           1/1     Running   0             3d21h
```

#### Check workload clusters from management clusters

Set up `CLUSTER_NAME` and `KUBECONFIG` environment variable for the management cluster:
```bash
export CLUSTER_NAME=mgmt
export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

##### Check control plane resources for all clusters

Use the command below to check the status of cluster control plane resources. This is useful to verify clusters with multiple control plane nodes after an upgrade. The status for the management cluster and all attached workload clusters is shown.

```bash
kubectl get kubeadmcontrolplanes.controlplane.cluster.x-k8s.io -n eksa-system
```
```
NAME   CLUSTER   INITIALIZED   API SERVER AVAILABLE   REPLICAS   READY   UPDATED   UNAVAILABLE   AGE     VERSION
mgmt   mgmt      true          true                   1          1       1                       3d22h   v1.27.1-eks-1-27-4
w01    w01       true          true                   1          1       1         0             16h     v1.27.1-eks-1-27-4
```

Use the command below to check the status of a cluster resource. This is useful to verify cluster health after any mutating cluster lifecycle operation. The status for the management cluster and all attached workload clusters is shown.
```bash
kubectl get clusters.cluster.x-k8s.io -A -o=custom-columns=NAME:.metadata.name,CONTROLPLANE-READY:.status.controlPlaneReady,INFRASTRUCTURE-READY:.status.infrastructureReady,MANAGED-EXTERNAL-ETCD-INITIALIZED:.status.managedExternalEtcdInitialized,MANAGED-EXTERNAL-ETCD-READY:.status.managedExternalEtcdReady
```
```
NAME   CONTROLPLANE-READY   INFRASTRUCTURE-READY   MANAGED-EXTERNAL-ETCD-INITIALIZED   MANAGED-EXTERNAL-ETCD-READY
mgmt   true                 true                   <none>                              <none>
w01    true                 true                   true                                true
```