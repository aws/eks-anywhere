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

# This is a diff impl than what it is eks-anywhere-build-tooling
# we can rely on the git_tag file existing in this repo
# instead we pull it from eks-anywhere-build-tooling git
function build::common::get_latest_eksa_asset_url() {
  local -r artifact_bucket=$1
  local -r project=$2
  local -r arch=${3-amd64}
  local -r latesttag=${4-latest}

  local branch=main
  if [[ "$latesttag" != "latest" ]]; then
    branch=$latesttag
  fi

  local -r git_tag=$(curl -sL https:/raw.githubusercontent.com/aws/eks-anywhere-build-tooling/$branch/projects/$project/GIT_TAG)
  local -r url="https://$(basename $artifact_bucket).s3-us-west-2.amazonaws.com/projects/$project/$latesttag/$(basename $project)-linux-$arch-${git_tag}.tar.gz"

  local -r http_code=$(curl -I -L -s -o /dev/null -w "%{http_code}" $url)
  if [[ "$http_code" == "200" ]]; then 
    echo "$url"
  else
    echo "https://$(basename $artifact_bucket).s3-us-west-2.amazonaws.com/projects/$project/latest/$(basename $project)-linux-$arch-${git_tag}.tar.gz"
  fi
}
