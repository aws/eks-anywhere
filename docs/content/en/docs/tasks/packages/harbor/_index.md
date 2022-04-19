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
   * Harbor web portal is exposed through `NodePort` by default and its default port number is `30003`
   * TLS is enabled by default for connections to Harbor web portal, and a secret resource named `tls-secret` is required for that purpose. It can be created with the following command:
      ```bash
      kubectl create secret tls tls-secret --cert=[path to your certificate file] --key=[path to your key file] -n eksa-packages
      ```
   * The `UpdateStrategy` for deployments with persistent volumes (jobservice, registry and chartmuseum) has to be set to `Recreate` when `ReadWriteMany` for volumes isn't supported in the cluster. 

      `RollingUpdate` strategy works with `ReadWriteMany` volumes only.

   {{% /alert %}}

   ```yaml
   apiVersion: packages.eks.amazonaws.com/v1alpha1
   kind: Package
   metadata:
      ...
      name: my-harbor
      namespace: eksa-packages
      ...
   spec:
      packageName: Harbor
      targetNamespace: eksa-packages
      config: |-
         externalURL: https://harbor.eksa.demo:30003
         expose:
            type: nodePort
            tls:
               enabled: true
               secret:
                  secretName: "tls-secret"
         UpdateStrategy:
            type: Recreate
         persistence:
            persistentVolumeClaim:
               jobservice:
                  accessMode: ReadWriteOnce
               registry:
                  accessMode: ReadWriteOnce
         ...
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
1. Generate a new package bundle
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: packages.eks.amazonaws.com/v1alpha1
   kind: PackageBundle
   metadata:
      name: v1.21-1001
      namespace: eksa-packages
   annotations:
      eksa.aws.com/excludes: LnNwZWMucGFja2FnZXNbXS5zb3VyY2UucmVnaXN0cnkKLnNwZWMucGFja2FnZXNbXS5zb3VyY2UucmVwb3NpdG9yeQo=
      eksa.aws.com/signature: MEUCIQD/PkoLGRI12jO8B8Y/m7spwNojs6AWXMLiLreoutpWvgIgBYYLYkiXKHfMWuICEKj6ERZceM6Lin3VgPeyYLvv3BI=
   spec:
      packages:
         - name: harbor
            source:
               registry: public.ecr.aws
               repository: harbor/harbor-helm
               versions:
                  - name: v2.4.2
                    digest: sha256:53f0c5dcd47c27072c027fd4d94a6658208378a233cb5a528a9454ad6a5e4eb8
      kubeVersion: "1.21"
   EOF
   ```

1. Check the new bundle
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
