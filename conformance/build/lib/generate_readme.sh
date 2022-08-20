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

set -o errexit
set -o nounset
set -o pipefail

EKSA_K8S_VERSION="${1?First argument is EKS-A k8s version}"
CONFORMANCE_VERSION="${2?Second argument is conformance test version}"
TARGET_DIR="${3?Third argument is target directory}"

TARGET=${TARGET_DIR}/README.md

cat template/README.md.begin >${TARGET}
sed -e 's/^/   /' -e "s/kubernetesVersion: .*$/kubernetesVersion: \"${EKSA_K8S_VERSION}\"/" <template/example.yaml >>${TARGET}
cat template/README.md.end >>${TARGET}
sed -i "" -e "s/{{eksa_k8s_version}}/${EKSA_K8S_VERSION}/" -e "s/{{conformance_version}}/${CONFORMANCE_VERSION}/" ${TARGET_DIR}/README.md
