---
title: "RECOMMENDED: Kube-Vip for Service-type Load Balancer"
linkTitle: "Kube-vip Service-type Load Balancer Setup"
weight: 30
date: 2017-01-05
description: >
  How to set up kube-vip for Service-type Load Balancer (Recommended)
---

<!-- overview -->

We recommend using Kube-Vip cloud controller to expose your services as service-type Load Balancer.
Detailed information about Kube-Vip can be found [here](https://kube-vip.io/).

There are two ways Kube-Vip can manage virtual IP addresses on your network.
Please see the following guides for ARP or BGP mode depending on your on-prem networking preferences.

## Setting up Kube-Vip for Service-type Load Balancer

Kube-vip Service-type Load Balancer can be set up in either ARP mode or BGP mode
