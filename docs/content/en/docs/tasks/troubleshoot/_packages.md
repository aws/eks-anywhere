
---
title: "Curating Packages Troubleshooting"
linkTitle: "Curating Packages Troubleshooting"
weight: 50
---


The major component of Curated Packages is the package controller. If the container is not running or not running correctly, packages will not be installed. Generally it should be debugged like any other Kubernetes application. The first step is to check that then pod is running.
```bash
kubectl get pods -n eksa-packages
```

You should see one pod running with two containers
```bash
NAME                                    READY   STATUS             RESTARTS   AGE
eks-anywhere-packages-fff64dc58-795mt   1/2     ImagePullBackOff   0          45s
```

The describe command might help to get more detail on why there is a problem
```bash
kubectl describe pods -n eksa-packages
```

Logs of the controller can be seen in a normal Kubernetes fashion
```bash
kubectl logs deploy/eks-anywhere-packages -n eksa-packages controller
```

The general state of the package can be seen through the custom resources
```bash
kubectl get packages,packagebundles,packagebundlecontrollers -A
```

This will generate output similar to this
```bash
NAMESPACE       NAME                                         PACKAGE   AGE     STATE        CURRENTVERSION   TARGETVERSION                                                   DETAIL
eksa-packages   package.packages.eks.amazonaws.com/my-test   Test      2m33s   installing                    v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm (latest)   loading helm chart my-test: locating helm chart oci://public.ecr.aws/l0g8r8j6/hello-eks-anywhere tag sha256:64ea03b119d2421f9206252ff4af4bf7cdc2823c343420763e0e6fc20bf03b68: failed to download "oci://public.ecr.aws/l0g8r8j6/hello-eks-anywhere" at version "v0.1.1-8b3810e1514b7432e032794842425accc837757a-helm"

NAMESPACE       NAME                                                   STATE
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-21-1001    active

NAMESPACE       NAME                                                                                 STATE
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/eksa-packages-bundle-controller   active
```

Looking at the output, you can see the active packagebundlecontroller and packagebundle. The state of the package is "installing".