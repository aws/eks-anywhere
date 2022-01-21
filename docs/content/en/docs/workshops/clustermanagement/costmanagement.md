---
title: "Cost management"
linkTitle: "Cost management"
weight: 9
date: 2021-11-11
description: >  
---

## Overview

We will see how to integrate kubecost to get real-time cost visibility and insights for teams using Kubernetes, helping you continuously reduce your infrastructure costs.

**Kubecost** helps you monitor and manage costs and capacity in Kubernetes environments.
It integrates with your infrastructure to help your team track, manage, and reduce spending.

## Installation

{{% alert title="Notes" color="primary" %}}
For demostration purpose, we will be installing `kubecost` with persistance **disabled**. We wouldn't recommend this setup for production.
{{% /alert %}}

### Kubecost install

1. Running the following commands will install Kubecost, along with Prometheus, Grafana, and kube-state-metrics in the namespace supplied.
View install configuration options [here](https://github.com/kubecost/cost-analyzer-helm-chart/blob/master/README.md#config-options).

    ```bash
    kubectl create namespace kubecost
    helm repo add kubecost https://kubecost.github.io/cost-analyzer/
    helm upgrade -i --create-namespace kubecost kubecost/cost-analyzer \
    --namespace kubecost \
    --set kubecostToken="aGVsbUBrdWJlY29zdC5jb20=xm343yadf98" \
    --set persistentVolume.enabled=false \
    --set prometheus.server.persistentVolume.enabled=false \
    --set prometheus.alertmanager.persistentVolume.enabled=false \
    --set prometheus.pushgateway.persistentVolume.enabled=false
    ```

{{% alert title="Notes" color="warning" %}}
All persistent volumes are disabled by default as part of the install.
{{% /alert %}}

2. In a few minutes all the pods related to kubecost will be up and running. You can verify this by running the following command:

    ```bash
    kubectl get pods -n kubecost
    ```

    **Output**

    ```
    NAME                                           READY   STATUS    RESTARTS   AGE
    kubecost-cost-analyzer-88bfdfff4-mx5vv         3/3     Running   0          53m
    kubecost-grafana-7cf6f7bc85-kzg2g              3/3     Running   0          53m
    kubecost-kube-state-metrics-6f69d8fd6b-2jkl4   1/1     Running   0          53m
    kubecost-prometheus-node-exporter-cwtql        1/1     Running   0          53m
    kubecost-prometheus-node-exporter-tm5qs        1/1     Running   0          53m
    kubecost-prometheus-server-7c7ff6ddcd-f8r82    2/2     Running   0          53m
    ```

3. Enable port forwarding for the kubecost dashboard, by running the following command:

    ```bash
    kubectl port-forward --namespace kubecost deployment/kubecost-cost-analyzer 9090
    ```

4. You can now view the deployed frontend by visiting the following link. Publish :9090 as a secure endpoint on your cluster to remove the need to port forward.

    ```bash
    http://localhost:9090
    ```

    **Output**
    ![dashboard](../images/dashboard.png)

## Cleanup

To uninstall Kubecost and its dependencies, run the following command:

```bash
helm uninstall kubecost -n kubecost
```
