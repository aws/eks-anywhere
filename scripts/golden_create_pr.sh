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
REPO="eks-anywhere"
ORIGIN_ORG="eks-distro-pr-bot"
UPSTREAM_ORG="aws"

PR_TITLE="Generate release testdata files"
COMMIT_MESSAGE="[PR BOT] Generate release testdata files"

PR_BODY=$(cat <<EOF
Generate release testdata files.

By submitting this pull request, I confirm that you can use, modify, copy, and redistribute this contribution, under the terms of your choice.
EOF
)

PR_BRANCH="update-release-test-file"

cd ${SCRIPT_ROOT}/..
git config --global push.default current
git config user.name "EKS Distro PR Bot"
git config user.email "aws-model-rocket-bots+eksdistroprbot@amazon.com"
git remote add origin git@github.com:${ORIGIN_ORG}/${REPO}.git
git remote add upstream git@github.com:${UPSTREAM_ORG}/${REPO}.git
git checkout -b $PR_BRANCH

git diff
git add release/cli/pkg/test/testdata/*.yaml
# If some other files get modified, the changes should be ignored
git restore .
FILES_ADDED=$(git diff --staged --name-only)
if [ "$FILES_ADDED" = "" ]; then
    exit 0
fi

git commit -m "$COMMIT_MESSAGE"
ssh-agent bash -c 'ssh-add /secrets/ssh-secrets/ssh-privatekey; ssh -o StrictHostKeyChecking=no git@github.com; git fetch upstream; git rebase -Xtheirs upstream/main; git push -u origin $PR_BRANCH -f'

gh auth login --with-token < /secrets/github-secrets/token

PR_EXISTS=$(GH_PAGER="" gh pr list --json number -H "${PR_BRANCH}")
if [ "$PR_EXISTS" = "[]" ]; then
  gh pr create --title "$PR_TITLE" --body "$PR_BODY"
fi
