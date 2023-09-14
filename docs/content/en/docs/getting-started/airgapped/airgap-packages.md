---
toc_hide: true
---
If you are running in an airgapped environment and you set up a local registry mirror, you can copy curated packages from Amazon ECR to your local registry mirror with the following command. 

The `$BUNDLE_RELEASE_YAML_PATH` should be set to the `eks-anywhere-downloads/bundle-release.yaml` location where you unpacked the tarball from the`eksctl anywhere download artifacts` command. The `$REGISTRY_MIRROR_CERT_PATH` and `$REGISTRY_MIRROR_URL` values must be the same as the `registryMirrorConfiguration` in your EKS Anywhere cluster specification.

```bash
eksctl anywhere copy packages \
  --bundle ${BUNDLE_RELEASE_YAML_PATH} \
  --dst-cert ${REGISTRY_MIRROR_CERT_PATH} \
  ${REGISTRY_MIRROR_URL}
```