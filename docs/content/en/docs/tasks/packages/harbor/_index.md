---
title: "Harbor package"
linkTitle: "Add Harbor"
weight: 13
date: 2022-04-12
description: >
  Install/upgrade/uninstall Harbor
---

{{% alert title="Important" color="warning" %}}

To install package controller, please follow the [installation guide.]({{< relref ".." >}})

{{% /alert %}}

## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package Harbor --source cluster -d .
   ```

1. Add the desired configuration to `curated-packages/my-harbor.yaml` 

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/harbor" >}}) for all configuration options and their default values.

   {{% alert title="Important" color="warning" %}}
   * All configuration options are listed in dot notations (e.g., `expose.tls.enabled`) in the doc, but they have to be transformed to **hierachical structures** when specified in the `config` section in the yaml spec.
   * Harbor web portal is exposed through `NodePort` by default, and its default port number is `30003` with TLS enabled and `30002` with TLS disabled.
   * TLS is enabled by default for connections to Harbor web portal, and a secret resource named `tls-secret` is required for that purpose. It can be provisioned through cert-manager or manually with the following command using self-signed certificate:
      ```bash
      kubectl create secret tls tls-secret --cert=[path to certificate file] --key=[path to key file] -n eksa-packages
      ```
   * The `UpdateStrategy` for deployments with persistent volumes (jobservice, registry and chartmuseum) has to be set to `Recreate` when `ReadWriteMany` for volumes isn't supported in the cluster. 

      `RollingUpdate` strategy works with `ReadWriteMany` volumes only.

   {{% /alert %}}

   TLS example with auto certificate generation
   ```yaml
   apiVersion: packages.eks.amazonaws.com/v1alpha1
   kind: Package
   metadata:
      name: my-harbor
      namespace: eksa-packages
   spec:
      packageName: Harbor
      config: |-
         externalURL: https://harbor.eksa.demo:30003
         expose:
            tls:
               certSource: auto
               auto:
                  commonName: "harbor.eksa.demo"
   ```

   Non-TLS example
   ```yaml
   apiVersion: packages.eks.amazonaws.com/v1alpha1
   kind: Package
   metadata:
      name: my-harbor
      namespace: eksa-packages
   spec:
      packageName: Harbor
      config: |-
         externalURL: http://harbor.eksa.demo:30002
         expose:
            tls:
               enabled: false
   ```

1. Install Harbor

   ```bash
   eksctl anywhere create packages -f curated-packages/my-harbor.yaml
   ```

1. Check Harbor

   ```bash
   eksctl anywhere get packages
   ```

   Example command output
   ```
   NAME        PACKAGE   AGE     STATE       CURRENTVERSION             TARGETVERSION        DETAIL
   my-harbor   Harbor    5m34s   installed   v2.4.1                     v2.4.1 (latest)
   ```

   Harbor web portal is accessible at whatever `externalURL` is set to. See [complete configuration options]({{< relref "../../../reference/packagespec/harbor" >}}) for all default values.

   ![Harbor web portal](/images/harbor-portal.png)

## Upgrade
1. Verify a new bundle is available
   ```bash
   eksctl anywhere get packagebundle
   ```

   Example command output
   ```bash
   NAME         VERSION   STATE
   v1.21-1000   1.21      active (upgrade available)
   v1.21-1001   1.21      inactive
   ```

1. Upgrade Harbor
   ```bash
   eksctl anywhere upgrade packages --bundleversion v1.21-1001
   ```

1. Check Harbor

   ```bash
   eksctl anywhere get packages
   ```

   Example command output
   ```
   NAME        PACKAGE   AGE     STATE       CURRENTVERSION             TARGETVERSION        DETAIL
   my-harbor   Harbor    14m     installed   v2.4.2                     v2.4.2 (latest)
   ```

## Uninstall
1. Uninstall Harbor

   {{% alert title="Important" color="warning" %}}

   * By default, PVCs created for jobservice, registry and chartmuseum are not removed during a package delete operation, which can be changed by leaving `persistence.resourcePolicy` empty. 

   {{% /alert %}}
   ```bash
   eksctl anywhere delete package my-harbor
   ```
