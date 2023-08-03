---
title: "Observability in EKS Anywhere"
linkTitle: "Observability in EKS Anywhere"
weight: 2
#aliases:
#    /docs/clustermgmt/oservability/
date: 2023-07-28
description: >  
---

This tutorial demonstrates how to configure EKS Anywhere Cluster to scrape metrics and send them to [Amazon CloudWatch](https://aws.amazon.com/cloudwatch/) (CloudWatch).

This tutorial walks through the following procedures:
- [Create a cluster with IAM Roles for Service Account (IRSA)](#create-a-cluster-with-irsa).
- [Verify Cluster]({{< relref "./cluster-verify" >}}).
- [Install Fluentbit and Configure an EKS service account to assume an IAM role]({{< relref "./loggingSetup" >}}).
- [Deploy Test Application](https://anywhere.eks.amazonaws.com/docs/workloadmgmt/test-app/)

{{% alert title="Note" color="primary" %}}

- We included `Test` sections below for critical steps to help users to validate they have completed such procedure properly. We recommend going through them in sequence as checkpoints of the progress.
- We recommend creating all resources in the `us-west-2` region.

{{% /alert %}}

## Create a cluster with IRSA
IAM role can be associated with a Kubernetes service account. This service account can then provide AWS permissions to the containers in any pod that uses that service account. With this feature, there will be no need to hardcode AWS security credentials as environment variables on your machine. [EKS Anywhere cluster spec for Pod IAM]({{< relref "../../getting-started/optional/irsa/" >}}) gives step-by-step guidance on how to do so.