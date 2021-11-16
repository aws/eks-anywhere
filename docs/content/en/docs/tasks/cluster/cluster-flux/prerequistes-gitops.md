---
toc_hide: true
hide_summary: true
---

In order to use GitOps to manage cluster scaling, you need a couple of things:

- A GitHub account
- A cluster configuration file with a `GitOpsConfig`, referenced with a `gitOpsRef` in your Cluster spec
- A [Personal Access Token (PAT) for the GitHub account](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token), with permissions to create, clone, and push to a repo

### Create a GitHub Personal Access Token

[Create a Personal Access Token (PAT)](https://github.com/settings/tokens/new) to access your provided github repository.
It must be scoped for all `repo` permissions.

>**_NOTE:_** GitOps configuration only works with hosted github.com and will not work on self hosted GitHub Enterprise instances.

This PAT should have at least the following permissions:

![GitHub PAT permissions](/images/ss5.png)

>**_NOTE:_** The PAT must belong to the `owner` of the `repository` or, if using an organization as the `owner`, the creator of the `PAT` must have repo permission in that organization.

You need to set your PAT as the environment variable $EKSA_GITHUB_TOKEN to use it during cluster creation:

```
export EKSA_GITHUB_TOKEN=ghp_MyValidPersonalAccessTokenWithRepoPermissions
```

### Create GitOps configuration repo

If you have an existing repo you can set that as your repository name in the configuration.
If you specify a repo in your `GitOpsConfig` which does not exist EKS Anywhere will create it for you.
If you would like to create a new repo you can [click here](https://github.new) to create a new repo.

If your repository contains multiple cluster specification files, store them in subfolders and specify the [configuration path]({{< relref "../../../reference/clusterspec/gitops/#__clusterconfigpath__-optional" >}}) in your cluster specification.

In order to accommodate the management cluster feature, the CLI will now structure the repo directory following a new convention:

```
clusters
└── management-cluster
    ├── flux-system
    │   └── ...
    ├── management-cluster
    │   └── eksa-system
    │       └── eksa-cluster.yaml
    ├── workload-cluster-1
    │   └── eksa-system
    │       └── eksa-cluster.yaml
    └── workload-cluster-2
        └── eksa-system
            └── eksa-cluster.yaml
```
*By default, Flux kustomization reconciles at the management cluster's root level (`./clusters/management-cluster`), so both the management cluster and all of the workload clusters it manages are synced.*