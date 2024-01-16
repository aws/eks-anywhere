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
set -o errexit
set -o nounset
set -o pipefail

delete_cluster_controller_image() {
    aws ecr-public batch-delete-image --repository-name eks-anywhere-cluster-controller --image-ids=imageTag=${PULL_PULL_SHA}.${PULL_BASE_REF} --region us-east-1
}

trap 'unset AWS_PROFILE; delete_cluster_controller_image' EXIT

if [ "$AWS_ROLE_ARN" == "" ]; then
    echo "Empty AWS_ROLE_ARN, this script must be run in a postsubmit pod with IAM Roles for Service Accounts"
    exit 1
fi

if [ "$TEST_ROLE_ARN" == "" ]; then
    echo "Empty TEST_ROLE_ARN, this script must be run in a postsubmit pod with IAM Roles for Service Accounts"
    exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel)
BIN_FOLDER=$REPO_ROOT/bin
TEST_REGEX="${1:-TestDockerKubernetes125SimpleFlow}"
BRANCH_NAME="${2:-main}"


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

INTEGRATION_TEST_INFRA_CONFIG="/tmp/test-infra.yml"
export T_TINKERBELL_S3_INVENTORY_CSV_KEY="inventory/den80/den80-hardware.csv"

cat << EOF > ${INTEGRATION_TEST_INFRA_CONFIG}
---

ec2:
  amiId: ${INTEGRATION_TEST_AL2_AMI_ID}
  subnetId:

EOF

export AWS_SDK_LOAD_CONFIG=true
export AWS_CONFIG_FILE=$(pwd)/config_file
export AWS_PROFILE=e2e-docker-test
unset AWS_ROLE_ARN AWS_WEB_IDENTITY_TOKEN_FILE

BUNDLES_OVERRIDE=false
if [ -f "$BIN_FOLDER/local-bundle-release.yaml" ]; then
    BUNDLES_OVERRIDE=true
fi

# In order to be uploaded from the sidecar and used by the junit lens
# the junit reports need to be in /logs/*/junit*.xml
TEST_REPORT_FOLDER=/logs/artifacts

$BIN_FOLDER/test e2e run \
    -c ${INTEGRATION_TEST_INFRA_CONFIG} \
    -s ${INTEGRATION_TEST_STORAGE_BUCKET} \
    -j ${JOB_ID} \
    -i ${INTEGRATION_TEST_INSTANCE_PROFILE} \
    -r ${TEST_REGEX} \
    --bundles-override=${BUNDLES_OVERRIDE} \
    --test-report-folder=${TEST_REPORT_FOLDER} \
    --baremetal-branch="" \
    -v4

# Faking cross-platform versioned folders for dry-run
mkdir -p $BIN_FOLDER/linux/amd64
cp $BIN_FOLDER/eksctl-anywhere $BIN_FOLDER/linux/amd64/eksctl-anywhere

$REPO_ROOT/cmd/integration_test/build/script/upload_artifacts.sh \
    "s3://artifacts-bucket" \
    $REPO_ROOT \
    "eks-a-cli" \
    $PROW_JOB_ID \
    $PULL_PULL_SHA \
    "linux" \
    "amd64" \
    $BRANCH_NAME \
    true
