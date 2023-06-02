---
title: "Deploy test workload"
linkTitle: "Deploy test workload"
weight: 11
aliases:
    /docs/tasks/workload/test-app/
date: 2017-01-05
description: >
  How to deploy a workload to check that your cluster is working properly
---

We've created a simple test application for you to verify your cluster is working properly.
You can deploy it with the following command:

```bash
kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
```

To see the new pod running in your cluster, type:

```bash
kubectl get pods -l app=hello-eks-a
```
Example output:
```
NAME                                     READY   STATUS    RESTARTS   AGE
hello-eks-a-745bfcd586-6zx6b   1/1     Running   0          22m
```

To check the logs of the container to make sure it started successfully, type:

```bash
kubectl logs -l app=hello-eks-a
```

There is also a default web page being served from the container.
You can forward the deployment port to your local machine with

```bash
kubectl port-forward deploy/hello-eks-a 8000:80
```

Now you should be able to open your browser or use `curl` to [http://localhost:8000](http://localhost:8000) to view the page example application.

```bash
curl localhost:8000
```
Example output:

```
⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

Thank you for using

███████╗██╗  ██╗███████╗
██╔════╝██║ ██╔╝██╔════╝
█████╗  █████╔╝ ███████╗
██╔══╝  ██╔═██╗ ╚════██║
███████╗██║  ██╗███████║
╚══════╝╚═╝  ╚═╝╚══════╝

 █████╗ ███╗   ██╗██╗   ██╗██╗    ██╗██╗  ██╗███████╗██████╗ ███████╗
██╔══██╗████╗  ██║╚██╗ ██╔╝██║    ██║██║  ██║██╔════╝██╔══██╗██╔════╝
███████║██╔██╗ ██║ ╚████╔╝ ██║ █╗ ██║███████║█████╗  ██████╔╝█████╗  
██╔══██║██║╚██╗██║  ╚██╔╝  ██║███╗██║██╔══██║██╔══╝  ██╔══██╗██╔══╝  
██║  ██║██║ ╚████║   ██║   ╚███╔███╔╝██║  ██║███████╗██║  ██║███████╗
╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝

You have successfully deployed the hello-eks-a pod hello-eks-a-c5b9bc9d8-qp6bg

For more information check out
https://anywhere.eks.amazonaws.com

⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢
```

If you would like to expose your applications with an external load balancer or an ingress controller, you can follow the steps in [Adding an external load balancer]({{< relref "../packages/metallb" >}}).
