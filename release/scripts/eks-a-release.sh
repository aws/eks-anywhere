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

RELEASE_VERSION="${1?Specify first argument - latest version}"
ARTIFACTS_DIR="${2?Specify second argument - artifacts path}"
SOURCE_BUCKET="${3?Specify third argument - source bucket}"
RELEASE_BUCKET="${4?Specify fourth argument - release bucket}"
CDN="${5?Specify fifth argument - cdn}"
BUNDLE_NUMBER="${6?Specify sixth argument - Bundle number}"
RELEASE_NUMBER="${7?Specify seventh argument - Release number}"
RELEASE_ENVIRONMENT="${8?Specify eighth argument - Release environment}"
CLI_REPO_BRANCH_NAME="${9?Specify ninth argument - Branch name}"
BUILD_REPO_URL="${10?Specify tenth argument - Build repo URL}"
CLI_REPO_URL="${11?Specify eleventh argument - CLI repo URL}"

set_aws_config "$RELEASE_ENVIRONMENT" "cli"

mkdir -p "${ARTIFACTS_DIR}"

${BASE_DIRECTORY}/release/bin/eks-anywhere-release release \
    --release-version "${RELEASE_VERSION}" \
    --bundle-number "${BUNDLE_NUMBER}" \
    --release-number "${RELEASE_NUMBER}" \
    --cli-repo-branch-name "${CLI_REPO_BRANCH_NAME}" \
    --artifact-dir "${ARTIFACTS_DIR}" \
    --source-bucket "${SOURCE_BUCKET}" \
    --cdn "${CDN}" \
    --release-bucket "${RELEASE_BUCKET}" \
    --release-environment ${RELEASE_ENVIRONMENT} \
    --dev-release=false \
    --bundle-release=false \
    --build-repo-url "${BUILD_REPO_URL}" \
    --cli-repo-url "${CLI_REPO_URL}"
