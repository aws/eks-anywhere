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

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
source ${SCRIPT_ROOT}/setup-aws-config.sh
set_aws_config "production"

EKSA_RELEASE_MANIFEST_URL="${1?Specify first argument - EKS-A release manifest URL}"

LATEST_EKSA_VERSION=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.latestVersion')
LATEST_EKSA_RELEASE_NUMBER=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.releases[] | select(.version == '\"$LATEST_EKSA_VERSION\"') | .number')
LATEST_BUNDLE_RELEASE_MANIFEST_URI=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.releases[] | select(.version == '\"$LATEST_EKSA_VERSION\"') | .bundleManifestUrl')
LATEST_EKSA_ADMIN_AMI_URI="https://anywhere-assets.eks.amazonaws.com/releases/eks-a/$LATEST_EKSA_RELEASE_NUMBER/artifacts/eks-a-admin-ami/$LATEST_EKSA_VERSION/eks-anywhere-admin-ami-$LATEST_EKSA_VERSION-eks-a-$LATEST_EKSA_RELEASE_NUMBER.raw"

NOTIFICATION_SUBJECT="New release of EKS Anywhere - version $LATEST_EKSA_VERSION"
NOTIFICATION_BODY="Amazon EKS Anywhere version $LATEST_EKSA_VERSION has been released. You can get the latest EKS-A CLI tarballs from GitHub releases. The bundle release manifest corresponding to this version of EKS-A is $LATEST_BUNDLE_RELEASE_MANIFEST_URI. This release includes a new build of the EKS-A Admin AMI image which is available at $LATEST_EKSA_ADMIN_AMI_URI"

echo "
Sending SNS message with the following details:
    - Topic ARN: $EKSA_RELEASE_SNS_TOPIC_ARN
    - Subject: $NOTIFICATION_SUBJECT
    - Body: $NOTIFICATION_BODY"

SNS_MESSAGE_ID=$(
    aws sns publish \
        --topic-arn "$EKSA_RELEASE_SNS_TOPIC_ARN" \
        --subject "$NOTIFICATION_SUBJECT" \
        --message "$NOTIFICATION_BODY" \
        --query "MessageId" \
        --output text
)

if [ "$SNS_MESSAGE_ID" ]; then
  echo -e "\nRelease notification published with SNS MessageId $SNS_MESSAGE_ID"
else
  echo "Received unexpected response while publishing to release SNS topic. An error may have occurred, and the \
notification may not have not have been published"
  exit 1
fi
