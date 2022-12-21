---
title: "Deploying Pod Identity Webhook"
linkTitle: "Deploying Pod Identity Webhook"
weight: 60
date: 2021-11-11
description: >  
---

1. youare deploying the pod Identity Webhook on the workload cluster w01-cluster. So, make sure the $KUBECONFIG env var is set to the path of the EKS workload Anywhere cluster w01-cluster. Then run the make command which basically deploy Kubernetes resources of the pod Identity Webhook to the workload cluster.

    ```BASH
    export KUBECONFIG=${PWD}/${WORKLOAD_CLUSTER_NAME}/${WORKLOAD_CLUSTER_NAME}-eks-a-cluster.kubeconfig

    cd amazon-eks-pod-identity-webhook/
    git checkout a65cc3d9c61cf6fc43f0f985818c474e0867d786

    make cluster-up IMAGE=amazon/amazon-eks-pod-identity-webhook:latest
    ```

1. After youhosted the service account public signing key and OIDC discovery documents in our S3 bucket, and deploying the pod identity webhook in the workload cluster, any application running in pods in the workload cluster can start accessing the desired AWS resources, as long as the pod is mounted with the right service account tokens. This part of configuring the pods with the right service account tokens and env vars is automated by the [amazon pod identity webhook](https://github.com/aws/amazon-eks-pod-identity-webhook) . the pod identity webhook mutates any pods launched using service accounts annotated with `eks.amazonaws.com/role-arn`. Let’s test that
