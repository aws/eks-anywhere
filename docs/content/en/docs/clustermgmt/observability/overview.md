---
title: "Overview"
linkTitle: "Overview"
date: 2017-01-05
weight: 5
description: >
  Overview of observability in EKS Anywhere
---

Most Kubernetes-conformant observability tools can be used with EKS Anywhere. You can optionally use the EKS Connector to view your EKS Anywhere cluster resources in the Amazon EKS console, reference the [Connect to console page]({{< relref "./cluster-connect" >}}) for details. EKS Anywhere includes the [AWS Distro for Open Telemetry (ADOT)]({{< relref "../../packages/adot/addadot" >}}) and [Prometheus]({{< relref "../../packages/prometheus/addpro" >}}) for metrics and tracing as EKS Anywhere Curated Packages. You can use popular tooling such as [Fluent Bit](https://docs.fluentbit.io/manual) for logging, and can track the progress of logging for ADOT on the [AWS Observability roadmap](https://github.com/aws-observability/aws-otel-community/issues/11). For more information on EKS Anywhere Curated Packages, reference the [Package Management Overview]({{< relref "../../packages/overview" >}}).

### AWS Integrations ###

AWS offers comprehensive monitoring, logging, alarming, and dashboard capabilities through services such as [Amazon CloudWatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/WhatIsCloudWatch.html), [Amazon Managed Prometheus (AMP)](https://docs.aws.amazon.com/prometheus/latest/userguide/what-is-Amazon-Managed-Service-Prometheus.html), and [Amazon Managed Grafana (AMG)](https://docs.aws.amazon.com/grafana/latest/userguide/what-is-Amazon-Managed-Service-Grafana.html). With CloudWatch, you can take advantage of a highly scalable, AWS-native centralized logging and monitoring solution for EKS Anywhere clusters. With AMP and AMG, you can monitor your containerized applications EKS Anywhere clusters at scale with popular Prometheus and Grafana interfaces.

### Resources ###
1. [Verify EKS Anywhere cluster status]({{< relref "./cluster-verify" >}})
1. [Use the EKS Connector to view EKS Anywhere clusters and resources in the EKS console]({{< relref "./cluster-connect" >}})
1. [Use Fluent Bit and Container Insights to send metrics and logs to CloudWatch]({{< relref "./fluentbit-logging" >}})
1. [Use ADOT to send metrics to AMP and AMG](https://aws.amazon.com/blogs/mt/using-curated-packages-and-aws-managed-open-source-services-to-observe-your-on-premise-kubernetes-environment/)