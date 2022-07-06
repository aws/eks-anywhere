# eksa controller

## Install kubebuilder
```sh
make hack/tools/bin/kubebuilder
```

## Create new webhooks with kubebuilder
Since we use a non standard (according to kubebuilder) repo structure, kubebuilder commands won't work. For this purpose, we have a script that temporally changes our folder structure to one that kubebuilder understands and restores the original one after executing the kubebuilder command.

Example
```sh
./hack/kubebuilder.sh create webhook --group anywhere --version v1alpha1 --programmatic-validation --kind WhateverKind
```

## Run controller from local repo source with tilt
When using tilt, any changes to the yaml files in `config` or `go` code in `pkg/api` and `controllers` will automatically rebuild and update your resources in the cluster.

Note: the folder `config/tilt` is ignored. This folder is supposed to contain tilt exclusive kustomize files and is not intended for manual changes (in order to keep the tilt environment as close as possible to the real one, its patches should be minimum). If you make changes to this folder you will need to restart tilt.
### Setup tilt config
Create a `tilt-settings.json` file in this folder
```json
{
  "default_registry": "public.ecr.aws/xxxxxx",
  "allowed_contexts": ["yyyyyy@zzzzz"]
}
```
* `default_registry`: your own registry where you want to push the controller images built by tilt. If using ECR, you will need to create the repository in advance (repo name is `cluster-controller`, same as the var `IMG` in the Tiltfile). You will need to be authenticated and have permissions to push images. Example for ECR:
```sh
aws ecr-public get-login-password --region ${REGION} | docker login --username AWS --password-stdin public.ecr.aws/${REGISTRY_ALIAS}
```
* `allowed_contexts`: list here the kube context of your cluster. By default, tilt won't interact with "non local" clusters and any eksa cluster, including the docker ones, are recognized as non local
### Point tilt to your cluster
Tilt uses whatever cluster `kubectl` is configured to use. The easiest option here is to set `KUBECONFIG` envar pointing to your eksa kubeconfig file:

```sh
export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

### Start tilt
```sh
make run-controller
```