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

In this step, you will get the Amazon ECR packages registry account associated with your subscription. Run the following command with the account that created the subscription and save the 12-digit account ID from the output string. 

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription-id` with the `id` for your subscription (for example `e29fd0d2-d8a8-4ed4-be54-c6c0dd0f7964`). 

```bash
aws eks describe-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id' \
  --query 'subscription.packageRegistry'
```

The output has the following structure: "<packages-account-id>.dkr.ecr.<region>.amazonaws.com". Save the `<packages-account-id>` for the next step.

Alternatively, you can use the following table to identify the packages registry account for the AWS Region hosting your subscription.

<details>
  <summary>Expand for packages registry to AWS Region table</summary>
  <br /> 
  {{% content "packages-registries.md" %}}
</details>
<br /> 

### 2. Create an IAM Policy with ECR Login and Read permissions

Run the following with the account that created the subscription (in this example `111111111111`).

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Policies** and then choose **Create policy**
1. On the **Specify permissions** page, select **JSON**
1. Paste the following permission specification into the **Policy editor**. Replace `<packages-account-id>` in the permission specification with the account you saved in the previous step.

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
      "Resource": "arn:aws:ecr:*:<packages-account-id>:repository/*"
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

### 3. Create an IAM role with permissions for EKS Anywhere Curated Packages

Run the following with the account that created the subscription.

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Roles** and then choose **Create role**
1. On the **Select trusted entity** page, choose **Custom trust policy** as the **Trusted entity type**. Add the following trust policy, replacing `999999999999` with the AWS account receiving permissions. This policy enables account `999999999999` to assume the role.

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
5. On the **Add permissions** page, search and select the policy you created in the previous step (for example `curated-packages-policy`). 
6. Choose **Next**
7. On the **Name, review, and create** page, enter a **Role name** such as `curated-packages-role`
8. Choose **Create role**

### 4. Create an IAM user with permissions to assume the IAM role from the source account

Run the following with the account that is receiving access to curated packages (in this example `999999999999`) .

**Create a policy to assume the IAM role**

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Policies** and then choose **Create policy**
1. On the **Specify permissions** page, select **JSON**
1. Paste the following permission specification into the **Policy editor**. Replace `111111111111` with the account used to create the subscription, and `curated-packages-role` with the name of the role you created in the previous step.

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

**Create an IAM user to assume the IAM role**

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Users** and then choose **Create user**
1. Enter a **User name** such as `curated-packages-user`
1. Choose **Next**
1. On the **Set permissions** page, choose **Attach policies directly**, and search and select the assume role policy you created above.
1. Choose **Next**
1. On the **Review and create** page, choose **Create user**

### 5. Generate access and secret key for IAM user

Run the following with the account that is receiving access to curated packages.

1. Open the [IAM console](https://console.aws.amazon.com/iam/)
1. In the navigation pane, choose **Users** and the user you created in the previous step.
1. On the users detail page in the top **Summary** section, choose **Create access key** under **Access key 1**
1. On the **Access key best practices & alternatives** page, select **Command Line Interface (CLI)**
1. Confirm that you understand the recommendation and want to proceed to create an access key. Choose **Next**.
1. On the **Set description tag** page, choose **Create access key**
1. On the **Retrieve access keys** page, copy the **Access key** and **Secret access key** to a safe location.
1. Choose **Done**

### 6. Create an AWS config file for IAM user

Run the following with the account that is receiving access to curated packages.

Create an AWS config file with the assumed role and the access/secret key you generated in the previous step. Replace the values in the example below based on your configuration.

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `role-arn` with the role you created in **Step 3**
- Replace `aws_access_key_id` and `aws_secret_access_key` that you created in **Step 5**

```
[default]
source_profile=curated-packages-user
role_arn=arn:aws:iam::111111111111:role/curated-packages-role
region=region-code

[profile curated-packages-user]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

### 7. Add the AWS config to your EKS Anywhere cluster

Run the following with the account that is receiving access to curated packages.

**New Clusters**

For new standalone or management clusters, pass the AWS config file path that you created in the previous step as the `EKSA_AWS_CONFIG_FILE` environment variable. The EKS Anywhere CLI detects the environment variable when you run `eksctl anywhere create cluster`. Note, the credentials are used by the Curated Packages Controller, which should only run on standalone or management clusters.

**Existing Clusters**

For existing standalone or management clusters, the AWS config information will be passed as a Kubernetes Secret. You need to generate the base64 encoded string from the AWS config file and then pass the encoded string in the `config` field of the `aws-secret` Secret in the `eksa-packages` namespace.

Encode the AWS config file. Replace `<aws-config-file>` with the name of the file you created in the previous step.

```bash
cat <aws-config-file> | base64
```

Create a yaml specification called `aws-secret.yaml`, replacing `<encoded-aws-config-file>` with the encoded output from the previous step.

```
apiVersion: v1
kind: Secret
metadata:
  name: aws-secret
  namespace: eksa-packages
type: Opaque
data:
  AWS_ACCESS_KEY_ID: ""
  AWS_SECRET_ACCESS_KEY: ""
  REGION: ""
  config: <encoded-aws-config-file>
```

Apply the Secret to your standalone or management cluster.

```bash
kubectl apply -f aws-secret.yaml
```
