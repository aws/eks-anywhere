#!/usr/bin/env bash

# Echo the names of go packages modified in the differences between the
# working tree and the AWS main branch in github.

# This is useful for limiting testing runs.

set -euo pipefail

function awsRemote () {
    git remote -v \
	| grep "github.com[:/]aws/eks-anywhere.git" \
	| head -n 1 \
	| awk '{print $1}' \
	| tr -d '\n'
}

ref="$(awsRemote)/main"

{
    set +e
    git diff --stat --name-only "$ref...HEAD"
    git status --untracked-files=normal --short | awk '{ print $NF }'
} | grep .\.go\$ \
    | xargs -r -n1 dirname \
    | sort \
    | uniq \
    | sed -e 's,^,./,' \
    | tr '\n' ' '

echo

