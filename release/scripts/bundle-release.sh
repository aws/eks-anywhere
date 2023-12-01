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

BASE_DIRECTORY=$(git rev-parse --show-toplevel)
SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
source ${SCRIPT_ROOT}/setup-aws-config.sh

ARTIFACTS_DIR="${1?Specify first argument - artifacts path}"
SOURCE_BUCKET="${2?Specify second argument - source bucket}"
RELEASE_BUCKET="${3?Specify third argument - release bucket}"
CDN="${4?Specify fourth argument - cdn}"
BUNDLE_NUMBER="${5?Specify fifth argument - Bundle number}"
CLI_MIN_VERSION="${6?Specify sixth argument - CLI min version}"
CLI_MAX_VERSION="${7?Specify seventh argument - CLI max version}"
SOURCE_CONTAINER_REGISTRY="${8?Specify eighth argument - source container registry}"
RELEASE_CONTAINER_REGISTRY="${9?Specify ninth argument - release container registry}"
RELEASE_ENVIRONMENT="${10?Specify tenth argument - Release environment}"
BUILD_REPO_BRANCH_NAME="${11?Specify eleventh argument - Build repo branch name}"
CLI_REPO_BRANCH_NAME="${12?Specify twelfth argument - CLI repo branch name}"
BUILD_REPO_URL="${13?Specify thirteenth argument - Build repo URL}"
CLI_REPO_URL="${14?Specify fourteenth argument - CLI repo URL}"

set_aws_config "$RELEASE_ENVIRONMENT"

mkdir -p "${ARTIFACTS_DIR}"

${BASE_DIRECTORY}/release/bin/eks-anywhere-release release \
    --bundle-number "${BUNDLE_NUMBER}" \
    --min-version "${CLI_MIN_VERSION}" \
    --max-version "${CLI_MAX_VERSION}" \
    --build-repo-branch-name "${BUILD_REPO_BRANCH_NAME}" \
    --cli-repo-branch-name "${CLI_REPO_BRANCH_NAME}" \
    --artifact-dir "${ARTIFACTS_DIR}" \
    --source-bucket "${SOURCE_BUCKET}" \
    --source-container-registry "${SOURCE_CONTAINER_REGISTRY}" \
    --cdn "${CDN}" \
    --release-bucket "${RELEASE_BUCKET}" \
    --release-container-registry "${RELEASE_CONTAINER_REGISTRY}" \
    --release-environment ${RELEASE_ENVIRONMENT} \
    --dev-release=false \
    --bundle-release=true \
    --build-repo-url "${BUILD_REPO_URL}" \
    --cli-repo-url "${CLI_REPO_URL}" \
    --aws-signer-profile-arn "${AWS_SIGNER_PROFILE_ARN}"
