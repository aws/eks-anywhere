---
title: "Curated packages troubleshooting"
linkTitle: "Packages troubleshooting"
weight: 7
description: >
  Troubleshooting specific to curated packages
aliases:
   - /docs/tasks/troubleshoot/_packages
   - /docs/tasks/packages/troubleshoot/
---

## General debugging
The major component of Curated Packages is the package controller. If the container is not running or not running correctly, packages will not be installed. Generally it should be debugged like any other Kubernetes application. The first step is to check that the pod is running.
```bash
kubectl get pods -n eksa-packages
```

You should see at least two pods with running and one or more refresher completed.
```bash
NAME                                     READY   STATUS      RESTARTS   AGE
eks-anywhere-packages-69d7bb9dd9-9d47l   1/1     Running     0          14s
eksa-auth-refresher-w82nm                0/1     Completed   0          10s
```

The describe command might help to get more detail on why there is a problem:
```bash
kubectl describe pods -n eksa-packages
```

Logs of the controller can be seen in a normal Kubernetes fashion:
```bash
kubectl logs deploy/eks-anywhere-packages -n eksa-packages controller
```

To get the general state of the package controller, run the following command:
```bash
kubectl get packages,packagebundles,packagebundlecontrollers -A
```

You should see an active packagebundlecontroller and an available bundle. The packagebundlecontroller should indicate the active bundle. It may take a few minutes to download and activate the latest bundle. The state of the package in this example is installing and there is an error downloading the chart.
```bash
NAMESPACE              NAME                                          PACKAGE              AGE   STATE       CURRENTVERSION                                   TARGETVERSION                                             DETAIL
eksa-packages-sammy    package.packages.eks.amazonaws.com/my-hello   hello-eks-anywhere   42h   installed   0.1.1-bc7dc6bb874632972cd92a2bca429a846f7aa785   0.1.1-bc7dc6bb874632972cd92a2bca429a846f7aa785 (latest)   
eksa-packages-tlhowe   package.packages.eks.amazonaws.com/my-hello   hello-eks-anywhere   44h   installed   0.1.1-083e68edbbc62ca0228a5669e89e4d3da99ff73b   0.1.1-083e68edbbc62ca0228a5669e89e4d3da99ff73b (latest)   

NAMESPACE       NAME                                                 STATE
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-21-83    available
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-23-70    available
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-23-81    available
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-23-82    available
eksa-packages   packagebundle.packages.eks.amazonaws.com/v1-23-83    available

NAMESPACE       NAME                                                        ACTIVEBUNDLE   STATE               DETAIL
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/sammy    v1-23-70       upgrade available   v1-23-83 available
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/tlhowe   v1-21-83       active       active   
```

### Package controller not running
If you do not see a pod or various resources for the package controller, it may be that it is not installed.

```
No resources found in eksa-packages namespace.
```

Most likely the cluster was created with an older version of the EKS Anywhere CLI. Curated packages became generally available with `v0.11.0`. Use the `eksctl anywhere version` command to verify you are running a new enough release and you can use the `eksctl anywhere install packagecontroller` command to install the package controller on an older release.

### Error: this command is currently not supported

```
Error: this command is currently not supported
```

Curated packages became generally available with version `v0.11.0`. Use the version command to make sure you are running version `v0.11.0` or later:

```bash
eksctl anywhere version
```

### Error: cert-manager is not present in the cluster
```
Error: curated packages cannot be installed as cert-manager is not present in the cluster
```
This is most likely caused by an action to install curated packages at a workload cluster with `eksctl anywhere` version older than `v0.12.0`. In order to use packages on workload clusters, please upgrade `eksctl anywhere` version to `v0.12+`. The package manager will remotely manage packages on the workload cluster from the management cluster.

## Package registry authentication

### Error: ImagePullBackOff on Package

If a package fails to start with ImagePullBackOff:
```
NAME                                     READY   STATUS             RESTARTS   AGE
generated-harbor-jobservice-564d6fdc87   0/1     ImagePullBackOff   0          2d23h
```

If a package pod cannot pull images, you may not have your AWS credentials set up properly. Verify that your credentials are working properly.

Make sure you are authenticated with the AWS CLI. Use the credentials you set up for packages. These credentials should have [limited capabilities]({{< relref "../packages/prereq#setup-authentication-to-use-curated-packages" >}}):

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
docker pull 783794618700.dkr.ecr.us-west-2.amazonaws.com/emissary-ingress/emissary:v3.5.1-bf70150bcdfe3a5383ec8ad9cd7eea801a0cb074
```
If the image downloads successfully, it worked!

You may need to create or update your credentials which you can do with a command like this. Set the environment variables to the proper values before running the command.
```bash
kubectl delete secret -n eksa-packages aws-secret
kubectl create secret -n eksa-packages generic aws-secret --from-literal=AWS_ACCESS_KEY_ID=${EKSA_AWS_ACCESS_KEY_ID} --from-literal=AWS_SECRET_ACCESS_KEY=${EKSA_AWS_SECRET_ACCESS_KEY}  --from-literal=REGION=${EKSA_AWS_REGION}
```

## Package on workload clusters

Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster. While interacting with the package resources by the following commands for a workload cluster, please make sure the kubeconfig is pointing to the **management cluster** that was used to create the workload cluster.

### Package manager is not managing packages on workload cluster

If the package manager is not managing packages on a workload cluster, make sure the management cluster has various resources for the workload cluster:
```bash
kubectl get packages,packagebundles,packagebundlecontrollers -A
```
You should see a PackageBundleController for the workload cluster named with the name of the workload cluster and the status should be set. There should be a namespace for the workload cluster as well:
```bash
kubectl get ns | grep eksa-packagess
```
Create a PackageBundlecController for the workload cluster if it does not exist (where billy here is the cluster name):
```bash
 cat <<! | kubectl apply -f -
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: PackageBundleController
metadata:
  name: billy
  namespace: eksa-packages
!
```
### Workload cluster is disconnected

Cluster is disconnected:
```bash
NAMESPACE       NAME                                                        ACTIVEBUNDLE   STATE               DETAIL
eksa-packages   packagebundlecontroller.packages.eks.amazonaws.com/billy                   disconnected        initializing target client: getting kubeconfig for cluster "billy": Secret "billy-kubeconfig" not found

```

In the example above, the secret does not exist which may be that the management cluster is not managing the cluster, the PackageBundleController name is wrong or the secret was deleted.

This also may happen if the management cluster cannot communicate with the workload cluster or the workload cluster was deleted, although the detail would be different.

### Error: the server doesn't have a resource type "packages"

All packages are remotely managed by the management cluster, and packages, packagebundles, and packagebundlecontrollers resources are all deployed on the management cluster. Please make sure the kubeconfig is pointing to the management cluster that was used to create the workload cluster while interacting with package-related resources.

### Error: packagebundlecontrollers.packages.eks.amazonaws.com "clusterName" not found

A package command run on a cluster that does not seem to be managed by the management cluster. To get a list of the clusters managed by the management cluster run the following command:
```
eksctl anywhere get packagebundlecontroller
NAME     ACTIVEBUNDLE   STATE     DETAIL
billy    v1-21-87       active
```
There will be one packagebundlecontroller for each cluster that is being managed. The only valid cluster name in the above example is `billy`.
