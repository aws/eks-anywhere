---
title: "Registry Mirror"
linkTitle: "Registry Mirror"
weight: 40
aliases:
    /docs/reference/clusterspec/optional/registrymirror/
description: >
  EKS Anywhere cluster yaml specification for registry mirror configuration
---

## Registry Mirror Support (optional)
You can configure EKS Anywhere to use a private registry as a mirror for pulling the required images.

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
## Registry Mirror Configuration Spec Details
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
* __Description__: when you need to mirror multiple registries, you can map each upstream registry to the "namespace" of its mirror.
* __Type__: array
* __Example__: <br/>
  ```yaml
  ociNamespaces:
    - registry: "public.ecr.aws"
      namespace: "eks-anywhere"
    - registry: "783794618700.dkr.ecr.us-west-2.amazonaws.com"
      namespace: "curated-packages"
  ```

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

## Import images into a private registry
You can use the `download images` and `import images` commands to pull images from `public.ecr.aws` and push them to your
private registry.
The `copy packages` must be used if you want to copy EKS Anywhere Curated Packages to your registry mirror.
The `download images` command also pulls the Cilium chart from `public.ecr.aws` and pushes it to the registry mirror. It requires the registry credentials for performing a login. Set the following environment variables for the login:
```bash
export REGISTRY_USERNAME=<username>
export REGISTRY_PASSWORD=<password>
```

{{% alert title="Warning" color="warning" %}}
`eksctl anywhere download images` and `eksctl anywhere import images` command need to be run on an amd64 machine to import amd64 images to the registry mirror.
{{% /alert %}}

Download the EKS Anywhere artifacts to get the EKS Anywhere bundle:
```bash
eksctl anywhere download artifacts
tar -xzf eks-anywhere-downloads.tar.gz
```

Download and import EKS Anywhere images:
```bash
REGISTRY_ENDPOINT=<registry_endpoint>
eksctl anywhere download images -o eks-anywhere-images.tar
docker login https://${REGISTRY_ENDPOINT}
...
eksctl anywhere import images -i eks-anywhere-images.tar --bundles eks-anywhere-downloads/bundle-release.yaml --registry ${REGISTRY_ENDPOINT}
```

Use the EKS Anywhere bundle to copy packages:
```bash
eksctl anywhere copy packages --bundle ./eks-anywhere-downloads/bundle-release.yaml --dst-cert rootCA.pem ${REGISTRY_ENDPOINT}
```

## Docker configurations
It is necessary to add the private registry's CA Certificate
to the list of CA certificates on the admin machine if your registry uses self-signed certificates.

For [Linux](https://docs.docker.com/engine/security/certificates/), you can place your certificate here: `/etc/docker/certs.d/<private-registry-endpoint>/ca.crt`

For [Mac](https://docs.docker.com/desktop/mac/#add-tls-certificates), you can follow this guide to add the certificate to your keychain: https://docs.docker.com/desktop/mac/#add-tls-certificates

{{% alert title="Note" color="primary" %}}
  You may need to restart Docker after adding the certificates.
{{% /alert %}}
