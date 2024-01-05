---
title: "Registry Mirror"
linkTitle: "Registry Mirror"
weight: 40
aliases:
    /docs/reference/clusterspec/optional/registrymirror/
description: >
  EKS Anywhere cluster specification for registry mirror configuration
---

## Registry Mirror Support (optional)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

You can configure EKS Anywhere to use a local registry mirror for its dependencies. When a registry mirror is configured in the EKS Anywhere cluster specification, EKS Anywhere will use it instead of defaulting to Amazon ECR for its dependencies. For details on how to configure your local registry mirror for EKS Anywhere, see the [Configure local registry mirror]({{< relref "./registrymirror/#configure-local-registry-mirror" >}}) section.

See the [airgapped documentation page]({{<relref "../airgapped" >}}) for instructions on downloading and importing EKS Anywhere dependencies to a local registry mirror.

## Registry Mirror Cluster Spec

The following cluster spec shows an example of how to configure registry mirror:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  ...
  registryMirrorConfiguration:
    endpoint: <private registry IP or hostname>
    port: <private registry port>
    ociNamespaces:
      - registry: <upstream registry IP or hostname>
        namespace: <namespace in private registry>
      ...
    caCertContent: |
      -----BEGIN CERTIFICATE-----
      MIIF1DCCA...
      ...
      es6RXmsCj...
      -----END CERTIFICATE-----  
```
## Registry Mirror Cluster Spec Details
### __registryMirrorConfiguration__ (optional)
* __Description__: top level key; required to use a private registry.
* __Type__: object

### __endpoint__ (required)
* __Description__: IP address or hostname of the private registry for pulling images
* __Type__: string
* __Example__: ```endpoint: 192.168.0.1```

### __port__ (optional)
* __Description__: port for the private registry. This is an optional field. If a port
  is not specified, the default HTTPS port `443` is used
* __Type__: string
* __Example__: ```port: 443```

### __ociNamespaces__ (optional)
* __Description__: when you need to mirror multiple registries, you can map each upstream registry to the "namespace" of its mirror. The namespace is appended with the endpoint, `<endpoint>/<namespace>` to setup the mirror for the registry specified.
Note while using `ociNamespaces`, you need to specify __all__ the registries that need to be mirrored. This includes an entry for the `public.ecr.aws` registry to pull EKS Anywhere images from.

* __Type__: array
* __Example__: <br/>
  ```yaml
  ociNamespaces:
    - registry: "public.ecr.aws"
      namespace: ""
    - registry: "783794618700.dkr.ecr.us-west-2.amazonaws.com"
      namespace: "curated-packages"
  ```
{{% alert title="Warning" color="warning" %}}
Currently only `public.ecr.aws` registry is supported for mirroring with Bottlerocket OS.
{{% /alert %}}



### __caCertContent__ (optional)
* __Description__: certificate Authority (CA) Certificate for the private registry . When using 
  self-signed certificates it is necessary to pass this parameter in the cluster spec. This __must__ be the individual public CA cert used to sign the registry certificate. This will be added to the cluster nodes so that they are able to pull images from the private registry.

  It is also possible to configure CACertContent by exporting an environment variable:<br/>
  `export EKSA_REGISTRY_MIRROR_CA="/path/to/certificate-file"`
* __Type__: string
* __Example__: <br/>
  ```yaml
  CACertContent: |
    -----BEGIN CERTIFICATE-----
    MIIF1DCCA...
    ...
    es6RXmsCj...
    -----END CERTIFICATE-----
  ```

### __authenticate__ (optional)

* __Description__: optional field to authenticate with a private registry. When using private registries that 
  require authentication, it is necessary to set this parameter to ```true``` in the cluster spec.
* __Type__: boolean

When this value is set to true, the following environment variables need to be set:
```bash
export REGISTRY_USERNAME=<username>
export REGISTRY_PASSWORD=<password>
```

### __insecureSkipVerify__ (optional)
* __Description__: optional field to skip the registry certificate verification. Only use this solution for isolated testing or in a tightly controlled, air-gapped environment. Currently only supported for Ubuntu and RHEL OS.
* __Type__: boolean

## Configure local registry mirror

### Project configuration
The following projects must be created in your registry before importing the EKS Anywhere images:

```
bottlerocket
eks-anywhere
eks-distro
isovalent
cilium-chart
```

For example, if a registry is available at `private-registry.local`, then the following projects must be created.

```
https://private-registry.local/bottlerocket
https://private-registry.local/eks-anywhere
https://private-registry.local/eks-distro
https://private-registry.local/isovalent
https://private-registry.local/cilium-chart
```

### Admin machine configuration
You must configure the Admin machine with the information it needs to communicate with your registry.

Add the registry's CA certificate to the list of CA certificates on the Admin machine if your registry uses self-signed certificates.

- For [Linux](https://docs.docker.com/engine/security/certificates/), you can place your certificate here: `/etc/docker/certs.d/<private-registry-endpoint>/ca.crt`
- For [Mac](https://docs.docker.com/desktop/mac/#add-tls-certificates), you can follow this guide to add the certificate to your keychain: https://docs.docker.com/desktop/mac/#add-tls-certificates

If your registry uses authentication, the following environment variables must be set on the Admin machine before running the `eksctl anywhere import images` command.
```bash
export REGISTRY_USERNAME=<username>
export REGISTRY_PASSWORD=<password>
```


