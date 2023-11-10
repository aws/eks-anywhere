---
title: "Managing package bundles"
linkTitle: "Manage package bundles"
weight: 5
aliases:
    /docs/tasks/packages/packagebundles/
---

### Getting new package bundles
Package bundle resources are created and managed in the management cluster, so first set up the `KUBECONFIG` environment variable for the management cluster.
```
export KUBECONFIG=mgmt/mgmt-eks-a-cluster.kubeconfig
```

The EKS Anywhere package controller periodically checks upstream for the latest package bundle and applies it to your management cluster, except for when in an [airgapped environment](https://anywhere.eks.amazonaws.com/docs/getting-started/airgapped/). In that case, you would have to get the package bundle manually from outside of the airgapped environment and apply it to your management cluster.

To view the available `packagebundles` in your cluster, run the following:

```
kubectl get packagebundles -n eksa-packages
NAMESPACE       NAME        STATE
eksa-packages   v1-27-125   available
```

To get a package bundle manually, you can use `oras` to pull the package bundle (bundle.yaml) from the `public.ecr.aws/eks-anywhere` repository. (See the [ORAS CLI official documentation](https://oras.land/docs/) for more details)

```
oras pull public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles:v1-27-latest
Downloading 1ba8253d19f9 bundle.yaml
Downloaded  1ba8253d19f9 bundle.yaml
Pulled [registry] public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles:v1-27-latest
```

Use `kubectl` to apply the new package bundle to your cluster to make it available for use.
```
kubectl apply -f bundle.yaml
```

The package bundle should now be available for use in the management cluster.

```
kubectl get packagebundles -n eksa-packages
NAMESPACE       NAME        STATE
eksa-packages   v1-27-125   available
eksa-packages   v1-27-126   available
```

### Activating a package bundle

There are multiple `packagebundlecontrollers` resources in the management cluster which allows for each cluster to activate different package bundle versions. The active package bundle determines the versions of the packages that are installed on that cluster.

To view which package bundle is active for each cluster, use the `kubectl` command to list the `packagebundlecontrollers` objects in the management cluster.
```
kubectl get packagebundlecontrollers -A
NAMESPACE       NAME   ACTIVEBUNDLE   STATE    DETAIL
eksa-packages   mgmt   v1-27-125     active   
eksa-packages   w01    v1-27-125     active 
```

Use the EKS Anywhere packages CLI to upgrade the active package bundle of the target cluster. This command can also be used to downgrade to a previous package bundle version.
```
export CLUSTER_NAME=mgmt
eksctl anywhere upgrade packages --bundle-version v1-27-126 --cluster $CLUSTER_NAME
```


