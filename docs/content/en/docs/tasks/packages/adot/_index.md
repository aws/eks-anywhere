---
title: "AWS Distro for OpenTelemetry (ADOT)"
linkTitle: "Add ADOT"
weight: 13
date: 2022-09-21
description: >
  Install/upgrade/uninstall ADOT
---

{{< content "../prereq.md" >}}


## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package adot --cluster clusterName > adot.yaml
   ```

1. Add the desired configuration to `adot.yaml`

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/adot" >}}) for all configuration options and their default values.

   Example package file with `daemonSet` mode and default configuration:
   ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-adot
      namespace: eksa-packages
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
      namespace: eksa-packages
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
   eksctl anywhere get packages
   ```

   Example command output
   ```
   NAME   PACKAGE   AGE   STATE       CURRENTVERSION                                                            TARGETVERSION                                                                   DETAIL
   my-adot   adot   19h   installed   0.21.1-1ba95f7be1f47c40a23956363d1eb836e60c0cef   0.21.1-1ba95f7be1f47c40a23956363d1eb836e60c0cef (latest)
   ```

## Upgrade

ADOT will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall ADOT, simply delete the package

```bash
eksctl anywhere delete package my-adot
```
