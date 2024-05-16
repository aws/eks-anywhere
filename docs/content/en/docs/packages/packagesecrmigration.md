---
title: "Packages regional ECR migration"
linkTitle: "Packages regional ECR migration"
weight: 4
description: >
    Migrating EKS Anywhere Curated Packages to latest regional ECR repositories
---

When you purchase an EKS Anywhere Enterprise Subscription through the Amazon EKS console or APIs, the AWS account that purchased the subscription is automatically granted access to EKS Anywhere Curated Packages in the AWS Region where the subscription is created. If you received trial access to EKS Anywhere Curated Packages or if you have an EKS Anywhere Enterprise Subscription that was created before October 2023, then you need to migrate your EKS Anywhere Curated Packages configuration to use the latest ECR regional repositories. This process would cause all the Curated Packages installed on the cluster to rollout and be deployed from the latest ECR regional repositories.

<details>
  <summary>Expand for packages registry to AWS Region table</summary>
  <br /> 
  {{% content "../clustermgmt/support/packages-registries.md" %}}
</details>

### Steps for Migration
1. Ensure you have an active EKS Anywhere Enterprise Subscription. For more information, refer [Purchase subscriptions.]({{< relref "../clustermgmt/support/purchase-subscription.md" >}})

2. If the AWS account that created the EKS Anywhere Enterprise Subscription through the Amazon EKS console or APIs and the AWS IAM user credentials for curated packages on your existing cluster are different, you need to update the aws-secret object on the cluster with new credentials. Refer [Updating the package credentials 
.]({{< relref "./packagecontroller.md#updating-the-package-credentials" >}})

3. Edit the `ecr-credential-provider-package` package on the cluster and update `matchImages` with the correct ECR package registry for the AWS Region where you created your subscription. Example, `346438352937.dkr.ecr.us-west-2.amazonaws.com` for `us-west-2`. Reference the table in the expanded output at the top of this page for a mapping of AWS Regions to ECR package registries.
    ```bash
    kubectl edit package ecr-credential-provider-package  -n eksa-packages-<cluster name>
    ```
    This causes `ecr-credential-provider-package` pods to rollout and the kubelet is configured to use AWS credentials for pulling images from the new regional ECR packages registry.

4. Edit the `PackageBundleController` object on the cluster and set the `defaultImageRegistry` and `defaultRegistry` to point to the ECR package registry for the AWS Region where you created your subscription.
    ```bash
    kubectl edit packagebundlecontroller <cluster name> -n eksa-packages
    ```
5. Restart the eks-anywhere-packages controller deployment.
    ```bash
    kubectl rollout restart deployment eks-anywhere-packages -n eksa-packages
    ```
    This step causes the package controller to pull down a new package bundle onto the cluster and marks the `PackageBundleController` as upgrade available. Example
    ```bash
    NAMESPACE       NAME              ACTIVEBUNDLE   STATE               DETAIL
    eksa-packages   my-cluster-name   v1-28-160      upgrade available   v1-28-274 available
    ```
6. Edit the `PackageBundleController` object on the cluster and set the `activeBundle` field to the new bundle number that is available.
    ```bash
    kubectl edit packagebundlecontroller <cluster name> -n eksa-packages
    ```
    This step causes all the packages on the cluster to be reinstalled and pods rolled out from the new registry.

7. Edit the `ecr-credential-provider-package` package again and now set the `sourceRegistry` to point to the ECR package registry for the AWS Region where you created your subscription.
    ```bash
    kubectl edit package ecr-credential-provider-package  -n eksa-packages-<cluster name>
    ```
    This causes `ecr-credential-provider-package` to be reinstalled from the new registry.
