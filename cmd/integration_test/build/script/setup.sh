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

if [ "$TEST_ROLE_ARN" == "" ]; then
    echo "Empty TEST_ROLE_ARN, this script is used in CodeBuild to set profile to assume an IAM role"
    exit 1
fi

export AWS_SDK_LOAD_CONFIG=1
export AWS_CONFIG_FILE=$(pwd)/awscliconfig
export AWS_PROFILE=e2e-test-account

cat << EOF > ${AWS_CONFIG_FILE}
[profile ${AWS_PROFILE}]
role_arn=${TEST_ROLE_ARN}
region=us-west-2
credential_source=EcsContainer
EOF

source /docker.sh
start::dockerd
wait::for::dockerd
