---
title: "EKS Anywhere curated package management"
linkTitle: "Package management"
date: 2022-04-12
weight: 40
description: >
  Common tasks for managing curated packages.
---

The main goal of EKS Anywhere curated packages is to make it easy to install, configure and maintain operational components in an EKS Anywhere cluster. EKS Anywhere curated packages offers to run secure and tested operational components on EKS Anywhere clusters. Please check out [EKS Anywhere curated packages concepts]({{< relref "../../concepts/packages" >}}) and [EKS Anywhere curated packages configurations]({{< relref "../../reference/clusterspec/packages.md" >}}) for more details.

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

Create credentials for this user and set and export the following environment variables:
```bash
export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
```
Make sure you are authenticated with the AWS CLI

```bash
export AWS_ACCESS_KEY_ID="your*access*id"
export AWS_SECRET_ACCESS_KEY="your*secret*key"
aws sts get-caller-identity
```

Login to docker

```bash
aws ecr get-login-password |docker login --username AWS --password-stdin 783794618700.dkr.ecr.us-west-2.amazonaws.com
```

Verify you can pull an image
```bash
docker pull 783794618700.dkr.ecr.us-west-2.amazonaws.com/emissary-ingress/emissary:v3.0.0-9ded128b4606165b41aca52271abe7fa44fa7109
```
If the image downloads successfully, it worked!

### Discover curated packages

You can get a list of the available packages from the command line:

```bash
eksctl anywhere list packages --source registry --kube-version 1.23
```

Example command output:
```              
Package                 Version(s)                                       
-------                 ----------                                       
hello-eks-anywhere      0.1.1-a217465b3b2d165634f9c24a863fa67349c7268a   
harbor                  2.5.1-a217465b3b2d165634f9c24a863fa67349c7268a   
metallb                 0.12.1-b9e4e5d941ccd20c72b4fec366ffaddb79bbc578  
emissary                3.0.0-a507e09c2a92c83d65737835f6bac03b9b341467
```

### Generate a curated-packages config

The example shows how to install the `harbor` package from the [curated package list]({{< relref "../../reference/packagespec" >}}).
```bash
eksctl anywhere generate package harbor --kube-version 1.23 > packages.yaml
```

Available curated packages are listed below.
