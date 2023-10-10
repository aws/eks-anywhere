---
title: "Share access to EKS Anywhere Curated Packages"
linkTitle: "Share curated packages access"
weight: 25
date: 2023-09-21
description: >
  Share access to EKS Anywhere Curated Packages with other AWS accounts
---

When an EKS Anywhere Enterprise Subscription is created, the AWS account that created the subscription is granted access to EKS Anywhere Curated Packages in the AWS Region where the subscription is created. To enable access to EKS Anywhere Curated Packages for other AWS accounts in your organization, follow the instructions below. The instructions below use `111111111111` as the source account, and `999999999999` as the destination account.

### 1. Save EKS Anywhere Curated Packages registry account for your subscription

In the following step, you will need the AWS account ID associated with your subscription's curated packages ECR registry. Run the following command and save the 12-digit AWS account ID from the output string.

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription-id` with the `id` for your subscription (for example `e29fd0d2-d8a8-4ed4-be54-c6c0dd0f7964`). 

```bash
aws eks describe-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id' \
  --query 'subscription.packageRegistry'
```

### 2. Create an IAM Policy with ECR Login and Read permissions

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Policies** and then choose **Create policy**
1. On the **Specify permissions** page, select **JSON**
1. Paste the following permission specification into the **Policy editor**. Replace `067575901363` in the permission specification with the package registry AWS account you saved in the previous step.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ECRRead",
      "Effect": "Allow",
      "Action": [
        "ecr:DescribeImageScanFindings",
        "ecr:GetDownloadUrlForLayer",
        "ecr:DescribeRegistry",
        "ecr:DescribePullThroughCacheRules",
        "ecr:DescribeImageReplicationStatus",
        "ecr:ListTagsForResource",
        "ecr:ListImages",
        "ecr:BatchGetImage",
        "ecr:DescribeImages",
        "ecr:DescribeRepositories",
        "ecr:BatchCheckLayerAvailability"
      ],
      "Resource": "arn:aws:ecr:*:067575901363:repository/*"
    },
    {
      "Sid": "ECRLogin",
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken"
      ],
      "Resource": "*"
    }
  ]
}
``` 

1. Choose **Next**
1. On the **Review and create** page, enter a **Policy name** such as `curated-packages-policy`
1. Choose **Create policy**

### 3. Create IAM role with permissions for EKS Anywhere Curated Packages

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Roles** and then choose **Create role**
1. On the **Select trusted entity** page, choose **Custom trust policy** as the **Trusted entity type**. Add the following trust policy, replacing `999999999999` with the AWS account receiving permissions.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::999999999999:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

4. Choose **Next**
5. On the **Add permissions** page, search and select the policy you created in the previous step. 
6. Choose **Next**
7. On the **Name, review, and create** page, enter a **Role name** such as `curated-packages-role`
8. Choose **Create role**

### 4. Create IAM user with permissions to assume the IAM role

**Create policy to assume role**

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Policies** and then choose **Create policy**
1. On the **Specify permissions** page, select **JSON**
1. Paste the following permission specification into the **Policy editor**. Replace `111111111111` in the permission specification with the package registry AWS account you saved in the previous step, and `curated-packages-role` with the name of the role you created in the previous step.

```json
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "sts:AssumeRole",
    "Resource": "arn:aws:iam::111111111111:role/curated-packages-role"
  }
}
```

5. Choose **Next**
6. On the **Review and create** page, enter a **Policy name** such as `curated-packages-assume-role-policy`
7. Choose **Create policy**

**Create user to assume IAM role**

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Users** and then choose **Create user**
1. Enter a **User name** such as `curated-packages-user`
1. Choose **Next**
1. On the **Set permissions** page, choose **Attach policies directly**, and search and select the assume role policy you created above.
1. Choose **Next**
1. On the **Review and create** page, choose **Create user**

### 5. Generate access and secret key

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Users** and the user you created in the previous step.
1. On the users detail page in the top **Summary** section, choose **Create access key** under **Access key 1**
1. On the **Access key best practices & alternatives** page, select **Command Line Interface (CLI)**
1. Confirm that you understand the recommendation and want to proceed to create an access key. Choose **Next**.
1. On the **Set description tag** page, choose **Create access key**
1. On the **Retrieve access keys** page, copy the **Access key** and **Secret access key** to a safe location.
1. Choose **Done**

### 6. Add the Access key and Secret access key to your EKS Anywhere cluster



