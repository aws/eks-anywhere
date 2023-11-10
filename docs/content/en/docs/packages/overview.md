---
title: "Overview of curated packages"
linkTitle: Overview
aliases:
    /docs/concepts/packages/artifacts/
    /docs/concepts/packages/cli/
    /docs/concepts/packages/package-controller/
    /docs/reference/packagespec/packages/
weight: 1
---

Components of EKS Anywhere curated packages consist of a controller, a CLI, and artifacts.

### Package controller
The package controller is responsible for installing, upgrading, configuring, and removing packages from the cluster. It performs these actions by watching the package and packagebundle custom resources. Moreover, it uses the packagebundle to determine which packages to run and sets appropriate configuration values.

Packages custom resources map to helm charts that the package controller uses to install packages workloads (such as cluster-autoscaler or metrics-server) on your clusters. The packagebundle object is the mapping between the package name and the specific helm chart and images that will be installed.

The package controller only runs on the management cluster, including single-node clusters, to perform the above outlined responsibilities. However, packages may be installed on both management and workload clusters. For more information, see the guide on [installing packages on workload clusters.]({{< relref "./overview/#installing-packages-on-workload-clusters" >}})

Package release information is stored in a package bundle manifest. The package controller will continually monitor and download new package bundles. When a new package bundle is downloaded, it will show up as "available" in the PackageBundleController resource's `status.detail` field. A package bundle upgrade always requires manual intervention as outlined in the [package bundles]({{< relref "./packagebundles/" >}}) docs.

Any changes to a package custom resource will trigger an install, upgrade, configuration or removal of that package. The package controller will use ECR or private registry to get all resources including bundle, helm charts, and container images.

You may [install the package controller]({{< relref "../packages/packagecontroller" >}}) after cluster creation to take advantage of curated package features.

### Packages CLI
The Curated Packages CLI provides the user experience required to manage curated packages.
Through the CLI, a user is able to discover, create, delete, and upgrade curated packages to a cluster.
These functionalities can be achieved during and after an EKS Anywhere cluster is created.

The CLI provides both imperative and declarative mechanisms to manage curated packages. These 
packages will be included as part of a `packagebundle` that will be provided by the EKS Anywhere team.
Whenever a user requests a package creation through the CLI (`eksctl anywhere create package`), a custom resource is created on the cluster
indicating the existence of a new package that needs to be installed. When a user executes a delete operation (`eksctl anywhere delete package`),
the custom resource will be removed from the cluster indicating the need for uninstalling a package. 
An upgrade through the CLI (`eksctl anywhere upgrade packages`) upgrades all packages to the latest release.

Please check out [Install EKS Anywhere]({{< relref "../getting-started/install" >}}) to install the `eksctl anywhere` CLI on your machine.

The create cluster page for each [EKS Anywhere provider]({{< relref "../getting-started/chooseprovider/" >}}) describes how to configure and install curated packages at cluster creation time.

### Curated packages artifacts
There are three types of build artifacts for packages: the container images, the helm charts and the package bundle manifests. The container images, helm charts and bundle manifests for all of the packages will be built and stored in EKS Anywhere ECR repository. Each package may have multiple versions specified in the packages bundle. The bundle will reference the helm chart tag in the ECR repository. The helm chart will reference the container images for the package.


### Installing packages on workload clusters

![Installing packages on workload clusters](/images/packages-controller-workload-cluster.svg)

The package controller only runs on the management cluster. It determines which cluster to install your package on based on the namespace specified in the `Package` resource.

See an example package below:
```
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: my-hello-eks-anywhere
  namespace: eksa-packages-wk0
spec:
  config: |
        title: "My Hello"
  packageName: hello-eks-anywhere
  targetNamespace: default
```

By specifying `metadata.namespace: eksa-packages-wk0`, the package controller will install the resource on workload cluster wk0.
The pattern for these namespaces is always `eksa-packages-<cluster-name>`.

By specifying `spec.targetNamespace: default`, the package controller will install the hello-eks-anywhere package helm chart in the `default` namespace in cluster wk0.
