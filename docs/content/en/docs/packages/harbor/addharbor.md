---
title: "Harbor"
linkTitle: "Add Harbor"
weight: 10
date: 2022-04-12
description: >
  Install/upgrade/uninstall Harbor
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

  {{% alert title="Important" color="warning" %}}
   * Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
   * While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
   {{% /alert %}}

## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->

1. Set the `KUBECONFIG` environment variable to use the config of the management cluster

   ```bash
   export KUBECONFIG=<path to management cluster kubeconfig>
   ```

1. Generate the package configuration
   ```bash
   eksctl anywhere generate package harbor --cluster <cluster-name> > harbor.yaml
   ```

1. Add the desired configuration to `harbor.yaml` 

   Please see [complete configuration options]({{< relref "../harbor" >}}) for all configuration options and their default values.

   {{% alert title="Important" color="warning" %}}
   * All configuration options are listed in dot notations (e.g., `expose.tls.enabled`) in the doc, but they have to be transformed to **hierachical structures** when specified in the `config` section in the YAML spec.
   * Harbor web portal is exposed through `NodePort` by default, and its default port number is `30003` with TLS enabled and `30002` with TLS disabled.
   * TLS is enabled by default for connections to Harbor web portal, and a secret resource named `harbor-tls-secret` is required for that purpose. It can be provisioned through cert-manager or manually with the following command using self-signed certificate:
      ```bash
      kubectl create secret tls harbor-tls-secret --cert=[path to certificate file] --key=[path to key file] -n eksa-packages
      ```
   * `secretKey` has to be set as a string of 16 characters for encryption.
   {{% /alert %}}

   TLS example with auto certificate generation
   ```yaml
   apiVersion: packages.eks.amazonaws.com/v1alpha1
   kind: Package
   metadata:
      name: my-harbor
      namespace: eksa-packages-<cluster-name>
   spec:
      packageName: harbor
      config: |-
         secretKey: "use-a-secret-key"
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
      namespace: eksa-packages-<cluster-name>
   spec:
      packageName: harbor
      config: |-
         secretKey: "use-a-secret-key"
         externalURL: http://harbor.eksa.demo:30002
         expose:
            tls:
               enabled: false
   ```

1. Install Harbor

   ```bash
   eksctl anywhere create packages -f harbor.yaml
   ```

1. Check Harbor

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAME        PACKAGE   AGE     STATE       CURRENTVERSION             TARGETVERSION        DETAIL
   my-harbor   harbor    5m34s   installed   v2.5.1                     v2.5.1 (latest)
   ```

   Harbor web portal is accessible at whatever `externalURL` is set to. See [complete configuration options]({{< relref "../harbor" >}}) for all default values.

   ![Harbor web portal](/images/harbor-portal.png)

## Update
To update package configuration, update harbor.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f harbor.yaml
```

## Upgrade
{{% alert title="Note" color="primary" %}}
* New versions of software packages will be automatically downloaded but not automatically installed. You can always manually run `eksctl` to check and install updates.
{{% /alert %}}
1. Verify a new bundle is available
   ```bash
   eksctl anywhere get packagebundle
   ```

   Example command output
   ```bash
   NAME         VERSION   STATE
   v1.25-120    1.25      active (upgrade available)
   v1.26-120    1.26      inactive
   ```

1. Upgrade Harbor
   ```bash
   eksctl anywhere upgrade packages --bundle-version v1.26-120
   ```

1. Check Harbor

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAME        PACKAGE   AGE     STATE       CURRENTVERSION             TARGETVERSION        DETAIL
   my-harbor   Harbor    14m     installed   v2.7.1                     v2.7.1 (latest)
   ```

## Uninstall
1. Uninstall Harbor

   {{% alert title="Important" color="warning" %}}

   * By default, PVCs created for jobservice and registry are not removed during a package delete operation, which can be changed by leaving `persistence.resourcePolicy` empty. 

   {{% /alert %}}
   ```bash
   eksctl anywhere delete package --cluster <cluster-name> my-harbor
   ```
