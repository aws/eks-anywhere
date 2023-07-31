---
title: "Credential Provider Package"
linkTitle: "Add Credential Provider Package"
weight: 13
date: 2023-03-31
description: >
  Install/upgrade/uninstall Credential Provider Package
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

{{% alert title="Important" color="warning" %}}
* Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
* While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
  {{% /alert %}}

## Install
By default an instance of this package is installed with the controller to help facilitate authentication for other packages. The following are instructions in case you want to tweak the default values.

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package credential-provider-package --cluster <cluster-name> > credential-provider-package.yaml
   ```
1. Add the desired configuration to `credential-provider-package.yaml`
    Please see [complete configuration options]({{< relref "../credential-provider-package" >}}) for all configuration options and their default values.
    Example default package using IAM User Credentials installed with the controller
    ```
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-credential-provider-package
      namespace: eksa-packages-<clusterName>
      annotations:
        "helm.sh/resource-policy": keep
        "anywhere.eks.aws.com/internal": "true"
    spec:
      packageName: credential-provider-package
      targetNamespace: eksa-packages
      config: |-
        tolerations:
          - key: "node-role.kubernetes.io/master"
            operator: "Exists"
            effect: "NoSchedule"
          - key: "node-role.kubernetes.io/control-plane"
            operator: "Exists"
            effect: "NoSchedule"
        sourceRegistry: public.ecr.aws/eks-anywhere
        credential:
          - matchImages:
            - 783794618700.dkr.ecr.us-west-2.amazonaws.com
            profile: "default"
            secretName: aws-secret
            defaultCacheDuration: "5h"
    ```

1. Create the secret. If you are changing the secret, see [complete configuration options]({{< relref "../credential-provider-package" >}}) for the format of the secret.

1. Create the namespace (if not installing to eksa-packages).
   If you are overriding `targetNamespace`, change `eksa-packages` to the value of `targetNamespace`.
   ```bash
   kubectl create namespace <namespace-name-here>
   ```

1. Install the credential-provider-package
   ```bash
   eksctl anywhere create packages -f credential-provider-package.yaml
   ```
   
1. Validate the installation
   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```
   
## Update
To update package configuration, update credential-provider-package.yaml file and run the following command:
```bash
eksctl anywhere apply package -f credential-provider-package.yaml
```

## Upgrade

Credential-Provider-Package will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall credential-provider-package, simply delete the package:

```bash
eksctl anywhere delete package --cluster <cluster-name> my-credential-provider-package
```
