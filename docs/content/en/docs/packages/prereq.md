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

### Identify AWS account ID for ECR packages registry

The AWS account ID for ECR packages registry depends on the EKS Anywhere Enterprise Subscription.

* For EKS Anywhere Enterprise Subscriptions purchased through the AWS console or APIs the AWS account ID for ECR packages registry varies depending on the region the Enterprise Subscription was purchased. Reference the table in the expanded output below for a mapping of AWS Regions to ECR package registries.
<details>
  <summary>Expand for packages registry to AWS Region table</summary>
  <br /> 
  {{% content "../clustermgmt/support/packages-registries.md" %}}
</details>
  <br/>

* For EKS Anywhere Curated Packages trials or EKS Anywhere Enterprise Subscriptions purchased before October 2023 the AWS account ID for ECR packages registry is `783794618700`. This supports pulling images from the following regions.
<details>
  <summary>Expand for AWS Regions table</summary>
  <br />
  {{% content "./legacypackagesregions.md" %}}
</details>
<br/>

After identifying the AWS account ID; export it for further reference. Example
```bash
export ECR_PACKAGES_ACCOUNT=346438352937
```

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
            "Resource": "arn:aws:ecr:*:<ECR_PACKAGES_ACCOUNT>:repository/*"
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

**Note** Use the corresponding `EKSA_AWS_REGION` prior to cluster creation to choose which region to pull form.

Create credentials for this user and set and export the following environment variables:
```bash
export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
export EKSA_AWS_REGION="aws*region"
```
Make sure you are authenticated with the AWS CLI

```bash
export AWS_ACCESS_KEY_ID="your*access*id"
export AWS_SECRET_ACCESS_KEY="your*secret*key"
aws sts get-caller-identity
```

Login to docker

```bash
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $ECR_PACKAGES_ACCOUNT.dkr.ecr.$EKSA_AWS_REGION.amazonaws.com
```

Verify you can pull an image
```bash
docker pull $ECR_PACKAGES_ACCOUNT.dkr.ecr.$EKSA_AWS_REGION.amazonaws.com/emissary-ingress/emissary:v3.9.1-828e7d186ded23e54f6bd95a5ce1319150f7e325
```
If the image downloads successfully, it worked!

### Prepare for using curated packages for airgapped environments

If you are running in an airgapped environment and you set up a local registry mirror, you can copy curated packages from Amazon ECR to your local registry mirror with the following command. 

The `$KUBEVERSION` should be set to the same as the Kubernetes version in `spec.kubernetesVersion` of your EKS Anywhere cluster specification. When using self-signed certificates for your registry, you should run with the `--dst-insecure` command line argument to indicate skipping TLS verification while copying curated packages. 

```bash
eksctl anywhere copy packages \
  ${REGISTRY_MIRROR_URL}/curated-packages \
  --kube-version $KUBEVERSION \
  --src-chart-registry public.ecr.aws/eks-anywhere \
  --src-image-registry ${ECR_PACKAGES_ACCOUNT}.dkr.ecr.${EKSA_AWS_REGION}.amazonaws.com
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
eksctl anywhere list packages --kube-version 1.33
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
eksctl anywhere generate package harbor --cluster ${CLUSTER_NAME} --kube-version 1.33 > harbor-spec.yaml
```
