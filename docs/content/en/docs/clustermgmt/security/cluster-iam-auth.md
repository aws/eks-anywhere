---
title: "Authenticate cluster with AWS IAM Authenticator"
linkTitle: "AWS IAM Authenticator"
weight: 30
aliases:
    /docs/tasks/cluster/cluster-iam-auth/
date: 2022-01-05
description: >
  Configure AWS IAM Authenticator to authenticate user access to the cluster
---

## AWS IAM Authenticator Support (optional)

EKS Anywhere supports configuring [AWS IAM Authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) as an authentication provider for clusters.

When you create a cluster with IAM Authenticator enabled, EKS Anywhere 
* Installs `aws-iam-authenticator` server as a DaemonSet on the workload cluster.
* Configures the Kubernetes API Server to communicate with iam authenticator using a [token authentication webhook](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication).
* Creates the necessary ConfigMaps based on user options.

{{% alert title="Note" color="primary" %}}
Enabling IAM Authenticator needs to be done during cluster creation.
{{% /alert %}}

### Create IAM Authenticator enabled cluster
Generate your cluster configuration and add the necessary IAM Authenticator configuration. For a full spec reference check [AWSIamConfig]({{< relref "../../getting-started/optional/iamauth" >}}).

Create an EKS Anywhere cluster as follows:

```bash
CLUSTER_NAME=my-cluster-name
eksctl anywhere create cluster -f ${CLUSTER_NAME}.yaml
```

#### Example AWSIamConfig configuration
This example uses a region in the default aws partition and `EKSConfigMap` as `backendMode`. Also, the IAM ARNs are mapped to the kubernetes `system:masters` group.
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
   # IAM Authenticator
   identityProviderRefs:
      - kind: AWSIamConfig
        name: aws-iam-auth-config
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: AWSIamConfig
metadata:
   name: aws-iam-auth-config
spec:
    awsRegion: us-west-1
    backendMode:
        - EKSConfigMap
    mapRoles:
        - roleARN: arn:aws:iam::XXXXXXXXXXXX:role/myRole
          username: myKubernetesUsername
          groups:
          - system:masters
    mapUsers:
        - userARN: arn:aws:iam::XXXXXXXXXXXX:user/myUser
          username: myKubernetesUsername
          groups:
          - system:masters
    partition: aws
```

{{% alert title="Note" color="primary" %}}
When using backend mode `CRD`, the `mapRoles` and `mapUsers` are not required. For more details on configuring CRD mode, refer to [CRD.](https://github.com/kubernetes-sigs/aws-iam-authenticator#crd-alpha)
{{% /alert %}}

### Authenticating with IAM Authenticator
After your cluster is created you may now use the mapped IAM ARNs to authenticate to the cluster. 

EKS Anywhere generates a `KUBECONFIG` file in your local directory that uses `aws-iam-authenticator client` to authenticate with the cluster. The file can be found at
```bash
${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-aws.kubeconfig
```
#### Steps
1. Ensure the IAM role/user ARN mapped in the cluster is configured on the local machine from which you are trying to access the cluster.
2. Install the `aws-iam-authenticator client` binary on the local machine. 
    * We recommend installing the binary referenced in the latest `release manifest` of the kubernetes version used when creating the cluster.
    * The below commands can be used to fetch the installation uri for clusters created with `1.27` kubernetes version and OS `linux`.
    ```bash
    CLUSTER_NAME=my-cluster-name
    KUBERNETES_VERSION=1.27

    export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig

    EKS_D_MANIFEST_URL=$(kubectl get bundles $CLUSTER_NAME -o jsonpath="{.spec.versionsBundles[?(@.kubeVersion==\"$KUBERNETES_VERSION\")].eksD.manifestUrl}")
    
    OS=linux
    curl -fsSL $EKS_D_MANIFEST_URL | yq e '.status.components[] | select(.name=="aws-iam-authenticator") | .assets[] | select(.os == '"\"$OS\""' and .type == "Archive") | .archive.uri' -
    ```

3. Export the generated IAM Authenticator based `KUBECONFIG` file.
    ```bash
    export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-aws.kubeconfig
    ```
4. Run `kubectl` commands to check cluster access. Example,
    ```bash
    kubectl get pods -A
    ```

### Modify IAM Authenticator mappings
EKS Anywhere supports modifying IAM ARNs that are mapped on the cluster. The mappings can be modified by either running the `upgrade cluster` command or using `GitOps`.

#### upgrade command
The `mapRoles` and `mapUsers` lists in `AWSIamConfig` can be modified when running the `upgrade cluster` command from EKS Anywhere.

As an example, let's add another IAM user to the above example configuration.
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: AWSIamConfig
metadata:
   name: aws-iam-auth-config
spec:
    ...
    mapUsers:
        - userARN: arn:aws:iam::XXXXXXXXXXXX:user/myUser
          username: myKubernetesUsername
          groups:
          - system:masters
        - userARN: arn:aws:iam::XXXXXXXXXXXX:user/anotherUser
          username: anotherKubernetesUsername
    partition: aws
```
and then run the upgrade command
```bash
CLUSTER_NAME=my-cluster-name
eksctl anywhere upgrade cluster -f ${CLUSTER_NAME}.yaml
```
EKS Anywhere now updates the role mappings for IAM authenticator in the cluster and a new user gains access to the cluster.

#### GitOps
If the cluster created has GitOps configured, then the `mapRoles` and `mapUsers` list in `AWSIamConfig` can be modified by the GitOps controller. For GitOps configuration details refer to [Manage Cluster with GitOps]({{< relref "../../clustermgmt/cluster-flux" >}}).

{{% alert title="Note" color="primary" %}}
GitOps support for the `AWSIamConfig` is currently only on management or self-managed clusters.
{{% /alert %}}

1. Clone your git repo and modify the cluster specification.
   The default path for the cluster file is:
    ```
    clusters/$CLUSTER_NAME/eksa-system/eksa-cluster.yaml
    ```
2. Modify the `AWSIamConfig` object and add to the `mapRoles` and `mapUsers` object lists.
3. Commit the file to your git repository
    ```bash
    git add eksa-cluster.yaml
    git commit -m 'Adding IAM Authenticator access ARNs'
    git push origin main
    ```
EKS Anywhere GitOps Controller now updates the role mappings for IAM authenticator in the cluster and users gains access to the cluster.
