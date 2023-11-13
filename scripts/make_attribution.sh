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

set -e
set -x
set -o errexit
set -o pipefail

GOLANG_VERSION="$1"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../" && pwd -P)"
if [ "$CI" = "true" ]; then
    # Prow clone process does not add remote so we have to explicitly add
    # it. Setting it to HTTPS URL for go-licenses 
    git remote add origin https://github.com/$REPO_OWNER/eks-anywhere.git
else
    # go-licenses supports only HTTPS URLs so if the CLI repo is cloned locally 
    # with SSH URL, we need to temporarily override the remote URL for go-licenses
    # to work as expected
    ORIGIN_URL=$(git remote get-url origin)
    if [[ "$ORIGIN_URL" =~ "git@github.com" ]]; then
        CLONE_METHOD="ssh"
        REPO_OWNER="$(dirname $ORIGIN_URL | cut -d':' -f2)"
        git remote set-url origin https://github.com/${REPO_OWNER}/eks-anywhere.git
    else
        CLONE_METHOD="https"
    fi
fi

source "$REPO_ROOT/scripts/common.sh"
source "$REPO_ROOT/scripts/attribution_helpers.sh"

function build::attribution::generate(){
    cd $REPO_ROOT
    $(build::common::get_go_path "$GOLANG_VERSION")/go mod vendor
    build::create_git_tag
    build::gather_licenses _output "./cmd/eksctl-anywhere ./controllers"
    build::exclude_own
    build::generate_attribution $GOLANG_VERSION
    # Removing temporary override by resetting remote origin URL to original SSH url 
    if [ "$CI" != "true" ] && [ "$CLONE_METHOD" = "ssh" ]; then
        git remote set-url origin $ORIGIN_URL
    elif [ "$CI" = "true" ] && [ "$JOB_TYPE" = "periodic" ]; then
        git remote remove origin
    fi
    rm -rf _output bin vendor GIT_TAG
}

build::attribution::generate
