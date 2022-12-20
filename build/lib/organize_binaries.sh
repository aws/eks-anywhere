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

BINARY_DEPS_DIR="$1"
ARCH="$2"

OUTPUT_DIR="${BINARY_DEPS_DIR}/linux-$ARCH"
EKS_A_TOOL_BINARY_DIR="${OUTPUT_DIR}/eks-a-tools/binary"
EKS_A_TOOL_LICENSE_DIR="${OUTPUT_DIR}/eks-a-tools/licenses"

mkdir -p $EKS_A_TOOL_BINARY_DIR
mkdir -p $EKS_A_TOOL_LICENSE_DIR

declare -A project_bin_licenses=(["eksa/vmware/govmomi"]="govc GOVMOMI"
                                 ["eksa/helm/helm"]="helm HELM"
                                 ["eksa/apache/cloudstack-cloudmonkey"]="cmk CMK")

for project in "${!project_bin_licenses[@]}"
do
  bin_license_map=(${project_bin_licenses[$project]})
  binary=${bin_license_map[0]}
  license_prefix=${bin_license_map[1]}
  cp $OUTPUT_DIR/$project/$binary $EKS_A_TOOL_BINARY_DIR/$(basename $binary)
  cp $OUTPUT_DIR/$project/ATTRIBUTION.txt $EKS_A_TOOL_LICENSE_DIR/${license_prefix}_ATTRIBUTION.txt
  cp -r $OUTPUT_DIR/$project/LICENSES $EKS_A_TOOL_LICENSE_DIR/${license_prefix}_LICENSES
done
