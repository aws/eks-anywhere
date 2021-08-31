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
#
# To deploy to a development environment make sure you have
# ${AWS_PROFILE} set to your account
# ${AMPLIFY_APP_ID} (optional) set to your Amplify app or a stack named eks-anywhere-docs-devstack
# ${AMPLIFY_APP_BRANCH} (optional) set to Amplify branch you want to deploy to [default: main]

set -euo pipefail

# if $AWS_PROFILE is not set we should assume we're in a deployment pipeline
if [ -z "${AWS_PROFILE}" ]; then

	# set up AWS credentials file
	cat << EOF > config
[default]
output = json
region = us-west-2
role_arn=$AWS_ROLE_ARN
web_identity_token_file=/var/run/secrets/eks.amazonaws.com/serviceaccount/token
[profile release-prod]
role_arn = arn:aws:iam::153288728732:role/DocsDeploymentRole
region = us-east-1
source_profile=default
EOF

	AWS_CONFIG_FILE=$(pwd)/config
	AWS_DEFAULT_PROFILE=release-prod
fi

# Use this function if the script is cancelled (ctrl+c) or times out
cancel_job() {
	aws amplify stop-job --app-id $APP --branch $BRANCH --job-id $JOB_ID
}

# vars for deployment
GIT_COMMIT=$(git rev-parse --short HEAD)
FILE=${1:-public.zip}
APP=${AMPLIFY_APP_ID:-$(aws cloudformation describe-stacks \
    --stack-name eks-anywhere-docs-devstack \
    --query 'Stacks[].Outputs[?OutputKey==`EksAnywhereDocsappId`].OutputValue' \
    --output text)}
BRANCH=${AMPLIFY_APP_BRANCH:-main}

# Check if we're on a tag commit
GIT_TAG=$(git tag --points-at ${GIT_COMMIT})

# If we're on a git tag we should create a branch so documentation versions work
# Branches are configured in config.toml
if [ -n "${GIT_TAG}" ]; then
	# replaces tag text like v0.2.0 with v0-2
	BRANCH="$(echo ${GIT_TAG%.*} | tr '.' '-')"
	# create an amplify app branch for the tag
	aws amplify create-branch --app-id $APP --branch-name $BRANCH --no-cli-pager
fi

# Create a deployment and save the job id and upload url
read -r JOB_ID URL < <(aws amplify create-deployment --app-id  $APP --branch $BRANCH --output text)

# cancel the amplify job if we exit or are inturrupted
trap cancel_job INT EXIT
# upload our zip
curl --silent --show-error --fail --upload-file $FILE "${URL}"

aws amplify start-deployment --app-id $APP --branch $BRANCH --job-id $JOB_ID

# get job status and wait for it to finish
STATUS=$(aws amplify get-job --app-id $APP --branch $BRANCH --job-id $JOB_ID --output text --query 'job.summary.status')
# Don't trap for exits after this point
trap - EXIT
# Total timeout will be 20+19+18...
TIMEOUT=20
while [ $TIMEOUT -gt 0 ]; do
if [ "${STATUS}" = 'SUCCEED' ]; then
	echo "Successfully deployed $BRANCH"
	exit 0
elif [ "${STATUS}" = 'FAILED' ]; then
	echo "Applify job $JOB_ID failed"
	aws amplify get-job --app-id $APP --branch $BRANCH --job-id $JOB_ID
	exit 1
else
	echo "Waiting for amplify job $JOB_ID on app $APP"
	STATUS=$(aws amplify get-job --app-id $APP --branch $BRANCH --job-id $JOB_ID --output text --query 'job.summary.status')
fi
sleep $(( TIMEOUT-- ))
done
# If we timeout waiting for the job this will run
cancel_job
