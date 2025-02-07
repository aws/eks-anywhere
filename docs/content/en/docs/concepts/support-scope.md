---
title: "Support for EKS Anywhere"
linkTitle: "Support"
weight: 40
date: 2023-09-21
aliases:
    /docs/reference/support/
    /docs/reference/support/support-scope/
description: >
  Overview of EKS Anywhere Enterprise Subscriptions
---

EKS Anywhere is available as open source software that you can run on hardware in your data center or edge environment. 

You can purchase EKS Anywhere Enterprise Subscriptions to receive support for your EKS Anywhere clusters and for access to [EKS Anywhere Curated Packages]({{< relref "./packages/">}}) and [extended support for Kubernetes versions]({{< relref "./support-versions">}}). You can only receive support for your EKS Anywhere clusters that are licensed under an active EKS Anywhere Enterprise Subscription. EKS Anywhere Enterprise Subscriptions are available for a 1-year or 3-year term, and are priced on a per cluster basis.

You must have an EKS Anywhere Enterprise Subscription to access EKS Anywhere Curated Packages and EKS Anywhere extended support for Kubernetes versions.

EKS Anywhere Enterprise Subscriptions include support for the following components. 

- EKS Distro (see [documentation](https://distro.eks.amazonaws.com/) for components)
- EKS Anywhere core components such as the Cilium CNI, Flux GitOps controller, kube-vip, EKS Anywhere CLI, EKS Anywhere controllers, image builder, and EKS Connector
- EKS Anywhere cluster lifecycle operations such as creating, scaling, and upgrading
- EKS Anywhere troubleshooting, general guidance, and best practices
- EKS Anywhere Curated Packages (see [Overview of curated packages]({{< relref "packages">}}) for more information)
- EKS Anywhere extended support for Kubernetes versions (see [Version lifecycle]({{< relref "support-versions">}}) for more information)
- Bottlerocket node operating system

Visit the following links for more information on EKS Anywhere Enterprise Subscriptions

- [EKS Anywhere Pricing Page](https://aws.amazon.com/eks/eks-anywhere/pricing/)
- [EKS Anywhere FAQ Page](https://aws.amazon.com/eks/eks-anywhere/faqs/)
- [Steps to purchase a subscription]({{< relref "../clustermgmt/support/purchase-subscription" >}})
- [Steps to license your cluster]({{< relref "../clustermgmt/support/cluster-license" >}})
- [Steps to share curated packages with another account]({{< relref "../clustermgmt/support/share-packages" >}})

If you are using EKS Anywhere and have not purchased a subscription, you can file an issue in the [EKS Anywhere GitHub Repository,](https://github.com/aws/eks-anywhere/issues) and someone will get back to you as soon as possible. If you discover a potential security issue in this project, we ask that you notify AWS/Amazon Security via the [vulnerability reporting page.](https://aws.amazon.com/security/vulnerability-reporting/) Please do not create a public GitHub issue for security problems.

### Frequently Asked Questions (FAQs)

**1. How much does an EKS Anywhere Enterprise Subscription cost?**

For pricing information, visit the [EKS Anywhere Pricing](https://aws.amazon.com/eks/eks-anywhere/pricing/) page.

**2. How can I purchase an EKS Anywhere Enterprise Subscription?**

Reference the [Purchase Subscriptions]({{< relref "../clustermgmt/support/purchase-subscription" >}}) documentation for instructions on how to purchase.

**3. Are subscriptions I previously purchased manually integrated into the EKS console?**

No, EKS Anywhere Enterprise Subscriptions purchased manually before October 2023 cannot be viewed or managed through the EKS console, APIs, and AWS CLI. 

**4. Can I cancel my subscription in the EKS console, APIs, and AWS CLI?**

You can cancel your subscription within the first 7 days of purchase by filing an AWS Support case. When you cancel your subscription within the first 7 days, you are not charged for the subscription. To cancel your subscription outside of the 7-day time period, contact your AWS account team.

**5. Can I cancel my subscription after I use it to file an AWS Support case?**

No, if you have used your subscription to file an AWS Support case requesting EKS Anywhere support, then we are unable to cancel the subscription or refund the purchase regardless of the 7-day grace period, since you have leveraged support as part of the subscription.

**6. In which AWS Regions can I purchase subscriptions?**

You can purchase subscriptions in all [AWS Regions](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/), except the Asia Pacific (Thailand), Mexico (Central), AWS GovCloud (US) Regions, and the China Regions.

**7. Can I renew my subscription through the EKS console, APIs, and AWS CLI?**

Yes, you can configure auto renewal during subscription creation or at any time during your subscription term. When auto renewal is enabled for your subscription, the subscription and associated licenses will be automatically renewed for the term of the existing subscription (1-year or 3-years). The 7-day cancellation period does not apply to renewals. You do not need to reapply licenses to your EKS Anywhere clusters when subscriptions are automatically renewed.

**8. Can I edit my subscription through the EKS console, APIs, and AWS CLI?**

You can edit the auto renewal and tags configurations for your subscription with the EKS console, APIs, and AWS CLI. To change the term or license quantity for a subscription, you must create a new subscription.

**9. What happens when a subscription expires?**

When subscriptions expire, licenses associated with the subscription can no longer be used for new support cases, access to EKS Anywhere Curated Packages is revoked, and you are no longer billed for the subscription. Support caes created during the active subscription period will continue to be serviced. You will receive emails 3 months, 1 month, and 1 week before subscriptions expire, and an alert is presented in the EKS console for approaching expiration dates. Subscriptions can be viewed with the EKS console, APIs, and AWS CLI after expiration.

**10. How do I use the licenses for my subscription?**

When you create a subscription, licenses are created based on the quantity you pass when you create the subscription and the licenses can be viewed in the EKS console. There are two key parts of the license, the license ID string and the license token. In EKS Anywhere versions v0.21 and older, the license ID string was applied as a Kubernetes Secret to EKS Anywhere clusters and used when support cases are created to validate the cluster is eligible for support. The license token was introduced in EKS Anywhere version v0.22 and all existing EKS Anywhere subscriptions have been updated with a license token for each license. 

You can use either the license ID string or the license token when you create AWS Support cases for your EKS Anywhere clusters. To use extended support for Kubernetes versions in EKS Anywhere, available for EKS Anywhere version v0.22 and above, your clusters must have a valid and active license token to be able to create and upgrade clusters using the Kubernetes extended support versions.

**11. How do I apply licenses to my EKS Anywhere clusters?**

Reference the [License cluster]({{< relref "../clustermgmt/support/cluster-license" >}}) documentation for instructions on how to apply licenses your EKS Anywhere clusters.

**12. Can I share licenses with other AWS accounts?**

The licenses created for your subscriptions cannot be shared with AWS Resource Access Manager but they can be shared through your own manual, secure mechanisms. It is generally recommended to create the subscription with the AWS account that will be used to operate the EKS Anywhere clusters. 

**13. Can I share access to curated packages with other AWS accounts?**

Yes, reference the [Share curated packages access]({{< relref "../clustermgmt/support/share-packages" >}}) documentation for instructions on how to share access to curated packages with other AWS accounts in your organization.

**14. Is there an option to pay for subscriptions upfront?**

If you need to pay upfront for subscriptions, please contact your AWS account team.

**15. Is there a free trial for EKS Anywhere Enterprise Subscriptions?**

Free trial access to EKS Anywhere Curated Packages is available upon request. Free trial access to EKS Anywhere Curated Packages does not include troubleshooting support for your EKS Anywhere clusters or access to extended support for Kubernetes versions. Contact your AWS account team for more information.
