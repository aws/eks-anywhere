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

export LANG=C.UTF-8

BASE_DIRECTORY=$(git rev-parse --show-toplevel)

ARTIFACTS_DIR="${1?Specify first argument - artifacts path}"
SOURCE_BUCKET="${2?Specify second argument - source bucket}"
RELEASE_BUCKET="${3?Specify third argument - release bucket}"
CDN="${4?Specify fourth argument - cdn}"
SOURCE_CONTAINER_REGISTRY="${5?Specify fifth argument - source container registry}"
RELEASE_CONTAINER_REGISTRY="${6?Specify sixth argument - release container registry}"

mkdir -p "${ARTIFACTS_DIR}"

# CODEBUILD_INITIATOR is a CodeBuild environment variable
# of the form `codepipeline/<pipeline_name>`
PIPELINE_NAME=$(cut -d/ -f2 <<< "$CODEBUILD_INITIATOR")

# Querying the execution ID corresponding to the pipeline
# execution that triggered this dev release
PIPELINE_EXECUTION_ID="$(aws codepipeline list-action-executions --pipeline-name $PIPELINE_NAME | jq -r '.actionExecutionDetails[] | select(.output.executionResult.externalExecutionId=='\"$CODEBUILD_BUILD_ID\"').pipelineExecutionId')"

# Using the pipeline execution ID to query the branch
# corresponding to the eks-anywhere-build-tooling repo
# source action for this particulat execution
BRANCH_NAME="$(aws codepipeline list-action-executions --pipeline-name $PIPELINE_NAME | jq -r '.actionExecutionDetails[] | select(.pipelineExecutionId=='\"$PIPELINE_EXECUTION_ID\"') | select(.input.configuration.RepositoryName=="aws.eks-anywhere-build-tooling").input.configuration.BranchName')"

# This is a catch-all for pipelines that do not source
# build-tooling repository, for example, the CLI pipeline
if [ "$BRANCH_NAME" = "" ]; then
	BRANCH_NAME="main"
fi

${BASE_DIRECTORY}/release/bin/eks-anywhere-release release \
	--artifact-dir "${ARTIFACTS_DIR}" \
	--branch-name "${BRANCH_NAME}" \
	--source-bucket "${SOURCE_BUCKET}" \
	--source-container-registry "${SOURCE_CONTAINER_REGISTRY}" \
	--cdn "${CDN}" \
	--release-bucket "${RELEASE_BUCKET}" \
	--release-container-registry "${RELEASE_CONTAINER_REGISTRY}" \
	--dev-release=true
