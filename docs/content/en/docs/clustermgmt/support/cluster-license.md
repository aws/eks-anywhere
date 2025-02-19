---
title: "License EKS Anywhere cluster"
linkTitle: "License cluster"
weight: 20
date: 2023-09-21
aliases:
    /docs/tasks/cluster/cluster-license/
description: >
  Apply EKS Anywhere Enterprise Subscription licenses to EKS Anywhere clusters
---

When you purchase an EKS Anywhere Enterprise Subscription, licenses are created in the AWS Region and account you used to purchase the subscription. After purchasing your subscription, you can view your licenses, accept the license grants, and apply the license IDs or license tokens to your EKS Anywhere clusters. 

## Get license ID string or license token

The two key parts of the license are the license ID string and the license token. In EKS Anywhere versions `v0.21.x` and below, the license ID string is applied as a Kubernetes Secret to EKS Anywhere clusters and is used for AWS Support cases to validate the cluster is eligible for support. The license token was introduced in EKS Anywhere version `v0.22.0` and all existing EKS Anywhere subscriptions have been updated with a license token for each license. The license token is applied in the EKS Anywhere cluster specification.

You can use either the license ID string or the license token when you create AWS Support cases for your EKS Anywhere clusters. To use extended support for Kubernetes versions in EKS Anywhere, available for EKS Anywhere versions `v0.22.0` and above, your clusters must have a valid and unexpired license token to be able to create and upgrade clusters using the Kubernetes extended support versions.

#### AWS Management Console

You can view the licenses for your subscription in the [EKS Anywhere section of the EKS console](https://console.aws.amazon.com/eks/home#/eks-anywhere) by clicking on the Name of your active subscription. The licenses panel is shown on the Subscription details page and contains the license ID string and the license token for each license associated with your subscription.

If you are applying a license to an EKS Anywhere cluster using EKS Anywhere version `v0.22.0` or above, copy the **license token** and proceed to [Apply license to EKS Anywhere cluster.]({{< relref "cluster-license#apply-license-to-eks-anywhere-cluster">}})

If you are applying a license to an EKS Anywhere cluster using EKS Anywhere version `v0.21.x` or below, copy the **license ID string** and proceed to [Apply license to EKS Anywhere cluster.]({{< relref "cluster-license#apply-license-to-eks-anywhere-cluster">}})

#### AWS CLI

Use the following command to get the license ID strings and license tokens for each license associated with a subscription. Note, the command must be run with the same account that created the subscription.

- Replace `region-code` with the AWS Region that hosts your subscription (for example `us-west-2`).
- Replace `my-subscription` in the `--query` string with the `name` for your subscription.

```bash
aws eks list-eks-anywhere-subscriptions \
  --region 'region-code' \
  --query 'subscriptions[?name==`my-subscription`].licenses[*]'
```

If you are applying a license to an EKS Anywhere cluster using EKS Anywhere version `v0.22.0` or above, copy the **license token** and proceed to [Apply license to EKS Anywhere cluster.]({{< relref "cluster-license#apply-license-to-eks-anywhere-cluster">}}) An example of the license token in the response is shown below in the `token` field.

```
[
    [
        {
            "id": "l-58dc5e15eb12396b86e5724db1a710d9",
            "token": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJsaWNlbnNlSWQiOiJsLTU4ZGM1ABC1ZWI5ODQ5NmI4NmU1NzI0ZGIxYTcxMGQ5IiwibGljZW5zZVZlcnNpb24iOiIxIiwiYmVnaW5WYWxpZGl0eSI6IjIwMjUtMDItMDhUMDA6MDY6MzYuMDAwWiIsImVuZFZhbGlkaXR5IjoiMjAyNi0wMi0wOVQwMDowNjozNi4wMDBaIiwic3Vic2NyaXB0aW9uSWQiOiI0YjMwNmM3Mi1kZmRmLTRlMWUtODQ1OS0wMWU2MWVkOGM1NGM6NWY5MjhiZTQiLCJzdWJzY3JpcHRpb25OYW1lIjoibXktdGVzdC1zdWJzY3JpcHRpb24iLCJhY2NvdW50SWQiOiI2NTkzNTYzOTg0MDQiLCJyZWdpb24iOiJ1cy13ZXN0LTIifQ.72Hiz4RqdNMQnObLTI0gCxT7vj1WBMNU8vvD2v0gbGl2Tas5VT30R-7GXCE6x73G613V6o12kqcnQM6DCwzeSg"
        }
    ]
]
```

If you are applying a license to an EKS Anywhere cluster using EKS Anywhere version `v0.21.x` or below, copy the **license ID string** and proceed to [Apply license to EKS Anywhere cluster.]({{< relref "cluster-license#apply-license-to-eks-anywhere-cluster">}}) An example of the license ID string in the response is shown below in the `id` field.

```
[
    [
        {
            "id": "l-58dc5e15eb12396b86e5724db1a710d9",
            "token": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJsaWNlbnNlSWQiOiJsLTU4ZGM1ABC1ZWI5ODQ5NmI4NmU1NzI0ZGIxYTcxMGQ5IiwibGljZW5zZVZlcnNpb24iOiIxIiwiYmVnaW5WYWxpZGl0eSI6IjIwMjUtMDItMDhUMDA6MDY6MzYuMDAwWiIsImVuZFZhbGlkaXR5IjoiMjAyNi0wMi0wOVQwMDowNjozNi4wMDBaIiwic3Vic2NyaXB0aW9uSWQiOiI0YjMwNmM3Mi1kZmRmLTRlMWUtODQ1OS0wMWU2MWVkOGM1NGM6NWY5MjhiZTQiLCJzdWJzY3JpcHRpb25OYW1lIjoibXktdGVzdC1zdWJzY3JpcHRpb24iLCJhY2NvdW50SWQiOiI2NTkzNTYzOTg0MDQiLCJyZWdpb24iOiJ1cy13ZXN0LTIifQ.72Hiz4RqdNMQnObLTI0gCxT7vj1WBMNU8vvD2v0gbGl2Tas5VT30R-7GXCE6x73G613V6o12kqcnQM6DCwzeSg"
        }
    ]
]
```

## Apply license to EKS Anywhere cluster

A license can only be bound to one EKS Anywhere cluster at a time, and you can only receive support for your EKS Anywhere cluster if it has a valid and active license. You can only create or update EKS Anywhere clusters with extended support for Kubernetes versions if there is a valid and active license token available for the cluster. Extended support for Kubernetes versions is available in EKS Anywhere versions `v0.22.0` and above.

#### Apply license to EKS Anywhere cluster with version `v0.22.0` or above

You can apply a license token to an EKS Anywhere cluster during or after cluster creation for standalone, management, and workload clusters. License tokens are configured in the EKS Anywhere cluster specification in the `spec.licenseToken` field. An example of a license token configuration in the EKS Anywhere cluster specification is shown below.

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
spec:
  kubernetesVersion: "1.28"
  licenseToken: "eyJsaWNlbnNlSWQiOiJsLTU4ZGM1ABC1ZWI5ODQ5NmI4NmU1NzI0ZGIxYTcxMGQ5IiwibGljZW5zZVZlcnNpb24iOiIxIiwiYmVnaW5WYWxpZGl0eSI6IjIwMjUtMDItMDhUMDA6MDY6MzYuMDAwWiIsImVuZFZhbGlkaXR5IjoiMjAyNi0wMi0wOVQwMDowNjozNi4wMDBaIiwic3Vic2NyaXB0aW9uSWQiOiI0YjMwNmM3Mi1kZmRmLTRlMWUtODQ1OS0wMWU2MWVkOGM1NGM6NWY5MjhiZTQiLCJzdWJzY3JpcHRpb25OYW1lIjoibXktdGVzdC1zdWJzY3JpcHRpb24iLCJhY2NvdW50SWQiOiI2NTkzNTYzOTg0MDQiLCJyZWdpb24iOiJ1cy13ZXN0LTIifQ.72Hiz4RqdNMQnObLTI0gCxT7vj1WBMNU8vvD2v0gbGl2Tas5VT30R-7GXCE6x73G613V6o12kqcnQM6DCwzeSg"
  ...
```
To apply the license token to your cluster, run the `eksctl anywhere create` or `eksctl anywhere upgrade` command, or use Kubernetes API-compatible tooling for workload clusters.

**eksctl anywhere CLI**

New cluster
```
eksctl anywhere create cluster -f my-cluster.yaml --kubeconfig my-cluster.kubeconfig
```
Existing cluster
```
eksctl anywhere upgrade cluster -f my-cluster.yaml --kubeconfig my-cluster.kubeconfig
```
**Kubernetes API-compatible tooling**
```
kubectl apply -f my-cluster.yaml --kubeconfig my-cluster.kubeconfig
```

#### Apply license to EKS Anywhere cluster with version `v0.21.x` or below

You can apply a license ID string to an EKS Anywhere cluster during or after cluster creation for standalone or management clusters. For workload clusters, you must apply the license after cluster creation. In the examples below, the `<license-id-string>` is the license ID string, for example `l-58dc5e15eb12396b86e5724db1a710d9`.

To apply a license during standalone or management cluster creation, export the `EKSA_LICENSE` environment variable before running the `eksctl anywhere create cluster` command.

```bash
export EKSA_LICENSE='<license-id-string>'
```

To apply a license to an existing cluster, apply the following Secret to your cluster, replacing `<license-id-string>` with your license ID string.

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

## AWS CLI commands to view license details

**Get license details for all licenses with the AWS CLI**

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