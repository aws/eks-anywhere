---
title: "IAM Authentication"
linkTitle: "IAM Authentication"
weight: 25
aliases:
    /docs/reference/clusterspec/optional/iamauth/
description: >
  EKS Anywhere cluster yaml specification AWS IAM Authenticator reference
---

## AWS IAM Authenticator support (optional)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

EKS Anywhere can create clusters that support AWS IAM Authenticator-based api server authentication.
In order to add IAM Authenticator support, you need to configure your cluster by updating the configuration file before creating the cluster.
This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
   # IAM Authenticator support
   identityProviderRefs:
      - kind: AWSIamConfig
        name: aws-iam-auth-config
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: AWSIamConfig
metadata:
   name: aws-iam-auth-config
spec:
    awsRegion: ""
    backendMode:
        - ""
    mapRoles:
        - roleARN: arn:aws:iam::XXXXXXXXXXXX:role/myRole
          username: myKubernetesUsername
          groups:
          - ""
    mapUsers:
        - userARN: arn:aws:iam::XXXXXXXXXXXX:user/myUser
          username: myKubernetesUsername
          groups:
          - ""
    partition: ""
```
### __identityProviderRefs__ (Under Cluster)
List of identity providers you want configured for the Cluster.
This would include a reference to the `AWSIamConfig` object with the configuration below.

### __awsRegion__ (required)
* __Description__: awsRegion can be any region in the aws partition that the IAM roles exist in.
* __Type__: string

### __backendMode__ (required)
* __Description__: backendMode configures the IAM authenticator server’s backend mode (i.e. where to source mappings from). We support [EKSConfigMap](https://github.com/kubernetes-sigs/aws-iam-authenticator#eksconfigmap) and [CRD](https://github.com/kubernetes-sigs/aws-iam-authenticator#crd-alpha) modes supported by AWS IAM Authenticator, for more details refer to [backendMode](https://github.com/kubernetes-sigs/aws-iam-authenticator#4-create-iam-roleuser-to-kubernetes-usergroup-mappings)
* __Type__: string

### __mapRoles__,  __mapUsers__ (recommended for `EKSConfigMap` backend)
* __Description__: When using `EKSConfigMap` `backendMode`, we recommend providing either `mapRoles` or `mapUsers` to set the IAM role mappings at the time of creation. This input is added to an EKS style ConfigMap. For more details refer to [EKS IAM](https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html)
* __Type__: list object

  #### __roleARN__, __userARN__ (required)
  * __Description__: IAM ARN to authenticate to the cluster. `roleARN` specifies an IAM role and `userARN` specifies an IAM user.
  * __Type__: string

  #### __username__ (required)
  * __Description__: The Kubernetes username the IAM ARN is mapped to in the cluster. The ARN gets mapped to the Kubernetes cluster permissions associated with the username.
  * __Type__: string

  #### __groups__
  * __Description__: List of kubernetes user groups that the mapped IAM ARN is given permissions to.
  * __Type__: list string

### __partition__
* __Description__: This field is used to set the aws partition that the IAM roles are present in. Default value is `aws`.
* __Type__: string
