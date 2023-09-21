---
title: "Purchase EKS Anywhere Enterprise Subscriptions"
linkTitle: "Purchase subscription"
weight: 5
date: 2023-09-21
description: >
  Steps to purchase an EKS Anywhere Enterprise Subscription
---

You can purchase EKS Anywhere Enterprise Subscriptions with the Amazon EKS console, API, or AWS CLI. When you purchase a subscription, you can choose a 1-year term or a 3-year term, and you are billed monthly throughout the term. You can configure your subscription to automatically renew at the end of the term, and you can cancel your subscription within the first 7 days of purchase at no charge. When the status of your subscription is Active, the term starts, licenses are available in AWS License Manager for your EKS Anywhere clusters, and your AWS account has access to Amazon EKS Anywhere Curated Packages. For pricing, reference the [EKS Anywhere Pricing Page.](https://aws.amazon.com/eks/eks-anywhere/pricing/)

## Purchase Subscriptions

>**_NOTE_** When you purchase the subscription, you have a 7-day grace period to cancel the contract by creating a ticket at [AWS Support Center.](https://console.aws.amazon.com/support/home) After that, your AWS account ID is invoiced. Payment is charged monthly. Only auto-renewal and tags can be changed after subscription creation. Other attributes such as the subscription name, number of licenses, or term length cannot be modified after creation.

### AWS Management Console

1. Open the Amazon EKS console at https://console.aws.amazon.com/eks/home#/eks-anywhere
1. Click the **Subscribe** button on the right side of the screen.
1. On the **Specify subscription details** page, select an offer (1 year term or 3 year term)
1. Configure the following fields

  - **Name** - a name for your subscription. It must be unique in your AWS account in the AWS Region you're creating the subscription in. The name can contain only alphanumeric characters (case-sensitive), hyphens, and underscores. It must start with an alphabetic character and can't be longer than 100 characters.
  - **Number of licenses** - the number of licenses to include in the subscription. This value cannot be changed after creating the subscription.
  - **Auto-renewal** - if enabled, the subscription will automatically renew at the end of the term.

5. (Optional) **Configure tags**. A tag is a label that you assign to an EKS Anywhere subscription. Each tag consists of a key and an optional value. You can use tags to search and filter your resources.
6. Click **Next**
7. On the **Review and purchase** page, confirm the specifications for your subscription are correct.
8. Click **Create** on the bottom right hand side of the screen to purchase your subscription. 

After the subscription is created, the next step is to license your EKS Anywhere clusters. Reference the [License EKS Anywhere Cluster page]({{< relref "./cluster-license">}}) for instructions.

### AWS CLI

To install or update the AWS CLI, reference the [AWS documentation.](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) You must be using at least version **TODO** to create EKS Anywhere subscriptions.

Create your subscription with the following command. Before running the command, make the following replacements:

- Replace `region-code` with the AWS Region that you want to create you subscription in. It is recommended to create your subscription in the AWS Region closest to your on-premises deployment.
- Replace `my-subscription` with a name for your subscription.  It must be unique in your AWS account in the AWS Region you're creating the subscription in. The name can contain only alphanumeric characters (case-sensitive), hyphens, and underscores. It must start with an alphabetic character and can't be longer than 100 characters.
- Replace `license-quantity 1` with the number of licenses to include in the subscription
- Replace `term 'unit=MONTHS,duration=12'` with your preferred term length. Valid options for `duration` are `12` and `36`. The only accepted unit is `months`.
- Replace `tags 'environment=prod'` with your preferred tags for your subscription.
- Optionally enable auto-renewal with the `--auto-renew` flag. Subscriptions will not auto-renew by default.


```bash
aws eks create-eks-anywhere-subscription \
  --region 'us-west-2' \
  --name 'my-subscription' \
  --license-quantity 1 \
  --term 'unit=MONTHS,duration=12' \
  --tags 'environment=prod' \
  --no-auto-renew
```

After the subscription is created, the next step is to license your EKS Anywhere clusters. Reference the [License EKS Anywhere Cluster page]({{< relref "./cluster-license">}}) for instructions.

## View and Update Subscriptions

**TODO**

## Delete Subscriptions
