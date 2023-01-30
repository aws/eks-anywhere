---
title: "EKS Anywhere curated package management"
linkTitle: "Package management"
date: 2022-04-12
weight: 40
description: >
  Common tasks for managing curated packages.
---

The main goal of EKS Anywhere curated packages is to make it easy to install, configure and maintain operational components in an EKS Anywhere cluster. EKS Anywhere curated packages offers to run secure and tested operational components on EKS Anywhere clusters. Please check out [EKS Anywhere curated packages concepts]({{< relref "../../concepts/packages" >}}) and [EKS Anywhere curated packages configurations]({{< relref "../../reference/packagespec/packages.md" >}}) for more details.

For proper curated package support, make sure the cluster `kubernetes` version is `v1.21` or above and `eksctl anywhere` version is `v0.11.0` or above (can be checked with the `eksctl anywhere version` command). Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/).

### Setup authentication to use curated-packages

When you have been notified that your account has been given access to curated packages, create a user in your account with a policy that only allows ECR read access similar to this:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "ecr:DescribeImageScanFindings",
                "ecr:GetDownloadUrlForLayer",
                "ecr:DescribeRegistry",
                "ecr:DescribePullThroughCacheRules",
                "ecr:DescribeImageReplicationStatus",
                "ecr:GetAuthorizationToken",
                "ecr:ListTagsForResource",
                "ecr:ListImages",
                "ecr:BatchGetImage",
                "ecr:DescribeImages",
                "ecr:DescribeRepositories",
                "ecr:BatchCheckLayerAvailability"
            ],
            "Resource": "*"
        }
    ]
}
```

**Note** Curated Packages now supports pulling images from the following regions. Use the corresponding `EKSA_AWS_REGION` prior to cluster creation to choose which region to pull form, if not set it will default to pull from `us-west-2`.
```
"us-east-2",
"us-east-1",
"us-west-1",
"us-west-2",
"ap-northeast-3",
"ap-northeast-2",
"ap-southeast-1",
"ap-southeast-2",
"ap-northeast-1",
"ca-central-1",
"eu-central-1",
"eu-west-1",
"eu-west-2",
"eu-west-3",
"eu-north-1",
"sa-east-1"
```


Create credentials for this user and set and export the following environment variables:
```bash
export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
export EKSA_AWS_REGION="us-west-2"
```
Make sure you are authenticated with the AWS CLI

```bash
export AWS_ACCESS_KEY_ID="your*access*id"
export AWS_SECRET_ACCESS_KEY="your*secret*key"
aws sts get-caller-identity
```

Login to docker

```bash
aws ecr get-login-password --region us-west-2 |docker login --username AWS --password-stdin 783794618700.dkr.ecr.us-west-2.amazonaws.com
```

Verify you can pull an image
```bash
docker pull 783794618700.dkr.ecr.us-west-2.amazonaws.com/emissary-ingress/emissary:v3.0.0-9ded128b4606165b41aca52271abe7fa44fa7109
```
If the image downloads successfully, it worked!

### Discover curated packages

You can get a list of the available packages from the command line:

```bash
export CLUSTER_NAME=nameofyourcluster
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
eksctl anywhere list packages --kube-version 1.23
```

Example command output:
```
Package                 Version(s)
-------                 ----------
hello-eks-anywhere      0.1.2-a6847010915747a9fc8a412b233a2b1ee608ae76
adot                    0.25.0-c26690f90d38811dbb0e3dad5aea77d1efa52c7b
cert-manager            1.9.1-dc0c845b5f71bea6869efccd3ca3f2dd11b5c95f
cluster-autoscaler      9.21.0-1.23-5516c0368ff74d14c328d61fe374da9787ecf437
harbor                  2.5.1-ee7e5a6898b6c35668a1c5789aa0d654fad6c913
metallb                 0.13.7-758df43f8c5a3c2ac693365d06e7b0feba87efd5
metallb-crds            0.13.7-758df43f8c5a3c2ac693365d06e7b0feba87efd5
metrics-server          0.6.1-eks-1-23-6-c94ed410f56421659f554f13b4af7a877da72bc1
emissary                3.3.0-cbf71de34d8bb5a72083f497d599da63e8b3837b
emissary-crds           3.3.0-cbf71de34d8bb5a72083f497d599da63e8b3837b
prometheus              2.39.1-a5a5cc31867fdec4da73f95483e6893cf424f80f
```

### Generate a curated-packages config

The example shows how to install the `harbor` package from the [curated package list]({{< relref "../../reference/packagespec" >}}).
```bash
export CLUSTER_NAME=nameofyourcluster
eksctl anywhere generate package harbor --cluster ${CLUSTER_NAME} --kube-version 1.23 > packages.yaml
```

Available curated packages and troubleshooting guides are listed below.
