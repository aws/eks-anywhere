---
title: "Managing the package controller"
linkTitle: "Manage package controller"
weight: 4
aliases:
    /docs/tasks/packages/packagecontroller/
---

### Installing the package controller

{{% alert title="Important" color="warning" %}}
The package controller installation creates a package bundle controller resource for each cluster, thus allowing each to activate a different package bundle version. Ideally, you should never delete this resource because it would mean losing that information and upon re-installing, the latest bundle would be selected. However, you can always go back to the previous bundle version. For more information, see [Managing package bundles.]({{< relref "./packagebundles" >}})
{{% /alert %}}

The package controller is typically installed during cluster creation, but may be disabled intentionally in your `cluster.yaml` by setting `spec.packages.disable` to `true`.

If you created a cluster without the package controller or if the package controller was not properly configured, you may need to manually install it.

1. Enable the package controller in your `cluster.yaml`, if it was previously disabled:
    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: mgmt
    spec:
      packages:
        disable: false
    ```

2. Make sure you are authenticated with the AWS CLI. Use the credentials you set up for packages. These credentials should have [limited capabilities]({{< relref "./prereq#setup-authentication-to-use-curated-packages" >}}):

    ```bash
    export AWS_ACCESS_KEY_ID="your*access*id"
    export AWS_SECRET_ACCESS_KEY="your*secret*key"
    export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
    export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
    ```

3. Verify your credentials are working:
    ```shell
    aws sts get-caller-identity
    ```

4. Authenticate docker to the private AWS ECR registry on account `783794618700` with your AWS credentials. It houses the EKS Anywhere packages artifacts. Authentication is required to pull images from it.
    ```bash
    aws ecr get-login-password | docker login --username AWS --password-stdin 783794618700.dkr.ecr.us-west-2.amazonaws.com
    ```

5. Verify you can pull an image from the packages registry:
    ```bash
    docker pull 783794618700.dkr.ecr.us-west-2.amazonaws.com/emissary-ingress/emissary:v3.5.1-bf70150bcdfe3a5383ec8ad9cd7eea801a0cb074
    ```
    If the image downloads successfully, it worked!

6. Now, install the package controller using the EKS Anywhere Packages CLI:
    ```shell
    eksctl anywhere install packagecontroller -f cluster.yaml
    ```

    The package controller should now be installed!

7. Use kubectl to check the eks-anywhere-packages pod is running in your management cluster:
    ```
    kubectl get pods -n eksa-packages 
    NAME                                     READY   STATUS    RESTARTS   AGE
    eks-anywhere-packages-55bc54467c-jfhgp   1/1     Running   0          21s
    ```

### Updating the package credentials

You may need to create or update your credentials which you can do with a command like this. Set the environment variables to the proper values before running the command.
  ```bash
  kubectl delete secret -n eksa-packages aws-secret
  kubectl create secret -n eksa-packages generic aws-secret \
    --from-literal=AWS_ACCESS_KEY_ID=${EKSA_AWS_ACCESS_KEY_ID} \
    --from-literal=AWS_SECRET_ACCESS_KEY=${EKSA_AWS_SECRET_ACCESS_KEY}  \
    --from-literal=REGION=${EKSA_AWS_REGION}
  ```

### Upgrade the packages controller

EKS Anywhere v0.15.0 (packages controller v0.3.9+) and onwards includes support for the eks-anywhere-packages controller as a self-managed package feature. The package controller now upgrades automatically according to the version specified within the management cluster's selected package bundle.

For any version prior to v0.3.X, manual steps must be executed to upgrade.

{{% alert title="Important" color="warning" %}}
This operation may change your cluster's selected package bundle to the latest version. However, you can always go back to the previous bundle version. For more information, see [Managing package bundles.]({{< relref "./packagebundles" >}})
{{% /alert %}}

To manually upgrade the package controller, do the following:

1. Ensure the namespace will be kept
```
kubectl annotate namespaces eksa-packages helm.sh/resource-policy=keep
```

2. Uninstall the eks-anywhere-packages helm release
```
helm uninstall -n eksa-packages eks-anywhere-packages
```

3. Remove the secret called aws-secret (we will need credentials when installing the new version)
```
kubectl delete secret -n eksa-packages aws-secret
```

4. Install the new version using the latest eksctl-anywhere binary on your management cluster
```
eksctl anywhere install packagecontroller -f eksa-mgmt-cluster.yaml
```
