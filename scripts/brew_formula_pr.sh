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

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="homebrew-tap"
ORIGIN_ORG="eks-anywhere-brew-pr-bot"
UPSTREAM_ORG="aws"

PR_BODY=$(cat <<EOF
Issue #, if available:

Description of changes:
Update EKS Anywhere formula to point at our latest release.

By submitting this pull request, I confirm that you can use, modify, copy, and redistribute this contribution, under the terms of your choice.
EOF
)

cd ${SCRIPT_ROOT}/
cd ../../
gh auth login --with-token < /secrets/github-secrets/token

cd ${SCRIPT_ROOT}
BREW_UPDATE_SCRIPT="brew_formula_update.sh"
if [ ! -f "$BREW_UPDATE_SCRIPT" ]
then
  echo "The script to update the brew formula does not exist in this folder, exiting.."
  exit 1
fi

LATEST_VERSION=$(echo $(/home/prow/go/src/github.com/aws/eks-anywhere/scripts/$BREW_UPDATE_SCRIPT))

cd ${SCRIPT_ROOT}/../../../${ORIGIN_ORG}/${REPO}
git config --global push.default current
git config user.name "EKS Anywhere Brew Update PR Bot"
git config user.email "aws-model-rocket-bots+eksbrewprbot@amazon.com"
git remote -v

git remote add origin git@github.com:${ORIGIN_ORG}/${REPO}.git
git remote add upstream git@github.com:${UPSTREAM_ORG}/${REPO}.git

PR_TITLE="chore: update eks-anywhere formula for ${LATEST_VERSION}"
COMMIT_MESSAGE="[PR BOT] Update eks-anywhere brew formula for ${LATEST_VERSION}"
PR_BRANCH="eks-anywhere-formula-update"

git checkout -b $PR_BRANCH

git diff
git add Formula/eks-anywhere.rb
# If some other files get modified, the changes should be ignored
git restore .
FILES_ADDED=$(git diff --staged --name-only)
if [ "$FILES_ADDED" = "" ]; then
    exit 0
fi

git commit -m "$COMMIT_MESSAGE"
ssh-agent bash -c 'ssh-add /secrets/ssh-secrets/ssh-privatekey; ssh -o StrictHostKeyChecking=no git@github.com; git fetch upstream; git rebase -Xtheirs upstream/master; git push -u origin $PR_BRANCH -f'

echo "Added ssh private key\n"

PR_EXISTS=$(GH_PAGER="" gh pr list --json number -H "${PR_BRANCH}")
if [ "$PR_EXISTS" = "[]" ]; then
  gh pr create --title "$PR_TITLE" --body "$PR_BODY" --repo "${UPSTREAM_ORG}/${REPO}"
fi

