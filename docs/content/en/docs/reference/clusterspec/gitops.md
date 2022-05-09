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
We currently support two types of configurations: `FluxConfig` and `GitOpsConfig`.

## Flux Configuration
The flux configuration spec has three optional fields, regardless of the chosen git provider.

### Flux Configuration Spec Details
### __systemNamespace__ (optional)
* __Description__: Namespace in which to install the gitops components in your cluster. Defaults to `flux-system`
* __Type__: string

### __clusterConfigPath__ (optional)

* __Description__: The path relative to the root of the git repository where EKS Anywhere will store the cluster configuration files. Defaults to the cluster name
* __Type__: string

### __branch__ (optional)

* __Description__: The branch to use when committing the configuration. Defaults to `main`
* __Type__: string

EKS Anywhere currently supports two git providers for FluxConfig: Github and Git.

## Github provider
Please note that for the Flux config to work successfully with the Github provider, the environment variable `EKSA_GITHUB_TOKEN` needs to be set with a valid [GitHub PAT](https://github.com/settings/tokens/new).
This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  ...
  #GitOps Support
  gitOpsRef:
    name: my-github-flux-provider
    kind: FluxConfig
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: FluxConfig
metadata:
  name: my-github-flux-provider
spec:
  systemNamespace: "my-alternative-flux-system-namespace"
  clusterConfigPath: "path-to-my-clusters-config"
  branch: "main"
  github:
    personal: true
    repository: myClusterGitopsRepo
    owner: myGithubUsername

---
```

### github Configuration Spec Details
### __repository__ (required)

* __Description__: The name of the repository where we will store your cluster configuration, and sync it to the cluster. If the repository exists, we will clone it from the git provider; if it does not exist, we will create it for you.
* __Type__: string

### __owner__ (required)

* __Description__: The owner of the git repository; either a github username or github organization name. The Personal Access Token used must belong to the owner if this is a personal repository, or have permissions over the organization if this is not a personal repository.
* __Type__: string

### __personal__ (optional)

* __Description__: Is the repository a personal or organization repository? If personal, this value is true; otherwise, false. If using an organizational repository (e.g. personal is false) the owner field will be used as the organization when authenticating to github.com (http://github.com/)
* __Default__: true
* __Type__: boolean

## Git provider

Before you create a cluster using the Git provider, you will need to set and export these environment variables.

`EKSA_GIT_KNOWN_HOSTS`: Path to your known hosts file

*-** TODO: insert info on how to obtain known hosts info for the git provider*

`EKSA_GIT_PRIVATE_KEY`: Path to your private key file associated with a valid SSH public key registered in your Git setup.

If your private key file is password protected, you must also set `EKSA_GIT_SSH_KEY_PASSPHRASE` with that value.

This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  ...
  #GitOps Support
  gitOpsRef:
    name: my-git-flux-provider
    kind: FluxConfig
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: FluxConfig
metadata:
  name: my-git-flux-provider
spec:
  systemNamespace: "my-alternative-flux-system-namespace"
  clusterConfigPath: "path-to-my-clusters-config"
  branch: "main"
  git:
    repositoryUrl: ssh://git@github.com/myAccount/myClusterGitopsRepo.git
    sshKeyAlgorithm: ecdsa
---
```

### git Configuration Spec Details
### repositoryUrl (required)

* __Description__: The URL of an existing repository where we will store your cluster configuration and sync it to the cluster.
* __Type__: string

### sshKeyAlgorithm (optional)

* __Description__: The SSH public key algorithm for the private key specified. Defaults to `ecdsa`
* __Type__: string

## GitOps Configuration

{{% alert title="Warning" color="warning" %}}
GitOps Config will be deprecated in v0.11.0 in lieu of using the Flux Config described above. 
{{% /alert %}}

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
* __Description__: The branch to use when committing the configuration.
* __Default__: `main`
* __Type__: string

