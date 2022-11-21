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

TARGET_DIR="${1?First argument is target directory}"

RESULTS_DIR=$(mktemp -d results.XXXXXX)
RESULTS=$(./sonobuoy retrieve ${RESULTS_DIR})
tar xzf $RESULTS -C ${RESULTS_DIR}
mv ${RESULTS_DIR}/plugins/e2e/results/global/* ${TARGET_DIR}
./sonobuoy results ${RESULTS}
rm -rf ${RESULTS_DIR}
