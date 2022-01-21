---
title: "Granting access to a user"
linkTitle: "Granting access to a user"
weight: 4
date: 2021-11-11
description: >  
---

Part of this article we will see how to grant additional IAM users access to the Amazon EKS console to view information about the Kubernetes workloads and pods running on your connected cluster.

## Amazon EKS Connector cluster role

To create and apply the Amazon EKS Connector cluster role

1. Download the eks-connector cluster role template.

    ```bash
    curl -o eks-connector-clusterrole.yaml https://amazon-eks.s3.us-west-2.amazonaws.com/eks-connector/manifests/eks-connector-console-roles/eks-connector-clusterrole.yaml
    ```

2. Edit the cluster role template YAML file, replace references of `%IAM_ARN%` with the Amazon Resource Name (ARN) of your IAM user or role.

3. Apply the Amazon EKS Connector cluster role YAML to your Kubernetes cluster.

    ```bash
    kubectl apply -f eks-connector-clusterrole.yaml
    ```

In order for an IAM user or role to vizualize the workloads on the Amazon EKS console, they must be associated with a Kubernetes `role` or `clusterrole` with necessary permissions to read these resources. For more information, see [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) in the Kubernetes documentation.

## Configure an IAM user to access the connected cluster

You can download the example manifest file to create a clusterrole and `clusterrolebinding` or a role and `rolebinding`:

1. The cluster role `eks-connector-console-dashboard-full-access-clusterrole`, gives access to all namespaces and resources that can be visualized in the console. You can change the name of the `role`, `clusterrole` and their corresponding binding before applying it to your cluster. Use the following command to download a sample file.

    ```bash
    curl -o eks-connector-console-dashboard-full-access-group.yaml https://amazon-eks.s3.us-west-2.amazonaws.com/eks-connector/manifests/eks-connector-console-roles/eks-connector-console-dashboard-full-access-group.yaml
    ```

2. Edit the full access or restricted access YAML file to replace references of `%IAM_ARN%` with the Amazon Resource Name (ARN) of your IAM user or role.

3. Apply the full access or restricted access YAML files to your Kubernetes cluster. Replace the YAML file value with your own.

    ```bash
    kubectl apply -f eks-connector-console-dashboard-full-access-group.yaml
    ```

To view your connected cluster and nodes, see [View nodes](https://docs.aws.amazon.com/eks/latest/userguide/view-nodes.html). To view workloads, see [View workloads](https://docs.aws.amazon.com/eks/latest/userguide/view-workloads.html). Keep in mind that some node and workload data will not be populated for connected clusters.
