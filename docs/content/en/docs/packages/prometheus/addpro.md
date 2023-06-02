---
title: "Prometheus"
linkTitle: "Add Prometheus"
weight: 13
date: 2022-09-21
description: >
  Install/upgrade/uninstall Prometheus
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

   {{% alert title="Important" color="warning" %}}
   * Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
   * While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
   {{% /alert %}}

## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package prometheus --cluster <cluster-name> > prometheus.yaml
   ```

1. Add the desired configuration to `prometheus.yaml`

   Please see [complete configuration options]({{< relref "../prometheus" >}}) for all configuration options and their default values.

   Example package file with default configuration, which enables prometheus-server and node-exporter:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: generated-prometheus
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: prometheus
   ```

   Example package file with prometheus-server (or node-exporter) disabled:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: generated-prometheus
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: prometheus
      config: |
        # disable prometheus-server
        server:
          enabled: false
        # or disable node-exporter
        # nodeExporter:
        #   enabled: false
   ```

   Example package file with prometheus-server deployed as a statefulSet with replicaCount 2, and set scrape config to collect Prometheus-server's own metrics only:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: generated-prometheus
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: prometheus
      targetNamespace: observability
      config: |
        server:
          replicaCount: 2
          statefulSet:
            enabled: true
        serverFiles:
          prometheus.yml:
            scrape_configs:
              - job_name: prometheus
                static_configs:
                  - targets:
                    - localhost:9090
   ```

1. Create the namespace
  (If overriding `targetNamespace`, change `observability` to the value of `targetNamespace`)
    ```bash
    kubectl create namespace observability
    ```

1. Install prometheus

    ```bash
    eksctl anywhere create packages -f prometheus.yaml
    ```

1. Validate the installation

    ```bash
    eksctl anywhere get packages --cluster <cluster-name>
    ```

   Example command output
    ```
    NAMESPACE                      NAME                   PACKAGE      AGE   STATE       CURRENTVERSION                                    TARGETVERSION                                              DETAIL
    eksa-packages-<cluster-name>   generated-prometheus   prometheus   17m   installed   2.41.0-b53c8be243a6cc3ac2553de24ab9f726d9b851ca   2.41.0-b53c8be243a6cc3ac2553de24ab9f726d9b851ca (latest)
    ```

## Update
To update package configuration, update prometheus.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f prometheus.yaml
```

## Upgrade

Prometheus will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall Prometheus, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> generated-prometheus
```
