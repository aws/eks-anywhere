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

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

TYPE="$1"
ENV="$2"

RELEASES_YAML="https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"
if [ "$ENV" = "development" ]; then
    RELEASES_YAML="https://beta-assets.eks-anywhere.model-rocket.aws.dev/releases/eks-a/manifest.yaml"
fi

BUNDLE_NUMBER="$(cat $REPO_ROOT/release/triggers/bundle-release/$ENV/BUNDLE_NUMBER)"
RELEASE_NUMBER="$(cat $REPO_ROOT/release/triggers/eks-a-release/$ENV/RELEASE_NUMBER)"

CLI_MAX_VERSION="$(cat $REPO_ROOT/release/triggers/bundle-release/$ENV/CLI_MAX_VERSION)"
RELEASE_VERSION="$(cat $REPO_ROOT/release/triggers/eks-a-release/$ENV/RELEASE_VERSION)"
if [ "$TYPE" = "bundle" ]; then
    if curl --retry 5 -s $RELEASES_YAML | yq  -e ".spec.releases[] | select(.number == \"$BUNDLE_NUMBER\") | .number" &> /dev/null; then
        echo "$BUNDLE_NUMBER already exists in $RELEASES_YAML!"
        echo "Double check the bundle numbers in other release branches"
        exit 1
    fi

    PREVIOUS_BUNDLE_NUMBER=$((BUNDLE_NUMBER-1))
    if ! curl --retry 5 -s $RELEASES_YAML | yq  -e ".spec.releases[] | select(.number == \"$PREVIOUS_BUNDLE_NUMBER\") | .number" &> /dev/null; then
        echo "Previous number ($PREVIOUS_BUNDLE_NUMBER) does not exist in $RELEASES_YAML!"
        echo "Double check that your new bundle number is +1 the previous release"
        exit 1
    fi


    CLI_MIN_VERSION="$(cat $REPO_ROOT/release/triggers/bundle-release/$ENV/CLI_MIN_VERSION)"
    if [ "$CLI_MAX_VERSION" != "$CLI_MIN_VERSION" ]; then
        echo "CLI_MAX_VERSION ($CLI_MAX_VERSION) should match CLI_MIN_VERSION ($CLI_MIN_VERSION)"
        exit 1
    fi

    if curl --retry 5 -s $RELEASES_YAML | yq  -e ".spec.releases[] | select(.version == \"$CLI_MAX_VERSION\") | .version" &> /dev/null; then
        echo "$CLI_MAX_VERSION already exists in $RELEASES_YAML!"
        echo "Double check the CLI_MAX_VERSION and CLI_MIN_VERSION numbers"
        exit 1
    fi

    if [ "$ENV" = "production" ]; then
        DEV_BUNDLE_NUMBER="$(cat $REPO_ROOT/release/triggers/bundle-release/development/BUNDLE_NUMBER)"
        if [ "$BUNDLE_NUMBER" != "$DEV_BUNDLE_NUMBER" ]; then
            echo "Production number ($BUNDLE_NUMBER) should equal development number ($DEV_BUNDLE_NUMBER)"
            exit 1
        fi

        DEV_CLI_MAX_VERSION="$(cat $REPO_ROOT/release/triggers/bundle-release/development/CLI_MAX_VERSION)"
        if [ "$CLI_MAX_VERSION" != "$DEV_CLI_MAX_VERSION" ]; then
            echo "Production CLI_MAX_VERSION ($CLI_MAX_VERSION) should equal development CLI_MAX_VERSION ($CLI_MAX_VERSION)"
            exit 1
        fi

        DEV_CLI_MIN_VERSION="$(cat $REPO_ROOT/release/triggers/bundle-release/development/CLI_MIN_VERSION)"
        if [ "$CLI_MIN_VERSION" != "$DEV_CLI_MIN_VERSION" ]; then
            echo "Production CLI_MIN_VERSION ($CLI_MIN_VERSION) should equal development CLI_MIN_VERSION ($DEV_CLI_MIN_VERSION)"
            exit 1
        fi
    fi
elif [ "$TYPE" = "eks-a" ]; then
    if [ "$BUNDLE_NUMBER" != "$RELEASE_NUMBER" ]; then
        echo "RELEASE_NUMBER ($RELEASE_NUMBER) should equal bundle number ($BUNDLE_NUMBER)"
        exit 1
    fi
    if [ "$CLI_MAX_VERSION" != "$RELEASE_VERSION" ]; then
        echo "RELEASE_VERSION ($RELEASE_VERSION) should equal CLI_MAX_VERSION ($CLI_MAX_VERSION)"
        exit 1
    fi
fi
