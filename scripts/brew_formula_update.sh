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
SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="homebrew-tap"
ORIGIN_ORG="eks-anywhere-brew-pr-bot"
UPSTREAM_ORG="aws"
EKS_ANYWHERE_RELEASES_MANIFEST_URL="https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml"
YQ_LATEST_RELEASE_URL="https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64"
source "$SCRIPT_ROOT/common.sh"

if ! command -v yq &> /dev/null
then
  wget -qO /usr/local/bin/yq $YQ_LATEST_RELEASE_URL
  chmod a+x /usr/local/bin/yq
fi

curl --silent $EKS_ANYWHERE_RELEASES_MANIFEST_URL -o release.yaml

BRANCH_NAME=$PULL_BASE_REF
ALL_RELEASE_BRANCHES=($(build::common::get_release_branches))
LATEST_RELEASE_BRANCHES=($(get_latest_release_branches "${ALL_RELEASE_BRANCHES[@]}"))

# Get the latest release version corresponding to this release branch
if [[ "$BRANCH_NAME" == "${LATEST_RELEASE_BRANCHES[0]}" ]]; then
  latest_release_version=$(cat release.yaml | yq e '.spec.latestVersion')
elif [[ "$BRANCH_NAME" == "${LATEST_RELEASE_BRANCHES[1]}" ]]; then
  latest_release_version=$(build::common::get_latest_release_for_release_branch $BRANCH_NAME)
else
  echo "Unsupported EKS Anywhere release branch $BRANCH_NAME!"
  exit 1
fi

latest_release=$(cat release.yaml | yq e '.spec.releases[] | select(.version == "'$latest_release_version'")' >> latest_release.yaml)

for os in darwin linux
do
  for arch in arm64 amd64
  do
    KEY_URL="${os}_${arch}_url"
    KEY_SHA="${os}_${arch}_sha256"
    export $KEY_URL=$(cat latest_release.yaml | yq e ".eksACLI.${os}.${arch}.uri")
    export $KEY_SHA=$(cat latest_release.yaml | yq e ".eksACLI.${os}.${arch}.sha256")
  done
done

# removes the char v from string like => v1.0.1
latest_release_version="${latest_release_version:1}"
export VERSION=$latest_release_version

EKSA_TEMPLATE="eks-anywhere.rb.tmpl"
if [[ "$BRANCH_NAME" == "${LATEST_RELEASE_BRANCHES[0]}" ]]; then
  EKSA_FORMULA="${SCRIPT_ROOT}/../../../${ORIGIN_ORG}/${REPO}/Formula/eks-anywhere.rb"
  export VERSION_SUFFIX=""
elif [[ "$BRANCH_NAME" == "${LATEST_RELEASE_BRANCHES[1]}" ]]; then
  minor_version="${BRANCH_NAME##release-}"
  EKSA_FORMULA="${SCRIPT_ROOT}/../../../${ORIGIN_ORG}/${REPO}/Formula/eks-anywhere@$minor_version.rb"
  export VERSION_SUFFIX="AT${minor_version/.}"
fi

if [ ! -f "$EKSA_TEMPLATE" ]
then
  echo "Template file ${EKSA_TEMPLATE} does not exist in this working folder, exiting.."
  exit 1
fi

envsubst '$VERSION:$VERSION_SUFFIX:$darwin_arm64_url:$darwin_arm64_sha256:$darwin_amd64_url:$darwin_amd64_sha256:$linux_arm64_url:$linux_arm64_sha256:$linux_amd64_url:$linux_amd64_sha256' \
 < "${EKSA_TEMPLATE}" \
 > "${EKSA_FORMULA}"

echo "v${VERSION}"

rm release.yaml
rm latest_release.yaml
