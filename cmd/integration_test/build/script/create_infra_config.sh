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

source ${CODEBUILD_SRC_DIR}/test/e2e/E2E_AMI_FILTER_VARS

INTEGRATION_TEST_AMI_ID=$(aws ec2 describe-images \
  --profile ${AWS_PROFILE} \
  --owners ${AMI_OWNER_ID_FILTER} \
  --filters "Name=name,Values=${AMI_NAME_FILTER}" "Name=description,Values=${AMI_DESCRIPTION_FILTER}" \
  --query 'sort_by(Images, &CreationDate)[-1].[ImageId]' \
  --output text
)

if [ -z "$INTEGRATION_TEST_AMI_ID" ]; then
  echo "INTEGRATION_TEST_AMI_ID cannot be empty. Exiting"
  exit 1
fi

cat << EOF > ${INTEGRATION_TEST_INFRA_CONFIG}
---

ec2:
  amiId: ${INTEGRATION_TEST_AMI_ID}
  subnetId: ${INTEGRATION_TEST_SUBNET_ID}

vSphere:
  url: ${TEST_RUNNER_GOVC_URL}
  insecure: True
  library: ${TEST_RUNNER_GOVC_LIBRARY}
  template: ${TEST_RUNNER_GOVC_TEMPLATE}
  datacenter: ${TEST_RUNNER_GOVC_DATACENTER}
  datastore: ${TEST_RUNNER_GOVC_DATASTORE}
  resourcePool: ${TEST_RUNNER_GOVC_RESOURCE_POOL}
  network: ${TEST_RUNNER_GOVC_NETWORK}
  folder: ${TEST_RUNNER_GOVC_FOLDER}
EOF
