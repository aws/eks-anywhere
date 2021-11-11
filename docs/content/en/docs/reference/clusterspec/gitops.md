---
title: "GitOpsConfig configuration"
linkTitle: "GitOps"
weight: 80
description: >
  Configuration reference for GitOps cluster management.
---

# GitOps Support (Optional)
EKS Anywhere can create clusters that supports GitOps configuration management with Flux. 
In order to add GitOps support, you need to configure your cluster by updating the configuration file before creating the cluster. 
Please note that for the GitOps config to work successfully the environment variable `EKSA_GITHUB_TOKEN` needs to be set with a valid [GitHub PAT](https://github.com/settings/tokens/new). This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  ...
  #GitOps Support
  gitOpsRef:
    name: my-gitops
    kind: GitOpsConfig
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: GitOpsConfig
metadata:
  name: my-gitops
spec:
  flux:
    github:
      personal: true
      repository: myClusterGitopsRepo
      owner: myGithubUsername
      fluxSystemNamespace: ""
      clusterConfigPath: ""
```

### GitOps Configuration Spec Details
### __flux__ (required)
* __Description__: our supported gitops provider is `flux`.
  This is the only supported value.
* __Type__: object

### Flux Configuration Spec Details
### __github__ (required)
* __Description__: `github` is the only currently supported git provider.
  This defines your github configuration to be used by EKS Anywhere and flux.
* __Type__: object

### github Configuration Spec Details
#### __repository__ (required)
* __Description__: The name of the repository where we will store your cluster configuration, and sync it to the cluster.
  If the repository exists, we will clone it from the git provider; if it does not exist, we will create it for you.
* __Type__: string

#### __owner__ (required)
* __Description__: The owner of the git repository; either a github username or github organization name.
  The Personal Access Token used must belong to the `owner` if this is a `personal` repository, or have permissions over the organization if this is not a `personal` repository.
* __Type__: string

#### __personal__ (optional)
* __Description__: Is the repository a personal or organization repository?
  If personal, this value is `true`; otherwise, `false`.
  If using an organizational repository (e.g. `personal` is `false`) the `owner` field will be used as the `organization` when authenticating to github.com
* __Default__: `true`
* __Type__: boolean

#### __clusterConfigPath__ (optional)
* __Description__: The path relative to the root of the git repository where EKS Anywhere will store the cluster configuration files.
* __Default__: `clusters/$MANAGEMENT_CLUSTER_NAME`
* __Type__: string

#### __fluxSystemNamespace__ (optional)
* __Description__: Namespace in which to install the gitops components in your cluster.
* __Default__: `flux-system`.
* __Type__: string

#### __branch__ (optional)
* __Description__: The branch to use when commiting the configuration.
* __Default__: `main`
* __Type__: string
