#!/usr/bin/env bash
# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# get_closest_ancestor_branch returns the branch that is the closest ancestor of the current branch.
# only main and release-* branches are considered as ancestors.
function eksa-version::get_closest_ancestor_branch() {
    # Get the name of the current branch
    local current_branch=$(git rev-parse --abbrev-ref HEAD)

    # Initialize variables to keep track of the closest branch and its merge-base
    local closest_branch=""
    local closest_date=""

    # List all branches except the current one, and iterate through them
    for branch in $(git branch --all | grep upstream | sed 's/\*//g' | sed 's/remotes\/upstream\///g' | sed 's/^[ \t]*//' | grep -e '^main$' -e '^release-' | sort -u); do
        # Avoid comparing with HEAD or detached states
        if [[ "$branch" == "HEAD" ]]; then
            continue
        fi

        # When running the script locally, filter remotes to get the remote corresponding to the primary repo
        # aws/eks-anywhere, otherwise the local fork's default origin remote will be used, which could potentially
        # be out of date and produce incorrect versions
        if ! [[ $CI == true || $CODEBUILD_CI == true ]]; then
            upstream_remote=$(git remote -v | grep "aws/eks-anywhere.git" | awk '{print $1}' | uniq)
            branch="$upstream_remote/$branch"
        fi
        # Find the common ancestor of the current branch and the iterated branch
        local merge_base=$(git merge-base "$current_branch" "$branch" 2>/dev/null)

        # Get the commit date of the common ancestor
        local merge_base_date=$(git show -s --format=%ci "$merge_base" 2>/dev/null)

        # Initialize closest_date if it's empty
        if [[ -z "$closest_date" ]]; then
            closest_date="$merge_base_date"
            closest_branch="$branch"
        else
            # Compare dates to find the most recent common ancestor
            if [[ "$merge_base_date" > "$closest_date" ]]; then
                closest_date="$merge_base_date"
                closest_branch="$branch"
            fi
        fi
    done

    echo "$closest_branch"

}

function eksa-version::get_next_eksa_version_for_ancestor() {
    local ancestor_branch=$1
    local release_version
    local latest_tag
    # This checks if main or upstream/main are the ancestors of the current branch.
    # If it is, then it is either main or a branch off main.
    # If it is not, then it is a release branch or a branch off a release branch.
    if [[ "$ancestor_branch" =~ "main" ]]; then
        #  If the branch is main, then get the latest release tag from the releases manifest and
        # bump one minor version and using 0 as the patch

        latest_tag=$(curl https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.latestVersion")

        release_version=$(echo "${latest_tag}" | awk -F. -v OFS=. '{$2++; $3=0; print}')
    else
        # For a release branch, get the latest tag for the release minor version and bump the patch version
        # If there is not tag yet, use latest tag by date but bumping the minor version and using 0 as the patch
        # Silence stderr as the command will fail if there are no tags
        minor_release=$(echo $ancestor_branch | grep -Eo "[[:digit:]]+.[[:digit:]]+")
        latest_tag=$(git tag -l "v${minor_release}.*" | sort -V | tail -n 1)
        if [[ -z "$latest_tag" ]]; then
            latest_tag=$(git describe --tags "$(git tag -l "v*" --sort -v:refname | head -1)")
            release_version=$(echo "${latest_tag}" | awk -F. -v OFS=. '{$2++; $3=0; print}')
        else
            release_version=$(echo "${latest_tag}" | awk -F. -v OFS=. '{$3++; print}')
        fi
    fi

    echo "$release_version"
}

function eksa-version::get_next_eksa_version() {
    local ancestor_branch
    ancestor_branch=$(eksa-version::get_closest_ancestor_branch)
    eksa-version::get_next_eksa_version_for_ancestor "$ancestor_branch"
}

function eksa-version::latest_release_version_in_manifest() {
    local manifest_file=$1
    local release_version
    release_version=$(curl -s -L "$manifest_file" | yq '.spec.latestVersion')
    echo "$release_version"
}
