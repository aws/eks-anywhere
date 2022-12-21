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
CLUSTER_CONTROLLER_OVERRIDE_IMAGE="$2"

SED=sed
if [ "$(uname -s)" = "Darwin" ]; then
    SED=gsed
fi

if ! yq --version; then
    echo "Cannot find yq executable. Build on Linux"
    echo "or install yq on Mac (brew install yq)."
    exit 1
fi

mkdir -p $REPO_ROOT/bin
make release-manifests RELEASE_DIR=$REPO_ROOT/bin RELEASE_MANIFEST_TARGET=local-eksa-components.yaml
curl $BUNDLE_MANIFEST_URL -o $REPO_ROOT/bin/local-bundle-release.yaml
$SED -i "s,public.ecr.aws/.*/eks-anywhere-cluster-controller:.*,${CLUSTER_CONTROLLER_OVERRIDE_IMAGE}," $REPO_ROOT/bin/local-eksa-components.yaml
yq e -i '.spec.versionsBundles[].eksa.components.uri |= "bin/local-eksa-components.yaml"' $REPO_ROOT/bin/local-bundle-release.yaml
