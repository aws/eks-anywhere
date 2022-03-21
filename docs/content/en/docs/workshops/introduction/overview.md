---
title: "Overview"
linkTitle: "Overview"
weight: 1
date: 2021-11-11
description: >  
---

## What is the purpose of this workshop?

The purpose of this workshop is to provide a more perscriptive walkthrough of building, deploying, and operating an EKS Anywhere cluster. This will use existing content from the documentation, just in a more condensed format for those wishing to get started. 

## EKS Anywhere Overview

Amazon EKS Anywhere is a new deployment option for Amazon EKS that allows customers to create and operate Kubernetes clusters on customer-managed infrastructure, supported by AWS. Customers can now run Amazon EKS Anywhere on their own on-premises infrastructure using VMware vSphere starting today, with support for other deployment targets in the near future, including support for bare metal coming in 2022.

Amazon EKS Anywhere helps simplify the creation and operation of on-premises Kubernetes clusters with default component configurations while providing tools for automating cluster management. It builds on the strengths of Amazon EKS Distro: the same Kubernetes distribution that powers Amazon EKS on AWS. AWS supports all Amazon EKS Anywhere components including the integrated 3rd-party software, so that customers can reduce their support costs and avoid maintenance of redundant open-source and third-party tools. In addition, Amazon EKS Anywhere gives customers on-premises Kubernetes operational tooling that’s consistent with Amazon EKS. You can leverage the EKS console to view all of your Kubernetes clusters (including EKS Anywhere clusters) running anywhere, through the [EKS Connector](https://docs.aws.amazon.com/eks/latest/userguide/aws-connector.html) (public preview)

![how-it-works](../images/how-it-works.png)