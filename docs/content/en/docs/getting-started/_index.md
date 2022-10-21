---
title: Getting started
main_menu: true
weight: 20
no_list: true
card:
  name: setup
  weight: 20
  anchors:
  - anchor: "#local-environment"
    title: Local environment
  - anchor: "#production-environment"
    title: Production environment
description: >
  The Getting started section includes information on starting to set up your own EKS Anywhere local or production environment.
---

<!-- overview -->

EKS Anywhere can be deployed as a simple, unsupported local environment or as a production-quality environment that can become a supported on-premises Kubernetes platform.
This section lists the different ways to set up and run EKS Anywhere.
When you install EKS Anywhere, choose an installation type based on: ease of maintenance, security, control, available resources, and expertise required to operate and manage a cluster.

<!-- body -->

## Install EKS Anywhere

To create an EKS Anywhere cluster you'll need to download the command line tool that is used to create and manage a cluster.
You can install it using the [installation guide]({{< relref "install" >}})

## Local environment

If you just want to try out EKS Anywhere, there is a single-system method for installing and running EKS Anywhere using Docker.
See [EKS Anywhere local environment]({{< relref "local-environment" >}}).

## Production environment

When evaluating a solution for a production environment
consider deploying EKS Anywhere on providers listed on the [Create production cluster]({{< relref "production-environment/" >}}) page.
