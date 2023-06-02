---
title: "EKS Anywhere curated packages"
linkTitle: "Curated packages"
weight: 60
date: 2022-05-09
description: >
  All information you may need for EKS Anywhere curated packages
---

{{% alert title="Note" color="primary" %}}
The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here.](https://aws.amazon.com/contact-us/sales-support-eks/)
{{% /alert %}}

### Overview
Amazon EKS Anywhere Curated Packages are Amazon-curated software packages that extend the core functionalities of Kubernetes on your EKS Anywhere clusters. If you operate EKS Anywhere clusters on-premises, you probably install additional software to ensure the security and reliability of your clusters. However, you may be spending a lot of effort researching for the right software, tracking updates, and testing them for compatibility. Now with the EKS Anywhere Curated Packages, you can rely on Amazon to provide trusted, up-to-date, and compatible software that are supported by Amazon, reducing the need for multiple vendor support agreements. 

* *Amazon-built*: All container images of the packages are built from source code by Amazon, including the open source (OSS) packages. OSS package images are built from the open source upstream.
* *Amazon-scanned*: Amazon scans the container images including the OSS package images daily for security vulnerabilities and provides remediation.
* *Amazon-signed*: Amazon signs the package bundle manifest (a Kubernetes manifest) for the list of curated packages. The manifest is signed with AWS Key Management Service (AWS KMS) managed private keys. The curated packages are installed and managed by a package controller on the clusters. Amazon provides validation of signatures through an admission control webhook in the package controller and the public keys distributed in the bundle manifest file. 
* *Amazon-tested*: Amazon tests the compatibility of all curated packages including the OSS packages with each new version of EKS Anywhere.
* *Amazon-supported*: All curated packages including the curated OSS packages are supported under the EKS Anywhere Support Subscription. 

The main components of EKS Anywhere Curated Packages are the [package controller]({{< relref "../packages/overview#package-controller" >}}), the [package build artifacts]({{< relref "../packages/overview#curated-packages-artifacts" >}}) and the [command line interface]({{< relref "../packages/overview#packages-cli" >}}). The package controller will run in a pod in an EKS Anywhere cluster. The package controller will manage the lifecycle of all curated packages.

### Curated packages
Please check out [curated package list]({{< relref "../packages/packagelist/" >}}) for the complete list of EKS Anywhere curated packages.


### FAQ
1. *Can I install software not from the curated package list?*

    Yes. You can install any optional software of your choice. Be aware you cannot use EKS Anywhere tooling to install or update your self-managed software. Amazon does not provide testing, security patching, software updates, or customer support for your self-managed software.


2. *Can I install software that’s on the curated package list but not sourced from EKS Anywhere repository?*

    If, for example, you deploy a Harbor image that is not built and signed by Amazon, Amazon will not provide testing or customer support to your self-built images.

### Curated package list

| Name                       | Description                | Versions                  | GitHub                      |
|----------------------------|----------------------------|---------------------------|-----------------------------|
| [ADOT]({{< relref "../packages/adot" >}}) | ADOT Collector is an AWS distribution of the OpenTelemetry Collector, which provides a vendor-agnostic solution to receive, process and export telemetry data. | [v0.25.0]({{< relref "../packages/adot/v0.25.0.md" >}}) | https://github.com/aws-observability/aws-otel-collector |
| [Cert-manager]({{< relref "../packages/cert-manager" >}}) | Cert-manager is a certificate manager for Kubernetes clusters. | [v1.9.1]({{< relref "../packages/cert-manager/v1.9.1.md" >}}) | https://github.com/cert-manager/cert-manager |
| [Cluster Autoscaler]({{< relref "../packages/cluster-autoscaler" >}}) | Cluster Autoscaler is a component that automatically adjusts the size of a Kubernetes Cluster so that all pods have a place to run and there are no unneeded nodes. | [v9.21.0]({{< relref "../packages/cluster-autoscaler/v9.21.0.md" >}}) | https://github.com/kubernetes/autoscaler |
| [Emissary Ingress]({{< relref "../packages/emissary" >}}) | Emissary Ingress is an open source `Ingress` supporting API Gateway + Layer 7 load balancer built on Envoy Proxy. | [v3.3.0]({{< relref "../packages/emissary/v3.3.0.md" >}}) | https://github.com/emissary-ingress/emissary/ |
| [Harbor]({{< relref "../packages/harbor" >}}) | Harbor is an open source trusted cloud native registry project that stores, signs, and scans content. | [v2.7.1]({{< relref "../packages/harbor/v2.7.1.md" >}})<br> [v2.5.1]({{< relref "../packages/harbor/v2.7.1.md" >}}) | https://github.com/goharbor/harbor<br>https://github.com/goharbor/harbor-helm |
| [MetalLB]({{< relref "../packages/metallb" >}}) | MetalLB is a virtual IP provider for services of type `LoadBalancer` supporting ARP and BGP. | [v0.13.7]({{< relref "../packages/metallb/v0.13.7.md" >}}) | https://github.com/metallb/metallb/ |
| [Metrics Server]({{< relref "../packages/metrics-server" >}}) | Metrics Server is a scalable, efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines. | [v3.8.2]({{< relref "../packages/metrics-server/v3.8.2.md" >}}) | https://github.com/kubernetes-sigs/metrics-server |
| [Prometheus]({{< relref "../packages/prometheus" >}}) | Prometheus is an open-source systems monitoring and alerting toolkit that collects and stores metrics as time series data. | [v2.41.0]({{< relref "../packages/prometheus/v2.41.0.md" >}}) | https://github.com/prometheus/prometheus |


