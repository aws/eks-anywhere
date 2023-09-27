---
title: "Prerequisites"
linkTitle: "Prerequisites"
weight: 2
aliases:
    /docs/tasks/packages/prereq/
description: >
  Prerequisites for using curated packages
---

## Prerequisites
Before installing any curated packages for EKS Anywhere, do the following:

* Check that the cluster `Kubernetes` version is `v1.21` or above. For example, you could run `kubectl get cluster -o yaml <cluster-name> | grep -i kubernetesVersion`
* Check that the version of `eksctl anywhere` is `v0.11.0` or above with the `eksctl anywhere version` command.
* It is recommended that the package controller is only installed on the management cluster.
* Check the existence of package controller:
    ```bash
    kubectl get pods -n eksa-packages | grep "eks-anywhere-packages"
    ```
    If the returned result is empty, you need to install the package controller.

* Install the package controller if it is not installed:
    Install the package controller
     
     *Note* This command is temporarily provided to ease integration with curated packages. This command will be deprecated in the future
 
     ```bash
     eksctl anywhere install packagecontroller -f $CLUSTER_NAME.yaml
     ```

To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/).

### Setup authentication to use curated packages

When you have been notified that your account has been given access to curated packages, create an IAM user in your account with a policy that only allows ECR read access to the Curated Packages repository; similar to this:

```
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
            "Resource": "arn:aws:ecr:*:783794618700:repository/*"
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
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 783794618700.dkr.ecr.us-west-2.amazonaws.com
```

Verify you can pull an image
```bash
docker pull 783794618700.dkr.ecr.us-west-2.amazonaws.com/emissary-ingress/emissary:v3.5.1-bf70150bcdfe3a5383ec8ad9cd7eea801a0cb074
```
If the image downloads successfully, it worked!

### Prepare for using curated packages for airgapped environments

If you are running in an airgapped environment and you set up a local registry mirror, you can copy curated packages from Amazon ECR to your local registry mirror with the following command. 

The `$BUNDLE_RELEASE_YAML_PATH` should be set to the `eks-anywhere-downloads/bundle-release.yaml` location where you unpacked the tarball from the`eksctl anywhere download artifacts` command. The `$REGISTRY_MIRROR_CERT_PATH` and `$REGISTRY_MIRROR_URL` values must be the same as the `registryMirrorConfiguration` in your EKS Anywhere cluster specification.

```bash
eksctl anywhere copy packages \
  --bundle ${BUNDLE_RELEASE_YAML_PATH} \
  --dst-cert ${REGISTRY_MIRROR_CERT_PATH} \
  ${REGISTRY_MIRROR_URL}
```

Once the curated packages images are in your local registry mirror, you must configure the curated packages controller to use your local registry mirror post-cluster creation. Configure the `defaultImageRegistry` and `defaultRegistry` settings for the `PackageBundleController` to point to your local registry mirror by applying a similar `yaml` definition as the one below to your standalone or management cluster. Existing `PackageBundleController` can be changed, and you do not need to deploy a new `PackageBundleController`. See the [Packages configuration documentation]({{< relref "./packages/#packagebundlecontrollerspec" >}}) for more information.

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: PackageBundleController
metadata:
  name: eksa-packages-bundle-controller
  namespace: eksa-packages
spec:
  defaultImageRegistry: ${REGISTRY_MIRROR_URL}/curated-packages
  defaultRegistry: ${REGISTRY_MIRROR_URL}/eks-anywhere
```

### Discover curated packages

You can get a list of the available packages from the command line:

```bash
export CLUSTER_NAME=<your-cluster-name>
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
eksctl anywhere list packages --kube-version 1.27
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
prometheus              2.41.0-b53c8be243a6cc3ac2553de24ab9f726d9b851ca
```

### Generate curated packages configuration

The example shows how to install the `harbor` package from the [curated package list]({{< relref "./packagelist/" >}}).

```bash
export CLUSTER_NAME=<your-cluster-name>
eksctl anywhere generate package harbor --cluster ${CLUSTER_NAME} --kube-version 1.27 > harbor-spec.yaml
```
