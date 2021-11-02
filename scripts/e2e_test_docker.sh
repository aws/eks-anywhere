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

set -x

if [ "$AWS_ROLE_ARN" == "" ]; then
    echo "Empty AWS_ROLE_ARN, this script must be run in a postsubmit pod with IAM Roles for Service Accounts"
    exit 1
fi

if [ "$TEST_ROLE_ARN" == "" ]; then
    echo "Empty TEST_ROLE_ARN, this script must be run in a postsubmit pod with IAM Roles for Service Accounts"
    exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel)
TEST_REGEX="${1:-TestDockerKubernetes121SimpleFlow}"

cat << EOF > config_file
[default]
output=json
region=${AWS_REGION:-${AWS_DEFAULT_REGION:-us-west-2}}
role_arn=$AWS_ROLE_ARN
web_identity_token_file=/var/run/secrets/eks.amazonaws.com/serviceaccount/token
[profile e2e-docker-test]
role_arn=$TEST_ROLE_ARN
region=${AWS_REGION:-${AWS_DEFAULT_REGION:-us-west-2}}
source_profile=default
EOF
export AWS_SDK_LOAD_CONFIG=true
export AWS_CONFIG_FILE=$(pwd)/config_file
export AWS_PROFILE=e2e-docker-test
unset AWS_ROLE_ARN AWS_WEB_IDENTITY_TOKEN_FILE

BUNDLES_OVERRIDE=false
if [ -f "$REPO_ROOT/bin/local-bundle-release.yaml" ]; then
    BUNDLES_OVERRIDE=true
fi
$REPO_ROOT/bin/test e2e run \
    -a ${INTEGRATION_TEST_AL2_AMI_ID} \
    -s ${INTEGRATION_TEST_STORAGE_BUCKET} \
    -j ${JOB_ID} \
    -i ${INTEGRATION_TEST_INSTANCE_PROFILE} \
    -r ${TEST_REGEX} \
    --bundles-override=${BUNDLES_OVERRIDE} \
    -g AWS_IAM_AUTHENTICATOR