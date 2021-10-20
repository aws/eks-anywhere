---
title: "Private Registry configuration"
linkTitle: "Private Registry"
weight: 90
description: >
  EKS Anywhere cluster yaml specification private registry configuration reference
---

## Private Repository Support (optional)
You can configure EKS Anywhere to use a private registry. The following cluster spec 
shows an example of how to configure private registries:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
  registryMirrorConfiguration:
    endpoint: <private registry IP/hostname address>
    caCertContent: |
      <CA certificate>
```
## Private Registry Configuration Spec Details
### __registryMirrorConfiguration__ (required)
* __Description__: top level key; required to use a private registry.
* __Type__: object

### __endpoint__ (required)
* __Description__: IP address or hostname of the private registry to be used when pulling images during cluster creation 
* __Type__: string
* __Example__: ```endpoint: 192.168.0.1```
### __caCertContent__ (optional)
* __Description__: Certificate Authority's Certificate used for TLS. When using 
  self-signed certificates it is necessary to pass this parameter in the cluster spec.
  It is also possible to configure CACertContent by exporting an environment variable:
  `export EKSA_REGISTRY_MIRROR_CA="/path/to/certificate-file"`
* __Type__: string
* __Example__: <br/>
  CACertContent: |<br/>
  &nbsp;&nbsp;&nbsp;&nbsp;-----BEGIN CERTIFICATE-----<br/>
  &nbsp;&nbsp;&nbsp;&nbsp;AbCdEfG........................................<br/>
  &nbsp;&nbsp;&nbsp;&nbsp;-----END CERTIFICATE-----<br/>

## Import images into a private registry
Use the import-images command to pull images from the default registry (public.ecr.aws) and push them to your
private registry.

```bash
docker login https://<private registry endpoint>
...
eksctl anywhere import-images -f cluster-spec.yaml
```
## Docker configurations
It will be necessary to add the Certificate Authority's Certificate
to the list of CA certificates on the admin machine if your registry uses self-signed certificates.

## Registry configurations
Depending on what registry you decide to use, creating the following projects 
 will be necessary:

```
bottlerocket
eks-anywhere
eks-distro
isovalent
```

For example, if a registry is available at `private-registry.local`, then the following 
projects will have to be created:

```
https://private-registry.local/bottlerocket
https://private-registry.local/eks-anywhere
https://private-registry.local/eks-distro
https://private-registry.local/isovalent
```
