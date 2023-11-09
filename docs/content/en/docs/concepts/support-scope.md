---
title: "Support"
linkTitle: "Support"
weight: 40
date: 2023-09-21
aliases:
    /docs/reference/support/
    /docs/reference/support/support-scope/
description: >
  Overview of support for EKS Anywhere
---

EKS Anywhere is available as open source software that you can run on hardware in your data center or edge environment. 

You can purchase EKS Anywhere Enterprise Subscriptions for 24/7 support from AWS subject matter experts and access to EKS Anywhere Curated Packages. You can only receive support for your EKS Anywhere clusters that are licensed under an active EKS Anywhere Enterprise Subscription. EKS Anywhere Enterprise Subscriptions are available for a 1-year or 3-year term, and are priced on a per cluster basis.

EKS Anywhere Enterprise Subscriptions include support for the following components:

- EKS Distro (see [documentation](https://distro.eks.aws.com/) for components)
- EKS Anywhere core components such as the Cilium CNI, Flux GitOps controller, kube-vip, EKS Anywhere CLI, EKS Anywhere controllers, image builder, and EKS Connector
- EKS Anywhere Curated Packages (see [curated packages list]({{< relref "../concepts/packages/#curated-package-list" >}}) for list of packages) 
- EKS Anywhere cluster lifecycle operations such as creating, scaling, and upgrading
- EKS Anywhere troubleshooting, general guidance, and best practices
- Bottlerocket node operating system

Visit the following links for more information on EKS Anywhere Enterprise Subscriptions

- [EKS Anywhere Pricing Page](https://aws.amazon.com/eks/eks-anywhere/pricing/)
- [EKS Anywhere FAQ Page](https://aws.amazon.com/eks/eks-anywhere/faqs/)
- [Steps to purchase a subscription]({{< relref "../clustermgmt/support/purchase-subscription" >}})
- [Steps to license your cluster]({{< relref "../clustermgmt/support/cluster-license" >}})
- [Steps to share curated packages with another account]({{< relref "../clustermgmt/support/share-packages" >}})

If you are using EKS Anywhere and have not purchased a subscription, you can file an [issue](https://github.com/aws/eks-anywhere/issues) in the EKS Anywhere GitHub Repository, and someone will get back to you as soon as possible. If you discover a potential security issue in this project, we ask that you notify AWS/Amazon Security via the [vulnerability reporting page.](http://aws..com/security/vulnerability-reporting/) Please do not create a public GitHub issue for security problems.

### FAQs

**1. How much does an EKS Anywhere Enterprise Subscription cost?**

For pricing information, visit the [EKS Anywhere Pricing](https://aws.amazon.com/eks/eks-anywhere/pricing/) page.

**2. How can I purchase an EKS Anywhere Enterprise Subscription?**

Reference the [Purchase Subscriptions]({{< relref "../clustermgmt/support/purchase-subscription" >}}) documentation for instructions on how to purchase.

**3. Are subscriptions I previously purchased manually integrated into the EKS console?**

No, EKS Anywhere Enterprise Subscriptions purchased manually before October 2023 cannot be viewed or managed through the EKS console, APIs, and AWS CLI. 

**4. Can I cancel my subscription in the EKS console, APIs, and AWS CLI?**

You can cancel your subscription within the first 7 days of purchase by filing an AWS Support ticket. When you cancel your subscription within the first 7 days, you are not charged for the subscription. To cancel your subscription outside of the 7-day time period, contact your AWS account team.

**5. In which AWS Regions can I purchase subscriptions?**

You can purchase subscriptions in US East (Ohio), US East (N. Virginia), US West (N. California), US West (Oregon), Africa (Cape Town), Asia Pacific (Hong Kong), Asia Pacific (Hyderabad), Asia Pacific (Jakarta), Asia Pacific (Melbourne), Asia Pacific (Mumbai), Asia Pacific (Osaka), Asia Pacific (Seoul), Asia Pacific (Singapore), Asia Pacific (Sydney), Asia Pacific (Tokyo), Canada (Central), Europe (Frankfurt), Europe (Ireland), Europe (London), Europe (Milan), Europe (Paris), Europe (Stockholm), Europe (Zurich), Israel (Tel Aviv), Middle East (Bahrain), Middle East (UAE), and South America (Sao Paulo).

**6. Can I renew my subscription through the EKS console, APIs, and AWS CLI?**

Yes, you can configure auto renewal during subscription creation or at any time during your subscription term. When auto renewal is enabled for your subscription, the subscription and associated licenses will be automatically renewed for the term of the existing subscription (1-year or 3-years). The 7-day cancellation period does not apply to renewals. You do not need to reapply licenses to your EKS Anywhere clusters when subscriptions are automatically renewed.

**7. Can I edit my subscription through the EKS console, APIs, and AWS CLI?**

You can edit the auto renewal and tags configurations for your subscription with the EKS console, APIs, and AWS CLI. To change the term or license quantity for a subscription, you must create a new subscription.

**8. What happens when a subscription expires?**

When subscriptions expire, licenses associated with the subscription can no longer be used for new support tickets, access to EKS Anywhere Curated Packages is revoked, and you are no longer billed for the subscription. Support tickets created during the active subscription period will continue to be serviced. You will receive emails 3 months, 1 month, and 1 week before subscriptions expire, and an alert is presented in the EKS console for approaching expiration dates. Subscriptions can be viewed with the EKS console, APIs, and AWS CLI after expiration.

**9. Can I share access to curated packages with other AWS accounts?**

Yes, reference the [Share curated packages access]({{< relref "../clustermgmt/support/share-packages" >}}) documentation for instructions on how to share access to curated packages with other AWS accounts in your organization.

**10. How do I apply licenses to my EKS Anywhere clusters?**

Reference the [License cluster]({{< relref "../clustermgmt/support/cluster-license" >}}) documentation for instructions on how to apply licenses your EKS Anywhere clusters.

**11. Is there an option to pay for subscriptions upfront?**

If you need to pay upfront for subscriptions, please contact your AWS account team.

**12. Is there a free-trial option for subscriptions?**

To request a free-trial, please contact your AWS account team.
