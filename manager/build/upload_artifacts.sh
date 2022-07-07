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

set -x
set -o errexit
set -o nounset
set -o pipefail

SRC_TAR_PATH="${1?Specify first argument - source tar path}"
ARTIFACTS_BUCKET="${2?Specify second argument - artifacts buckets}"
PROJECT_PATH="${3? Specify third argument - project path}"
BUILD_IDENTIFIER="${4? Specify fourth argument - build identifier}"
GIT_HASH="${5?Specify fifth argument - git hash of the tar builds}"
LATEST_PATH="${6?Specify sixth argument - latest S3 folder path for artifacts}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "$REPO_ROOT/scripts/common.sh"

build::common::upload_artifacts $SRC_TAR_PATH $ARTIFACTS_BUCKET $PROJECT_PATH $BUILD_IDENTIFIER $GIT_HASH $LATEST_PATH
