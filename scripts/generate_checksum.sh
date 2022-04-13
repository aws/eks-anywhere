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
SHASUM_FILES="shasums-eksctl-anywhere"

#Clean up any existing old folders for shasum files
if [ -d "$SHASUM_FILES" ]; then rm -Rf $SHASUM_FILES; fi

YQ_LATEST_RELEASE_URL="https://github.com/mikefarah/yq/releases/latest"

if ! command -v yq &> /dev/null
then
    echo "Please install YQ from $YQ_LATEST_RELEASE_URL and rerun this script."
    exit
fi

echo "YQ is installed, continuing.."

FILE_PREFIX="eksctl-anywhere"

TAR="amd64.tar.gz"


CURRENT_YQ_VERSION=$(yq -V | awk '{print $NF}')

versionSplit=($(echo $CURRENT_YQ_VERSION | tr "." "\n"))
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

mkdir -p $SHASUM_FILES
cd $SHASUM_FILES
curl --silent 'https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml' -o release.yaml

latest_release_version=$(cat release.yaml | yq e '.spec.latestVersion')

latest_release=$(cat release.yaml | yq e '.spec.releases[] | select(.version == "'$latest_release_version'")' >> latest_release.yaml)

for sha in 256 512
do
  for os in linux darwin
  do
    SHA_VALUE=$(cat latest_release.yaml | yq e ".eksABinary.${os}.sha${sha}")
    echo "${SHA_VALUE}  ${FILE_PREFIX}-${latest_release_version}-${os}-${TAR}" >> "${FILE_PREFIX}-${latest_release_version}-sha${sha}sums.txt"
  done
done

echo "Generated the check sum files inside the ${SHASUM_FILES} folder at the directory where you ran this script from."