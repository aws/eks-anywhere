---
title: "Enable Pod IAM and Create Cluster"
linkTitle: "Enable Pod IAM and Create Cluster"
weight: 40
date: 2021-11-11
description: >  
---

1. When creating the EKS Anywhere cluster, you need to configure the kube-apiserver’s `service-account-issuer` flag so it can issue and mount projected service account tokens in pods. For this, use the value obtained in the first section for `$ISSUER_HOSTPATH` as the `podIamConfig.serviceAccountIssuer`. Configure the kube-apiserver by setting this value through the EKS Anywhere cluster spec as follows:

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
    name: w01-cluster
    spec:
    podIamConfig:
        serviceAccountIssuer: https://[ISSUER_HOSTPATH]
    ```

1. Leave the remaining fields in [cluster spec](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/) as they were and create the cluster using the eksctl anywhere create cluster command.

    ```bash
    eksctl anywhere create cluster -f $WORKLOAD_CLUSTER_NAME.yaml \
    --kubeconfig mgmt-cluster/mgmt-cluster-eks-a-cluster.kubeconfig \
    -v 9 > $WORKLOAD_CLUSTER_NAME-$(date "+%Y%m%d%H%M").log 2>&1
    ```

1. You can list the workload clusters managed by the management cluster.

    ```bash
    export KUBECONFIG=${PWD}/${MGMT_CLUSTER_NAME}/${MGMT_CLUSTER_NAME}-eks-a-cluster.kubeconfig
    kubectl get clusters
    ```

1. Once the workload cluster w01-cluster is created you can use it with the generated KUBECONFIG file in your local directory:

    ```bash
    export KUBECONFIG=${PWD}/${WORKLOAD_CLUSTER_NAME}/${WORKLOAD_CLUSTER_NAME}-eks-a-cluster.kubeconfig
    kubectl get nodes
    ```