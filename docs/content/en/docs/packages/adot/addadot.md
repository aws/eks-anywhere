---
title: "AWS Distro for OpenTelemetry (ADOT)"
linkTitle: "Add ADOT"
weight: 13
aliases:
    /docs/reference/packagespec/adot/
date: 2022-09-21
description: >
  Install/upgrade/uninstall ADOT
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
   eksctl anywhere generate package adot --cluster <cluster-name> > adot.yaml
   ```

1. Add the desired configuration to `adot.yaml`

   Please see [complete configuration options]({{< relref "../adot" >}}) for all configuration options and their default values.

   Example package file with `daemonSet` mode and default configuration:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-adot
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: adot
      targetNamespace: observability
      config: | 
        mode: daemonset
   ```

   Example package file with `deployment` mode and customized collector components to scrap
   ADOT collector's own metrics:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-adot
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: adot
      targetNamespace: observability
      config: | 
        mode: deployment
        replicaCount: 2
        config:
          receivers:
            prometheus:
              config:
                scrape_configs:
                  - job_name: opentelemetry-collector
                    scrape_interval: 10s
                    static_configs:
                      - targets:
                          - ${MY_POD_IP}:8888
          processors:
            batch: {}
            memory_limiter: null
          exporters:
            logging:
              loglevel: debug
            prometheusremotewrite:
              endpoint: "<prometheus-remote-write-end-point>"
          extensions:
            health_check: {}
            memory_ballast: {}
          service:
            pipelines:
              metrics:
                receivers: [prometheus]
                processors: [batch]
                exporters: [logging, prometheusremotewrite]
            telemetry:
              metrics:
                address: 0.0.0.0:8888
   ```

1. Create the namespace
  (If overriding `targetNamespace`, change `observability` to the value of `targetNamespace`)
   ```bash
   kubectl create namespace observability
   ```

1. Install adot

   ```bash
   eksctl anywhere create packages -f adot.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAME   PACKAGE   AGE   STATE       CURRENTVERSION                                                            TARGETVERSION                                                                   DETAIL
   my-adot   adot   19h   installed   0.25.0-c26690f90d38811dbb0e3dad5aea77d1efa52c7b   0.25.0-c26690f90d38811dbb0e3dad5aea77d1efa52c7b (latest)
   ```

## Update
To update package configuration, update adot.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f adot.yaml
```

## Upgrade

ADOT will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall ADOT, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> my-adot
```
