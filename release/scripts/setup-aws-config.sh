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

function set_aws_config() {
    release_environment="$1"
    if [ "$release_environment" = "" ] || [ "$release_environment" = "development" ]; then
        cat << EOF > awscliconfig
[profile packages-beta]
role_arn=$PACKAGES_ECR_ROLE
region=us-west-2
credential_source=EcsContainer
EOF
    fi
    if [ "$release_environment" = "development" ] || [ "$release_environment" = "production" ]; then
        if [ "$STAGING_ARTIFACT_DEPLOYMENT_ROLE" = "" ]; then
            echo "Empty STAGING_ARTIFACT_DEPLOYMENT_ROLE"
            exit 1
        fi
        cat << EOF >> awscliconfig
[profile artifacts-staging]
role_arn=$STAGING_ARTIFACT_DEPLOYMENT_ROLE
region=us-east-1
credential_source=EcsContainer
EOF

        if [ "$release_environment" = "production" ]; then
            if [ "$PROD_ARTIFACT_DEPLOYMENT_ROLE" = "" ]; then
                echo "Empty PROD_ARTIFACT_DEPLOYMENT_ROLE"
                exit 1
            fi
            cat << EOF >> awscliconfig
[profile artifacts-production]
role_arn=$PROD_ARTIFACT_DEPLOYMENT_ROLE
region=us-east-1
credential_source=EcsContainer
EOF
        fi
fi

    export AWS_CONFIG_FILE=$(pwd)/awscliconfig
}
