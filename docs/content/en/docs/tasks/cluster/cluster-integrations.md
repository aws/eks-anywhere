---
title: "Add cluster integrations"
linkTitle: "Add cluster integrations"
weight: 11
description: >
  How to add integrations to an EKS Anywhere cluster
---

EKS Anywhere offers AWS support for certain third-party vendor components,
namely Ubuntu TLS, Cilium, and Flux.
It also provides flexibility for you to integrate with your choice of tools in other areas.
Below is a list of example third-party tools your consideration.

For a full list of partner integration options, please visit [Amazon EKS Anywhere Partner page](https://aws.amazon.com/eks/eks-anywhere/partners/).

{{% alert title="Note" color="primary" %}}
The solutions listed on this page have not been tested by AWS and are not covered by the EKS Anywhere Support Subscription.
{{% /alert %}}

| Feature                       | Example third-party tools                 |
|-------------------------------|-------------------------------------------|
| Ingress controller            | [Gloo Edge](https://www.solo.io/products/gloo-edge/), [Emissary-ingress](https://www.getambassador.io/products/api-gateway/) (previously Ambassador)          |
| Service type load balancer    | [MetalLB]({{< relref "../workload/loadbalance/" >}})|
| Local container repository    | [Harbor]({{< relref "../packages/harbor" >}})                                    |
| Monitoring                    | [Prometheus](https://sysdig.com/products/monitor/prometheus-monitoring/), [Grafana](https://grafana.com/), [Datadog](https://www.datadoghq.com/blog/monitoring-kubernetes-with-datadog/), or [NewRelic](https://newrelic.com/platform/kubernetes/monitoring-guide) |
| Logging                 | [Splunk](https://www.splunk.com/en_us/blog/platform/introducing-the-splunk-operator-for-kubernetes.html) or [Fluentbit](https://fluentbit.io/kubernetes/)                                    |
| Secret management             | [Hashi Vault](https://www.vaultproject.io/docs/platform/k8s)                               |
| Policy agent                  | [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/kubernetes-introduction/)                                       |
| Service mesh                  | [Istio](https://istio.io/), [Gloo Mesh](https://www.solo.io/products/gloo-mesh/), or [Linkerd](https://linkerd.io/)                         |
| Cost management               | [KubeCost](https://www.kubecost.com/)                                  |
| Etcd backup and restore       | [Velero](https://velero.io/)                                    |
| Storage                       | Default storage class, any compatible CSI |

