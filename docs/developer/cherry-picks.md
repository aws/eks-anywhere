# Release Branches

eks-anywhere maintains multiple release branches to represent the currently supported versions.  

In general, bug fixes should be backported to currently supported release branches, new features should not.

## Currently supported release branches

- [release-0.6](https://github.com/aws/eks-anywhere/tree/release-0.6)

# "Automated" Cherry Picks

The [cherry_pick_pull.sh](../../hack/cherry_pick_pull.sh) is provided 
to assist in backporting.  The script is the same used by upstream [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes/blob/master/hack/cherry_pick_pull.sh)

## To create a Cherry Pick

- Open + Merge PR in eks-anywhere repo
- Run `GITHUB_USER=<github_user> ./build/lib/cherry_pick_pull.sh upstream/<release-branch> <pr number>`
	- The script assumes your remotes are setup such that `upstream` points the `aws/eks-anywhere` and `origin`
	points to your fork at `<github_user>/eks-anywhere-build-tooling`
	- If your remotes are not setup this way you can set `UPSTREAM_REMOTE=<upstream remote name>` `FORK_REMOTE=<fork remote name>`
	when calling the script to override the defaults
	- There a couple of pre-reqs, having a `GITHUB_TOKEN` set and the `gh` cli installed.  The script will let you know if you are missing any of these
- If there is a merge conflict, the script will wait and give you a chance to fix conflicts in another terminal before continuing
- The script will push a new branch and open a PR automatically
- Run the above script for each currently supported release branch

Refer to the upstream [doc](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-release/cherry-picks.md) for more information.
