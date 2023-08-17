---
title: "License EKS Anywhere cluster"
linkTitle: "License cluster"
weight: 20
date: 2023-08-11
aliases:
    /docs/tasks/cluster/cluster-license/
description: >
  Apply an EKS Anywhere Enterprise Subscription license to your EKS Anywhere cluster.
---

EKS Anywhere is open source and free to use at no cost. To receive support for your EKS Anywhere clusters, you can optionally purchase [EKS Anywhere Enterprise Subscriptions]({{< relref "../../concepts/support-scope/">}}) for 24/7 support from AWS subject matter experts and access to [EKS Anywhere Curated Packages]({{< relref "../../concepts/packages/">}}).

When you purchase an EKS Anywhere Enterprise Subscription, licenses for your clusters are provisioned in [AWS License Manager](https://docs.aws.amazon.com/license-manager/latest/userguide/license-manager.html) in the AWS account you used to purchase the subscription. After purchasing your subscription, navigate to the AWS License Manager console and accept the license grants following the steps in the [AWS License Manager documentation](https://docs.aws.amazon.com/license-manager/latest/userguide/granted-licenses.html). Save the License ID strings for your licenses, as you will need them to license your clusters.

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
