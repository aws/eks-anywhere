---
title: "Purchase EKS Anywhere Enterprise Subscriptions"
linkTitle: "Purchase subscriptions"
weight: 5
date: 2023-09-21
description: >
  Steps to purchase an EKS Anywhere Enterprise Subscription
---

You can purchase EKS Anywhere Enterprise Subscriptions with the Amazon EKS console, API, or AWS CLI. When you purchase a subscription, you can choose a 1-year term or a 3-year term, and you are billed monthly throughout the term. You can configure your subscription to automatically renew at the end of the term, and you can cancel your subscription within the first 7 days of purchase at no charge. When the status of your subscription is Active, the subscription term starts, licenses are available in AWS License Manager for your EKS Anywhere clusters, and your AWS account has access to Amazon EKS Anywhere Curated Packages. 

For pricing, reference the [EKS Anywhere Pricing Page.](https://aws.amazon.com/eks/eks-anywhere/pricing/)

## Create Subscriptions

>**_NOTE_** When you purchase the subscription, you have a 7-day grace period to cancel the contract by creating a ticket at [AWS Support Center.](https://console.aws.amazon.com/support/home) After the 7-day grace period, if you do not cancel the contract, your AWS account ID is invoiced. Payment is charged monthly. 

### Prerequisites

- Before you create a subscription, you must onboard to use AWS License Manager. See the [AWS License Manager](https://docs.aws.amazon.com/license-manager/latest/userguide/getting-started.html) documentation for instructions.
- Only auto renewal and tags can be changed after subscription creation. Other attributes such as the subscription name, number of licenses, or term length cannot be modified after subscription creation.
- You can purchase Amazon EKS Anywhere Enterprise Subscriptions in the following AWS Regions: US East (Ohio), US East (N. Virginia), US West (N. California), US West (Oregon), Africa (Cape Town), Asia Pacific (Hong Kong), Asia Pacific (Hyderabad), Asia Pacific (Jakarta), Asia Pacific (Melbourne), Asia Pacific (Mumbai), Asia Pacific (Osaka), Asia Pacific (Seoul), Asia Pacific (Singapore), Asia Pacific (Sydney), Asia Pacific (Tokyo), Canada (Central), Europe (Frankfurt), Europe (Ireland), Europe (London), Europe (Milan), Europe (Paris), Europe (Stockholm), Europe (Zurich), Israel (Tel Aviv), Middle East (Bahrain), Middle East (UAE), and South America (Sao Paulo). 
- An individual subscription can have up to 100 licenses.
- An individual account can have up to 10 subscriptions.
- You can create a single subscription at a time.

### AWS Management Console

1. Open the Amazon EKS console at https://console.aws.amazon.com/eks/home#/eks-anywhere.
1. Click the **Create subscription** button on the right side of the screen.
1. On the **Specify subscription details** page, select an offer (1 year term or 3 year term).
1. Configure the following fields:

  - **Name** - a name for your subscription. It must be unique in your AWS account in the AWS Region you're creating the subscription in. The name can contain only alphanumeric characters (case-sensitive), hyphens, and underscores. It must start with an alphabetic character and can't be longer than 100 characters. This value cannot be changed after creating the subscription.
  - **Number of licenses** - the number of licenses to include in the subscription. This value cannot be changed after creating the subscription.
  - **Auto renewal** - if enabled, the subscription will automatically renew at the end of the term.

5. (Optional) **Configure tags**. A tag is a label that you assign to an EKS Anywhere subscription. Each tag consists of a key and an optional value. You can use tags to search and filter your resources.
6. Click **Next**.
7. On the **Review and purchase** page, confirm the specifications for your subscription are correct.
8. Click **Purchase** on the bottom right hand side of the screen to purchase your subscription. 

After the subscription is created, the next step is to apply the licenses to your EKS Anywhere clusters. Reference the [License cluster]({{< relref "./cluster-license">}}) page for instructions.

### AWS CLI

To install or update the AWS CLI, reference the [AWS documentation.](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) If you already have the AWS CLI installed, update to the latest version of the CLI before running the following commands.

Create your subscription with the following command. Before running the command, make the following replacements:

- Replace `region-code` with the AWS Region that will host your subscription (for example `us-west-2`). It is recommended to create your subscription in the AWS Region closest to your on-premises deployment.
- Replace `my-subscription` with a name for your subscription.  It must be unique in your AWS account in the AWS Region you're creating the subscription in. The name can contain only alphanumeric characters (case-sensitive), hyphens, and underscores. It must start with an alphabetic character and can't be longer than 100 characters.
- Replace `license-quantity` `1` with the number of licenses to include in the subscription.
- Replace `term` `'unit=MONTHS,duration=12'` with your preferred term length. Valid options for `duration` are `12` and `36`. The only accepted `unit` is `MONTHS`.
- Optionally, replace `tags` `'environment=prod'` with your preferred tags for your subscription.
- Optionally, enable auto renewal with the `--auto-renew` flag. Subscriptions will not auto renew by default.


```bash
aws eks create-eks-anywhere-subscription \
  --region 'region-code' \
  --name 'my-subscription' \
  --license-quantity 1 \
  --term 'unit=MONTHS,duration=12' \
  --tags 'environment=prod' \
  --no-auto-renew
```

<details>
  <summary>Expand for sample command output</summary>
  <br /> 
  {{% content "create-subscription-output.md" %}}
</details>
<br /> 

It may take several minutes for the subscription to become `ACTIVE`. You can query the status of your subscription with the following command. Replace `my-subscription-id` with the `id` of your subscription. Do not proceed to license your EKS Anywhere clusters until the output of the command returns `ACTIVE`.

```bash
aws eks describe-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id' \
  --query 'subscription.status'
```

After the subscription is created, the next step is to apply the licenses to your EKS Anywhere clusters. Reference the [License cluster]({{< relref "./cluster-license">}}) page for instructions.

## View and Update Subscriptions

After you create a subscription, you can only update the auto renewal and tags configurations.

### AWS Management Console

1. Open the Amazon EKS console at https://console.aws.amazon.com/eks/home#/eks-anywhere.
1. Navigate to the **Active Subscriptions** or **Inactive Subscriptions** tab.
1. Optionally, choose the selection button for your EKS Anywhere subscription and click the **Change auto renewal** button to change your auto renewal setting.
1. Click the link of your EKS Anywhere subscription name to view details including subscription start and end dates, associated licenses, and tags.
1. Optionally, edit tags by clicking the **Manage Tags** button.

### AWS CLI

**List EKS Anywhere subscriptions**

- Replace `region-code` with the AWS Region that hosts your subscription(s) (for example `us-west-2`).

```bash
aws eks list-eks-anywhere-subscriptions --region 'region-code'
```

<details>
  <summary>Expand for sample command output</summary>
  <br /> 
  {{% content "list-subscription-output.md" %}}
</details>
<br /> 

**Describe EKS Anywhere subscriptions**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription-id` with the `id` for your subscription (for example `e29fd0d2-d8a8-4ed4-be54-c6c0dd0f7964`). 
- Replace `my-subscription` with the `name` for your subscription.

Get subscription details for a single subscription.

```bash
aws eks describe-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id'
```

<details>
  <summary>Expand for sample command output</summary>
  <br /> 
  {{% content "describe-subscription-output.md" %}}
</details>
<br /> 

Get subscription `id` with subscription `name`.

```bash
aws eks list-eks-anywhere-subscriptions \
  --region 'region-code' \
  --query 'subscriptions[?name==`my-subscription`].id'
```

Get subscription `arn` with subscription `name`.

```bash
aws eks list-eks-anywhere-subscriptions \
  --region 'region-code' \
  --query 'subscriptions[?name==`my-subscription`].arn'
```

**Update EKS Anywhere subscriptions**

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription-id` with the `id` for your subscription (for example `e29fd0d2-d8a8-4ed4-be54-c6c0dd0f7964`). 

Disable auto renewal

```bash
aws eks update-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id' \
  --no-auto-renew
```

Enable auto renewal

```bash
aws eks update-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id' \
  --auto-renew
```

Update tags

```bash
aws eks tag-resource \
  --region 'region-code' \
  --resource-arn 'my-subscription-arn' \
  --tags 'geo=boston'
```

## Delete Subscriptions

>**_NOTE_** Only inactive subscriptions can be deleted. Deleting inactive subscriptions removes them from the AWS Management Console view and API responses.

### AWS Management Console

1. Open the Amazon EKS console at https://console.aws.amazon.com/eks/home#/eks-anywhere.
1. Click the **Inactive Subscriptions** tab.
1. Choose the name of the EKS Anywhere subscription to delete and click the **Delete subscription**.
1. On the delete subscription confirmation screen, choose **Delete**.

### AWS CLI

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription-id` with the `id` for your subscription (for example `e29fd0d2-d8a8-4ed4-be54-c6c0dd0f7964`). 

```bash
aws eks delete-eks-anywhere-subscription \
  --region 'region-code' \
  --id 'my-subscription-id'
```
