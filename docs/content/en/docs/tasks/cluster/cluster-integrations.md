---
title: "Add integrations to cluster"
linkTitle: "Add integrations to cluster"
weight: 11
description: >
  How to add integrations to an EKS-A cluster
---

EKS Anywhere offers AWS support for certain third-party vendor components,
namely Ubuntu TLS, Cilium, and Flux.
It also provides flexibility for you to integrate with your choice of tools in other areas.
Below is a list of suggested third-party tools your consideration.

| Feature                       | Suggest third-party tools                 |
|-------------------------------|-------------------------------------------|
| Ingress controller            | [Emissary-ingress](https://www.getambassador.io/products/api-gateway/) (previously Ambassador)          |
| Service type load balancer    | [KubeVip]({{< relref "../workload/loadbalance.md#current-recommendation-kube-vip" >}}) (recommended) or [Metal LB]({{< relref "../workload/loadbalance.md#alternatives" >}})|
| Local container repository    | [Harbor](https://goharbor.io/)                                    |
| Monitoring                    | [Prometheus](https://sysdig.com/products/monitor/prometheus-monitoring/)+[Grafana](https://grafana.com/)+[FluentBit](https://fluentbit.io/kubernetes/), [Datadog](https://www.datadoghq.com/blog/monitoring-kubernetes-with-datadog/), or [NewRelic](https://newrelic.com/platform/kubernetes/monitoring-guide)  |
| Log analytics                 | [Splunk](https://www.splunk.com/en_us/blog/platform/introducing-the-splunk-operator-for-kubernetes.html)                                    |
| Secret management             | [Hashi Vault](https://www.vaultproject.io/docs/platform/k8s)                               |
| Policy agent                  | [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/kubernetes-introduction/)                                       |
| Service mesh                  | [Linkerd](https://linkerd.io/) or [Istio](https://istio.io/)                         |
| Infrastructure-as-code        | [Terraform](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/guides/getting-started)                                 |
| Cost management               | [KubeCost](https://www.kubecost.com/)                                  |
| Etcd backup and restore       | [Velero](https://velero.io/)                                    |
| Storage                       | Default storage class, any compatible CSI |
