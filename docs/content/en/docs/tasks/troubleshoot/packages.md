
---
title: "Curated Packages Troubleshooting"
linkTitle: "Curated Packages Troubleshooting"
weight: 50
description: >
  Troubleshooting specific to curated packages
aliases:
   - /docs/tasks/troubleshoot/_packages
---


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

The describe command might help to get more detail on why there is a problem
```bash
kubectl describe pods -n eksa-packages
```

Logs of the controller can be seen in a normal Kubernetes fashion
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

### Error: this command is currently not supported

```
Error: this command is currently not supported
```

Curated packages became generally available with version `v0.11.0`. Use the version command to make sure you are running version v0.11.0 or later:

```bash
eksctl anywhere version
```

### Package controller not running
If you do not see a pod or various resources for the package controller, it may be that it is not installed.

```
No resources found in eksa-packages namespace.
```

Most likely the cluster was created with an older version of the EKS Anywhere CLI. Curated packages became generally available with `v0.11.0`. Use the `eksctl anywhere version` command to verify you are running a new enough release and you can use the `eksctl anywhere install packagecontroller` command to install the package controller on an older release.

### Warnings during cluster creation

During cluster creation, you should see messages after the cluster is created when the package controller and any packages are installed.

```
🎉 Cluster created!
----------------------------------------------------------------------------------------------------------------
The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription.
----------------------------------------------------------------------------------------------------------------
Installing helm chart on cluster	{"chart": "eks-anywhere-packages", "version": "0.2.0-eks-a-v0.0.0-dev-build.3842"}
Warning: No AWS key/license provided. Please be aware this will prevent the package controller from installing curated packages.
```

If the `No AWS key/license provided` message appears during package controller installation, make sure you set and export the `EKSA_AWS_ACCESS_KEY_ID` and `EKSA_AWS_SECRET_ACCESS_KEY` variables to the access key and secret key of your AWS account. This will allow you to get access to container images in private ECR. A subscription is required to access the packages. If you forgot to set those values before cluster creation, the next section describes how you would create or update the secret after creation.

### ImagePullBackOff on Package or Package Controller

If a package or the package controller fails to start with ImagePullBackOff

```
NAME                                     READY   STATUS             RESTARTS   AGE
eks-anywhere-packages-6589449669-q7rjr   0/1     ImagePullBackOff   0          13h
```

This is most like because the machine running kubelet in your Kubernetes cluster cannot access the registry with the images or those images do not exist on that registry. Log into the machine and see if it has access to the images:

```bash
ctr image pull public.ecr.aws/eks-anywhere/eks-anywhere-packages@sha256:whateveritis
```

### Package pod cannot pull images
If a package pod cannot pull images, you may not have your AWS credentials set up properly. Verify that your credentials are working properly.

Make sure you are authenticated with the AWS CLI. Use the credentials you set up for packages. These credentials should have [limited capabilities]({{< relref "../packages/#setup-authentication-to-use-curated-packages" >}}):

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

You may need to create or update your credentials which you can do with a command like this. Set the environment variables to the proper values before running the command.
```bash
kubectl delete secret -n eksa-packages aws-secret
kubectl create secret -n eksa-packages generic aws-secret --from-literal=ID=${EKSA_AWS_ACCESS_KEY_ID} --from-literal=SECRET=${EKSA_AWS_SECRET_ACCESS_KEY}
```

If you recreate secrets, you can manually run the job to update the image pull secrets:
```bash
kubectl create job -n eksa-packages --from=cronjob/cron-ecr-renew run-it-now
```

### Error: cert-manager is not present in the cluster
```
Error: curated packages cannot be installed as cert-manager is not present in the cluster
```
This is most likely caused by an action to install curated packages at a workload cluster. Curated packages installation at workload cluster creation is currently not supported. The package manager will remotely manage packages on the workload cluster from the management cluster.

### Warning: not able to trigger cron job
```
secret/aws-secret created
Warning: not able to trigger cron job, please be aware this will prevent the package controller from installing curated packages.
```
This is most likely caused by an action to install curated packages in a cluster that is running `Kubernetes` at version `v1.20` or below. Note curated packages only support `Kubernetes` `v1.21` and above.

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
 cat <<! | k apply -f -
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

### Error: Error from server (NotFound): packagebundlecontrollers.packages.eks.amazonaws.com "clusterName" not found

A package command run run on a cluster that does not seem to be managed by the management cluster. To get a list of the clusters managed by the management cluster run the following command:
```
eksctl anywhere get packagebundlecontroller
NAME     ACTIVEBUNDLE   STATE     DETAIL
billy    v1-21-87       active
```
There will be one packagebundlecontroller for each cluster that is being managed. The only valid cluster name in the above example is `billy`.
