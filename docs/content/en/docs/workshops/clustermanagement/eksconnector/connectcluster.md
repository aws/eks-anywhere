---
title: "Connecting a cluster"
linkTitle: "Connecting a cluster"
weight: 3
date: 2021-11-11
description: >  
---

## Overview

Connecting EKS Anywhere cluster to AWS console involves two steps:

1. Registering the cluster
2. Applying the manifest file to enable connectivity

## Step 1 - Registering the cluster

To register your Kubernetes cluster with the console:

1. Open the Amazon EKS console at [https://console.aws.amazon.com/eks/home#/clusters](https://console.aws.amazon.com/eks/home#/clusters).
2. Click **Add cluster** and select **Register** to bring up the configuration page.
3. On the Configure cluster section, fill in the following fields:
    - **Name** – A unique name for your cluster.
    - **Provider** – Click to display the drop-down list of Kubernetes cluster providers. If you do not know the provider, select Other.
    - **EKS Connector role** – Select the role to use for connecting the cluster.
4. Select **Register cluster**.
5. The Cluster overview page displays. Click **Download YAML file** to download the manifest file to your local drive.

{{% alert title="Notes" color="warning" %}}
This is your only opportunity to download this file. Do not navigate away from this page, as the link will not be accessible and you must deregister the cluster and start the steps from the beginning. The manifest file can only be used once for the registered cluster. If you delete resources from the Kubernetes cluster, you must re-register the cluster and obtain a new manifest file.
{{% /alert %}}

## Step 2 - Applying the manifest file

Complete the connection by applying the Amazon EKS Connector manifest file to your Kubernetes cluster.
This must be done using the AWS CLI for both registration methods described above.
The Amazon EKS Connector registration expires if the manifest is not applied within three days.
If the cluster connection expires, the cluster must be deregistered before connecting the cluster again.

1. In the cluster's native environment, you can now apply the updated manifest file with the following command:

    ```bash
    kubectl apply -f eks-connector.yaml
    ```

2. Once the Amazon EKS Connector manifest and role binding YAML files are applied to your Kubernetes cluster, confirm that the cluster is now connected.

    ```bash
    aws eks describe-cluster \
        --name "my-first-registered-cluster" \
        --region AWS_REGION
    ```

    The output should include `status=ACTIVE`.

    To grant additional IAM users access to the Amazon EKS console to view the connected clusters, see [Granting access to a user to view a cluster](https://docs.aws.amazon.com/eks/latest/userguide/connector-grant-access.html). Your clusters will now be viewable in the AWS Management Console, as well as your connected [nodes](https://docs.aws.amazon.com/eks/latest/userguide/view-nodes.html) and [workloads](https://docs.aws.amazon.com/eks/latest/userguide/view-workloads.html).
