---
title: "Scale CloudStack cluster"
linkTitle: "Scale CloudStack cluster"
weight: 20
date: 2017-01-05
description: >
  How to scale your CloudStack cluster
---

### Autoscaling

EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package](../../../../reference/packagespec/cluster-autoscaler/).

See [here](../../../../reference/clusterspec/optional/autoscaling) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.
