---
title: "RECOMMENDED: Kube-Vip for Service-type Load Balancer"
linkTitle: "Kube-vip Service-type Load Balancer Setup"
weight: 30
date: 2017-01-05
description: >
  How to set up kube-vip for Service-type Load Balancer (Recommended)
---

<!-- overview -->

The purpose of this document is to walk you through getting set up with the Kube-vip Kubernetes Load Balancer for EKS Anywhere.

<!-- body -->

We currently recommend using Kube-Vip Kubernetes Service-type Load Balancer. Previously designed to support control-plane resiliency, it has since been expanded to provide load-balancing for applications and services within a Kubernetes cluster. Detailed information about Kube-Vip can be found [here](https://kube-vip.io/).

## Setting up Kube-Vip for Service-type Load Balancer

Kube-vip Service-type Load Balancer can be set up in either ARP mode or BGP mode
