---
title: "Prometheus & Grafana"
linkTitle: "Prometheus & Grafana"
weight: 1
date: 2021-11-11
description: >  
---

## Overview

We will see how to install prometheus, deploy a sample workload and view the metrics using Grafana.

**Prometheus** is a monitoring solution for recording and processing any purely numeric time-series. It gathers, organizes, and stores metrics along with unique identifiers and timestamps. Prometheus is open-source software that collects metrics from targets by "scraping" metrics HTTP endpoints.

**Grafana** is a multi-platform open source analytics and interactive visualization web application. It provides charts, graphs, and alerts for the web when connected to supported data sources, that allows you to visualize and analyze your data.

## Installation

{{% alert title="Notes" color="primary" %}}
Since EKS Anywhere cluster comes preinstalled with Helm we will be using it for setting up Prometheus and Grafana.
{{% /alert %}}

1. Run the following command to add the `helm` repository:

    ```bash
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    ```

2. Run the following command to create a new namespace and install Prometheus and Grafana inside the namespace.

    ```bash
    kubectl create ns prom
    helm install prometheus prometheus-community/kube-prometheus-stack -n prom
    ```

    If you are getting a validation error while running the install, like below:

    ```bash
    Error: unable to build kubernetes objects from release manifest: error validating "": error validating data: [ValidationError(Alertmanager.spec): unknown field "alertmanagerConfigNamespaceSelector" in com.coreos.monitoring.v1.Alertmanager.spec, ValidationError(Alertmanager.spec): unknown field "alertmanagerConfigSelector" in com.coreos.monitoring.v1.Alertmanager.spec]
    ```

    Run the following command to fix the validation error and re-run the install command:

    ```bash
    kubectl delete crd alertmanagerconfigs.monitoring.coreos.com
    kubectl delete crd alertmanagers.monitoring.coreos.com
    kubectl delete crd podmonitors.monitoring.coreos.com
    kubectl delete crd probes.monitoring.coreos.com
    kubectl delete crd prometheuses.monitoring.coreos.com
    kubectl delete crd prometheusrules.monitoring.coreos.com
    kubectl delete crd servicemonitors.monitoring.coreos.com
    kubectl delete crd thanosrulers.monitoring.coreos.com
    ```

3. Check the status of deployment by running the following command.

    ```bash
    kubectl get pods -n prom
    ```

    **Output**

    ```bash
    alertmanager-prometheus-kube-prometheus-alertmanager-0   2/2     Running   0          21s
    prometheus-grafana-788747867d-86swc                      2/2     Running   0          24s
    prometheus-kube-prometheus-operator-686b89b849-hc2xj     1/1     Running   0          24s
    prometheus-kube-state-metrics-58c5cd6ddb-lpd8s           1/1     Running   0          24s
    prometheus-prometheus-kube-prometheus-prometheus-0       2/2     Running   0          21s
    prometheus-prometheus-node-exporter-rz6s4                1/1     Running   0          24s
    prometheus-prometheus-node-exporter-wsl6c                1/1     Running   0          24s
    ```

    > **Note:** All the pods should be up and running.

4. Deploy a sample workload to the cluster, by running the following command:

    ```bash
    kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
    ```

    **Output**

    ```bash
    deployment.apps/hello-eks-a created
    service/hello-eks-a created
    ```

5. (Only for dev/local deployment) View Prometheus - Run the following command to port forward the Prometheus service:

    ```bash
    kubectl port-forward -n prom service/prometheus-kube-prometheus-prometheus 9090
    ```

    **Output**

    ![prom](../images/prom.png)

6. (Only for dev/local deployment) View Grafana - Run the following command to port forward the Prometheus service to localhost:

    ```bash
    kubectl port-forward -n prom deployment/prometheus-grafana 3000
    ```

## Visualize the metrics

1. Open the browser and navigate to `http://localhost:3000/` to see the dashboard. You the default username `admin` and password `prom-operator` to login to the dashboard. Here is the default console view:

    ![dashboard](../images/dashboard.png)

2. Click on **Manage** submenu under **Dashboard** (in the sidecar) and select `Kubernetes / Compute Resources / Pod` to see the metrics of all the pods. Here is the view of the pods running the sample application:

    ![metrics](../images/metrics.png)

    **Note:** Make sure to select the right namespace and pod name.

## Cleanup

To uninstall Promethus and its dependencies, run the following command:

```bash
helm uninstall prometheus -n prom
kubectl delete -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
```

**Output**

```bash
release "prometheus" uninstalled
deployment.apps "hello-eks-a" deleted
service "hello-eks-a" deleted
```
