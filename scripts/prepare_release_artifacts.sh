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

YQ_LATEST_RELEASE_URL="https://github.com/mikefarah/yq/releases/latest"

if ! command -v yq &> /dev/null
then
    echo "Please install YQ  from $YQ_LATEST_RELEASE_URL and rerun this script."
    exit
fi

CURRENT_YQ_VERSION=$(yq -V | awk '{print $NF}')

versionSplit=($(echo $CURRENT_YQ_VERSION | tr -d "v" | tr "." "\n"))
CURRENT_YQ_MAJOR_VERSION=${versionSplit[0]}
CURRENT_YQ_MINOR_VERSION=${versionSplit[1]}
EXPECTED_YQ_MAJOR_VERSION=4
EXPECTED_YQ_MINOR_VERSION=18

if [ $CURRENT_YQ_MAJOR_VERSION -lt $EXPECTED_YQ_MAJOR_VERSION ]; then
    echo "ERROR: Current yq major version v$CURRENT_YQ_MAJOR_VERSION.x is older than expected v$EXPECTED_YQ_MAJOR_VERSION.x"
    echo "Please install the latest version of yq from $YQ_LATEST_RELEASE_URL"
    exit 1
fi

if [ $CURRENT_YQ_MAJOR_VERSION == $EXPECTED_YQ_MAJOR_VERSION ] && [ $CURRENT_YQ_MINOR_VERSION -lt $EXPECTED_YQ_MINOR_VERSION ]; then
    echo "ERROR: Current yq minor version v$CURRENT_YQ_MAJOR_VERSION.$CURRENT_YQ_MINOR_VERSION.x is older than expected v$EXPECTED_YQ_MAJOR_VERSION.$EXPECTED_YQ_MINOR_VERSION.x"
    echo "Please install the latest version of yq from $YQ_LATEST_RELEASE_URL"
    exit 1
fi

OUT_DIR="release-artifacts"

# Clean up any existing old folders
if [ -d "$OUT_DIR" ]; then rm -Rf $OUT_DIR; fi

mkdir -p $OUT_DIR
( 
cd $OUT_DIR

release_yaml="$(curl --silent 'https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml')"
latest_release_version=$(echo "$release_yaml" | yq '.spec.latestVersion')
latest_release=$(echo "$release_yaml" | yq '.spec.releases[] | select(.version == "'$latest_release_version'")')


echo "Building checksum files..."
for platform in amd64 arm64; do
    for os in linux darwin; do
        checksum256=$(echo "$latest_release" | yq ".eksACLI.${os}.${platform}.sha256")
        checksum512=$(echo "$latest_release" | yq ".eksACLI.${os}.${platform}.sha512")
        uri=$(echo "$latest_release" | yq ".eksACLI.${os}.${platform}.uri")
        file=${uri##*/}
        echo "${checksum256} ${file}" >> "eksctl-anywhere-${latest_release_version}-sha256sums.txt"
        echo "${checksum512} ${file}" >> "eksctl-anywhere-${latest_release_version}-sha512sums.txt"
    done
done

release_urls=($(echo "$release_yaml" | yq e '.spec.releases[] | select(.version == "'$latest_release_version'") | .eksACLI.*.*.uri'))

echo "Downloading release binaries..."
for uri in ${release_urls[@]}; do
    curl -fSsLO "$uri"
done;

echo "Release artifacts can be found in ${OUT_DIR}."
)