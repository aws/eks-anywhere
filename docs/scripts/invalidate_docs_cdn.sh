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

set -o errexit
set -o nounset
set -o pipefail

SANITIZED_BRANCH_NAME=$(echo $AWS_BRANCH | tr -d '.-')
DISTRIBUTION_ID_PARAMETER=$(aws ssm describe-parameters --no-shared --query "Parameters[?contains(Name, '${SANITIZED_BRANCH_NAME}')].Name" --output text)
DISTRIBUTION_ID=$(aws ssm get-parameter --name $DISTRIBUTION_ID_PARAMETER --with-decryption --query "Parameter.Value" --output text)
echo "Invalidating distribution $DISTRIBUTION_ID"

COMPLETED_STATUS="Completed"
INVALIDATION_RESPONSE=$(aws cloudfront create-invalidation --distribution-id $DISTRIBUTION_ID --paths "/*")
INVALIDATION_ID=$(echo $INVALIDATION_RESPONSE | jq -r '.Invalidation.Id')
INVALIDATION_STATUS=$(echo $INVALIDATION_RESPONSE | jq -r '.Invalidation.Status')

until [[ "$INVALIDATION_STATUS" == "$COMPLETED_STATUS" ]]; do
  echo "Invalidation status: $INVALIDATION_STATUS"
  sleep 5

  GET_RESPONSE=$(aws cloudfront get-invalidation --distribution-id $DISTRIBUTION_ID --id $INVALIDATION_ID)
  INVALIDATION_STATUS=$(echo $GET_RESPONSE | jq -r '.Invalidation.Status')
done
echo "Invalidation status: $INVALIDATION_STATUS"
