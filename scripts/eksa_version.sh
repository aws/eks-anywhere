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
    # This checks if main in the ancestors of the current branch.
    # If it is, then it is either main or a branch off main.
    # If it is not, then it is a release branch or a branch off a release branch.
    if [[ "$ancestor_branch" == "main" ]]; then
        #  If the branch is main, then get the latest tag by date and
        # bump one minor version and use patch 0
        
        latest_tag=$(git describe --tags "$(git rev-list --tags --max-count=1)")
        
        release_version=$(echo "${latest_tag}" | awk -F. -v OFS=. '{$2++; $3=0; print}')
    else
        # For release branhc, get the latest tag that matches current commit
        # and bump the patch version
        latest_tag=$(git describe --tags --abbrev=0)
        release_version=$(echo "${latest_tag}" | awk -F. -v OFS=. '{$3++; print}')
    fi

    echo "$release_version"
}

function eksa-version::get_next_eksa_version() {
    local ancestor_branch
    ancestor_branch=$(eksa-version::get_closest_ancestor_branch)
    eksa-version::get_next_eksa_version_for_ancestor "$ancestor_branch"
}

function eksa-version::latest_release_verison_in_manifest() {
    local manifest_file=$1
    local release_version
    release_version=$(curl -s -L "$manifest_file" | yq '.spec.latestVersion')
    echo "$release_version"
}