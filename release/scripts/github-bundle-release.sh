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

ARTIFACTS_DIR="${1?Specify first argument - artifacts path}"
BUNDLE_MANIFEST_URL="${2?Specify second argument - weekly bundle manifest URL}"

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
RELEASE_NOTES_PATH="$SCRIPT_ROOT/github-bundle-release-notes"
export DATE_YYYYMMDD=$(date "+%F")
RELEASE_TAG="weekly.$DATE_YYYYMMDD"

# Filling in values for the GitHub Release notes template
envsubst '$DATE_YYYYMMDD:$BUILD_REPO_HEAD:$CLI_REPO_HEAD' \
    < "$RELEASE_NOTES_PATH.tmpl" \
    > "$RELEASE_NOTES_PATH"

# Downloading the weekly bundle release manifest
mkdir -p $ARTIFACTS_DIR
wget $BUNDLE_MANIFEST_URL -O $ARTIFACTS_DIR/$DATE_YYYYMMDD-bundle-release.yaml

# Publish the asset as a Github pre-release on main branch with a new dated tag
gh release create $RELEASE_TAG $ARTIFACTS_DIR/$DATE_YYYYMMDD-bundle-release.yaml --notes-file "$RELEASE_NOTES_PATH" --prerelease --repo "github.com/aws/eks-anywhere" --title "Weekly Release $DATE_YYYYMMDD" --target "main"
