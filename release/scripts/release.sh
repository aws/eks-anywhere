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
set -o pipefail

export LANG=C.UTF-8

BASE_DIRECTORY=$(git rev-parse --show-toplevel)

source "$BASE_DIRECTORY"/scripts/eksa_version.sh

ARTIFACTS_DIR="${1?Specify first argument - artifacts path}"
SOURCE_BUCKET="${2?Specify second argument - source bucket}"
RELEASE_BUCKET="${3?Specify third argument - release bucket}"
CDN="${4?Specify fourth argument - cdn}"
SOURCE_CONTAINER_REGISTRY="${5?Specify fifth argument - source container registry}"
RELEASE_CONTAINER_REGISTRY="${6?Specify sixth argument - release container registry}"
BUILD_REPO_URL="${7?Specify seventh argument - Build repo URL}"
CLI_REPO_URL="${8?Specify eighth argument - CLI repo URL}"
BUILD_REPO_BRANCH_NAME="${9?Specify ninth argument - Build repo branch name}"
CLI_REPO_BRANCH_NAME="${10?Specify tenth argument - CLI repo branch name}"
DRY_RUN="${11?Specify eleventh argument - Dry run}"
WEEKLY="${12?Specify twelfth argument - Weekly release}"

mkdir -p "${ARTIFACTS_DIR}"

release_version=$(eksa-version::get_next_eksa_version_for_ancestor "$CLI_REPO_BRANCH_NAME")

${BASE_DIRECTORY}/release/bin/eks-anywhere-release release \
    --release-version "${release_version}" \
    --artifact-dir "${ARTIFACTS_DIR}" \
    --build-repo-url "${BUILD_REPO_URL}" \
    --cli-repo-url "${CLI_REPO_URL}" \
    --build-repo-branch-name "${BUILD_REPO_BRANCH_NAME}" \
    --cli-repo-branch-name "${CLI_REPO_BRANCH_NAME}" \
    --source-bucket "${SOURCE_BUCKET}" \
    --source-container-registry "${SOURCE_CONTAINER_REGISTRY}" \
    --cdn "${CDN}" \
    --release-bucket "${RELEASE_BUCKET}" \
    --release-container-registry "${RELEASE_CONTAINER_REGISTRY}" \
    --dev-release=true \
    --dry-run=${DRY_RUN} \
    --weekly=${WEEKLY} \
    --aws-signer-profile-arn "${AWS_SIGNER_PROFILE_ARN}"
