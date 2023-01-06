# Create new webhooks

## Install kubebuilder
```sh
make hack/tools/bin/kubebuilder
```

## Add API to PROJECT https://github.com/aws/eks-anywhere/blob/main/PROJECT
```sh
run ./hack/kubebuilder.sh create api --group anywhere --version v1alpha1 --kind WhateverKind
``` 

## Create new webhooks with kubebuilder
Since we use a non standard (according to kubebuilder) repo structure, kubebuilder commands won't work. For this purpose, we have a script that temporally changes our folder structure to one that kubebuilder understands and restores the original one after executing the kubebuilder command.

Example
```sh
./hack/kubebuilder.sh create webhook --group anywhere --version v1alpha1 --defaulting --programmatic-validation --kind WhateverKind
```
`--defaulting` creates mutation webhooks and `--programmatic-validation` creates validation webhooks.

Resources:

https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html

https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#request

## Change webhook marker name to under anywhere.amazonaws.com
The default name in webhook marker is `v<kind>.kb.io`, we need to change it to `<validation/mutation>.<kind>.anywhere.amazonaws.com`

## Generate manifests
```sh
run make release-manifests
```
