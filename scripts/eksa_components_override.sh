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
set -o pipefail
set -x
set -o nounset

REPO_ROOT=$(git rev-parse --show-toplevel)
BUNDLE_MANIFEST_URL="$1"
CI="${CI:-false}"
CODEBUILD_CI="${CODEBUILD_CI:-false}"

SED=sed
if [ "$(uname -s)" = "Darwin" ]; then
    SED=gsed
fi

if ! yq --version; then
    echo "Cannot find yq executable. Build on Linux"
    echo "or install yq on Mac (brew install yq)."
    exit 1
fi

# In Prow, we can get the base commit and PR commit using these
# environment variables and check their diff to check if files under
# config folder changed.
if [ "$CI" = "true" ]; then
    CONFIG_FILES_CHANGED="$(git --no-pager diff --pretty=format:"" --name-only $PULL_BASE_SHA $PULL_PULL_SHA | grep "^config/.*" || true)"
# In Codebuild, we can do a git show on the latest commit to check
# if files under config folder changed.
elif [ "$CODEBUILD_CI" = "true" ]; then
    CONFIG_FILES_CHANGED="$(git --no-pager show --pretty=format:"" --name-only HEAD | grep "^config/.*" || true)"
else
# For local testing, we can either do a git diff or git show on HEAD
# based on whether the local changes have been committed
    while true; do
        read -p "Have you committed your local changes? (yes/no) " response
        case $response in
            [Yy]* ) CONFIG_FILES_CHANGED="$(git --no-pager show --pretty=format:"" --name-only HEAD | grep "^config/.*" || true)"; break;;
            [Nn]* ) CONFIG_FILES_CHANGED="$(git --no-pager diff --pretty=format:"" --name-only HEAD | grep "^config/.*" || true)"; break;;
            * ) echo "Please answer yes or no.";;
        esac
    done
fi

if [[ "$CONFIG_FILES_CHANGED" != "" ]]; then
    mkdir -p $REPO_ROOT/bin
    make release-manifests RELEASE_DIR=$REPO_ROOT/bin RELEASE_MANIFEST_TARGET=local-eksa-components.yaml
    curl $BUNDLE_MANIFEST_URL -o $REPO_ROOT/bin/local-bundle-release.yaml
    CLUSTER_CONTROLLER_OVERRIDE_IMAGE=$(yq e ".spec.versionsBundles[0].eksa.clusterController.uri" $REPO_ROOT/bin/local-bundle-release.yaml)
    $SED -i "s,public.ecr.aws/.*/eks-anywhere-cluster-controller:.*,${CLUSTER_CONTROLLER_OVERRIDE_IMAGE}," $REPO_ROOT/bin/local-eksa-components.yaml
    yq e -i '.spec.versionsBundles[].eksa.components.uri |= "bin/local-eksa-components.yaml"' $REPO_ROOT/bin/local-bundle-release.yaml
fi