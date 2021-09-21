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

BINARY_NAME="${1?Specify first argument - binary name}"
TAG="${2?Specify second argument - git version tag}"
TAR_PATH="${3?Specify third argument - tarball output path}"
BIN_ROOT="_output/bin"
LICENSES_PATH="_output/LICENSES"
MANIFESTS_PATH="_output/manifests"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "$REPO_ROOT/scripts/common.sh"

function build::eks-anywhere-cluster-controller::create_tarball() {
  platform=${1}
  OS="$(cut -d '/' -f1 <<< ${platform})"
  ARCH="$(cut -d '/' -f2 <<< ${platform})"
  TAR_FILE="${BINARY_NAME}-${OS}-${ARCH}-${TAG}.tar.gz"

  cp -rf $LICENSES_PATH $BIN_ROOT/$BINARY_NAME/${OS}-${ARCH}/
  cp ATTRIBUTION.txt $BIN_ROOT/$BINARY_NAME/${OS}-${ARCH}/
  build::common::create_tarball ${TAR_PATH}/${TAR_FILE} $BIN_ROOT/$BINARY_NAME/${OS}-${ARCH} .
  cp -rf $MANIFESTS_PATH/ ${TAR_PATH}
}

function build::eks-anywhere-cluster-controller::tarball() {
  build::common::ensure_tar
  mkdir -p "$TAR_PATH"
  build::eks-anywhere-cluster-controller::create_tarball "linux/amd64"
  rm -rf $BIN_ROOT
  rm -rf $LICENSES_PATH
}

build::eks-anywhere-cluster-controller::tarball

build::common::generate_shasum "${TAR_PATH}"
