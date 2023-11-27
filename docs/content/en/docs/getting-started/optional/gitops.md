---
title: "GitOps"
linkTitle: "GitOps"
weight: 50
aliases:
    /docs/reference/clusterspec/optional/gitops/
description: >
  Configuration reference for GitOps cluster management.
---

# GitOps Support (Optional)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

EKS Anywhere can create clusters that supports GitOps configuration management with Flux. 
In order to add GitOps support, you need to configure your cluster by specifying the configuration file with `gitOpsRef` field when creating or upgrading the cluster.
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

### Github provider
Please note that for the Flux config to work successfully with the Github provider, the environment variable `EKSA_GITHUB_TOKEN` needs to be set with a valid [GitHub PAT](https://github.com/settings/tokens/new).
This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
  namespace: default
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
  namespace: default
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

* __Description__: The name of the repository where EKS Anywhere will store your cluster configuration, and sync it to the cluster. If the repository exists, we will clone it from the git provider; if it does not exist, we will create it for you.
* __Type__: string

### __owner__ (required)

* __Description__: The owner of the Github repository; either a Github username or Github organization name. The Personal Access Token used must belong to the owner if this is a personal repository, or have permissions over the organization if this is not a personal repository.
* __Type__: string

### __personal__ (optional)

* __Description__: Is the repository a personal or organization repository?
  If personal, this value is `true`; otherwise, `false`.
  If using an organizational repository (e.g. `personal` is `false`) the `owner` field will be used as the `organization` when authenticating to github.com
* __Default__: true
* __Type__: boolean

### Git provider

Before you create a cluster using the Git provider, you will need to set and export the `EKSA_GIT_KNOWN_HOSTS` and `EKSA_GIT_PRIVATE_KEY` environment variables.

#### `EKSA_GIT_KNOWN_HOSTS`

EKS Anywhere uses the provided known hosts file to verify the identity of the git provider when connecting to it with SSH.
The `EKSA_GIT_KNOWN_HOSTS` environment variable should be a path to a known hosts file containing entries for the git server to which you'll be connecting.

For example, if you wanted to provide a known hosts file which allows you to connect to and verify the identity of github.com using a private key based on the key algorithm ecdsa, you can use the OpenSSH utility [ssh-keyscan](https://manpages.ubuntu.com/manpages/xenial/man1/ssh-keyscan.1.html) to obtain the known host entry used by github.com for the `ecdsa` key type.
EKS Anywhere supports `ecdsa`, `rsa`, and `ed25519` key types, which can be specified via the `sshKeyAlgorithm` field of the git provider config.

`ssh-keyscan -t ecdsa github.com >> my_eksa_known_hosts`

This will produce a file which contains known-hosts entries for the `ecdsa` key type supported by github.com, mapping the host to the key-type and public key.

`github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=`

EKS Anywhere will use the content of the file at the path `EKSA_GIT_KNOWN_HOSTS` to verify the identity of the remote git server, and the provided known hosts file must contain an entry for the remote host and key type.


#### `EKSA_GIT_PRIVATE_KEY`

The `EKSA_GIT_PRIVATE_KEY` environment variable should be a path to the private key file associated with a valid SSH public key registered with your Git provider.
This key must have permission to both read from and write to your repository.
The key can use the key algorithms `rsa`, `ecdsa`, and `ed25519`.

This key file must have restricted file permissions, allowing only the owner to read and write, such as octal permissions `600`.

If your private key file is passphrase protected, you must also set `EKSA_GIT_SSH_KEY_PASSPHRASE` with that value.

This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
  namespace: default
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
  namespace: default
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
* __Description__: The URL of an existing repository where EKS Anywhere will store your cluster configuration and sync it to the cluster. For private repositories, the SSH URL will be of the format `ssh://git@provider.com/$REPO_OWNER/$REPO_NAME.git`
* __Type__: string
* __Value__: A common `repositoryUrl` value can be of the format `ssh://git@provider.com/$REPO_OWNER/$REPO_NAME.git`. This may differ from the default SSH URL given by your provider. Consider these differences between github and CodeCommit URLs:
  * The github.com user interface provides an SSH URL containing a `:` before the repository owner, rather than a `/`. Make sure to replace this `:` with a `/`, if present. 
  * The CodeCommit SSH URL must include SSH-KEY-ID in format `ssh://<SSH-Key-ID>@git-codecommit.<region>.amazonaws.com/v1/repos/<repository>`.

### sshKeyAlgorithm (optional)

* __Description__: The SSH key algorithm of the private key specified via `EKSA_PRIVATE_KEY_FILE`. Defaults to `ecdsa`
* __Type__: string

Supported SSH key algorithm types are `ecdsa`, `rsa`, and `ed25519`.

Be sure that this SSH key algorithm matches the private key file provided by `EKSA_GIT_PRIVATE_KEY_FILE` and that the known hosts entry for the key type is present in `EKSA_GIT_KNOWN_HOSTS`.

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
  namespace: default
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
  namespace: default
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
* __Description__: The name of the repository where EKS Anywhere will store your cluster configuration, and sync it to the cluster.
  If the repository exists, we will clone it from the git provider; if it does not exist, we will create it for you.
* __Type__: string

#### __owner__ (required)
* __Description__: The owner of the Github repository; either a Github username or Github organization name.
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
