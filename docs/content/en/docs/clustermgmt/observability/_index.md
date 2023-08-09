---
title: "Observability in EKS Anywhere"
linkTitle: "Observability in EKS Anywhere"
date: 2017-01-05
weight: 40
description: >
  Observability in EKS Anywhere
---

Amazon Web Services (AWS) offers comprehensive monitoring, logging, alarming, and dashboard capabilities through Amazon CloudWatch. This powerful tool allows for centralized logging for EKS Anywhere, an on-premises version of AWS's managed Kubernetes. With CloudWatch, organizations can take advantage of a highly scalable, native AWS logging and monitoring solution.

CloudWatch Logs delivers enhanced log visibility, independent of the log source. CloudWatch Logs is available for EKS environments, as well as for EKS Anywhere providers such as vSphere, Bare Metal, Snow, CloudStack, and Nutanix. This unified solution ensures a consistent flow of events, seamlessly ordered by time, facilitating ease of analysis and troubleshooting.

The flexibility of CloudWatch Logs enables users to perform queries and sorting based on various dimensions, providing a deep level of granularity in log analysis. Additionally, logs can be grouped using specific fields, allowing for more focused insights and targeted problem-solving.

Amazon CloudWatch also facilitates the creation of intuitive, data-rich dashboards, enabling log data visualization for enhanced monitoring and decision-making.

## Observability in EKS Anywhere
Step through the following content to configure EKS Anywhere Cluster to scrape metrics and send them to [Amazon CloudWatch](https://aws.amazon.com/cloudwatch/) (CloudWatch).


1. [Create a cluster with IAM Roles for Service Account (IRSA)](#create-a-cluster-with-irsa).
1. [Verify Cluster]({{< relref "./cluster-verify" >}}).
1. [Install Fluentbit and Configure an EKS service account to assume an IAM role]({{< relref "./loggingSetup" >}}).
1. [Deploy Test Application]({{< relref "../../workloadmgmt/test-app" >}})

{{% alert title="Note" color="primary" %}}
- We included `Test` sections for critical steps to help users to validate they have completed such procedure properly. We recommend going through them in sequence as checkpoints of the progress.
- We recommend creating all resources in the `us-west-2` region.
{{% /alert %}}

## Create a cluster with IRSA
Create a cluster with IAM Roles for Service Account (IRSA). An IAM role can be associated with a Kubernetes service account. This service account can then provide AWS permissions to the containers in any pod that uses that service account. With this feature, there will be no need to hardcode AWS security credentials as environment variables on your machine. [EKS Anywhere cluster spec for Pod IAM]({{< relref "../../getting-started/optional/irsa/" >}}) gives step-by-step guidance on how to do so.