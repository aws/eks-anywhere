---
title: "License EKS Anywhere cluster"
linkTitle: "License cluster"
weight: 20
date: 2023-09-21
aliases:
    /docs/tasks/cluster/cluster-license/
description: >
  Apply an EKS Anywhere Enterprise Subscription license to your EKS Anywhere cluster
---

When you purchase an EKS Anywhere Enterprise Subscription, licenses are provisioned in [AWS License Manager](https://docs.aws.amazon.com/license-manager/latest/userguide/license-manager.html) in the AWS account and region you used to purchase the subscription. After purchasing your subscription, you can view your licenses, accept the license grants, and apply the license IDs to your EKS Anywhere clusters. The License ID strings are used when you create support cases to validate your cluster is eligible to receive support.

### View licenses for an EKS Anywhere subscription

You can view the licenses associated with an EKS Anywhere Enterprise Subscription in the [Amazon EKS Console.](https://console.aws.amazon.com/eks/home#/eks-anywhere)

Follow the steps below to view EKS Anywhere licenses with the AWS CLI.

**Get license ARNs based on subscription `name` with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription` in the `--query` string with the `name` for your subscription.

```bash
aws eks list-eks-anywhere-subscriptions \
  --region 'region-code' \
  --query 'subscriptions[?name==`my-subscription`].licenseArns[]'
```

The License ID is the last part of the ARN string. For example, the License ID is shown in bold in the following example: *arn:aws:license-manager::12345678910:**license:l-4f36acf12e6d491484812927b327c066***

**Get all EKS Anywhere license details with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).

```bash
aws license-manager list-received-licenses \
  --region 'region-code' \
  --filter 'Name=IssuerName,Values=Amazon EKS Anywhere'
```

**Get license details with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-license-arn` with the license ARN returned from the previous command.

```bash
aws license-manager get-license \
  --region 'region-code' \
  --license-arn 'my-license-arn'
```

<details>
  <summary>Expand for sample command output</summary>
  <br /> 
  {{% content "get-license-output.md" %}}
</details>

### Accept EKS Anywhere license grant

You can accept the license grants associated with an EKS Anywhere Enterprise Subscription in the [AWS License Manager Console](https://console.aws.amazon.com/license-manager/) following the instructions in the [AWS License Manager documentation](https://docs.aws.amazon.com/license-manager/latest/userguide/granted-licenses.html). Navigate to the license for your subscription and client **Accept and Activate** in the top right of the license detail page.

See the steps below for accepting EKS Anywhere license grants with the AWS CLI.

**Get license grant ARNs with subscription `name` with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription` in the `--query` string with the `name` for your subscription.

```bash
aws license-manager list-received-licenses \
  --region 'region-code' \
  --filter 'Name=IssuerName,Values=Amazon EKS Anywhere' \
  --query 'Licenses[?LicenseName==`EKS Anywhere license for subscription my-subscription`].LicenseMetadata[].Value'
```

**Accept the license grant with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-grant-arn` with the grant ARN returned from the previous command. If you have multiple grants, repeat for each grant ARN.

```bash
aws license-manager accept-grant \
  --region 'region-code' \
  --grant-arn 'my-grant-arn'
```

**Activate license grant with the AWS CLI**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-grant-arn` with the grant ARN returned from the previous command. If you have multiple grants, repeat for each grant ARN.
- Replace `my-client-token` with a unique, case-sensitive identifier that you provide to ensure the idempotency of the request (for example `e75f7f81-1b0b-47b4-85b4-5cbeb7ffb921`).

```bash
aws license-manager create-grant-version \
  --region 'region-code' \
  --grant-arn 'my-grant-arn' \
  --status 'ACTIVE' \
  --client-token 'my-client-token'
```

### Apply a license to an EKS Anywhere cluster

You can apply a license to an EKS Anywhere cluster during or after cluster creation for standalone or management clusters. For workload clusters, you must apply the license after cluster creation. A license can only be bound to one EKS Anywhere cluster at a time, and you can only receive support for your EKS Anywhere cluster if it has a valid and active license. In the examples below, the `<license-id-string>` is the License ID, for example `l-93ea2875c88f455288737835fa0abbc8`.

To apply a license during standalone or management cluster creation, export the `EKSA_LICENSE` environment variable before running the `eksctl anywhere create cluster` command.

```bash
export EKSA_LICENSE='<license-id-string>'
```

To apply a license to an existing cluster, apply the following Secret to your cluster, replacing `<license-id-string>` with your License ID.

   ```bash
   kubectl apply -f - <<EOF 
   apiVersion: v1
   kind: Secret
   metadata:
     name: eksa-license
     namespace: eksa-system
   stringData:
     license: "<license-id-string>"
   type: Opaque
   EOF
   ```
