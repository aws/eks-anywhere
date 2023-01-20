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

EKSA_RELEASE_CDN="${1?Specify first argument - EKS-A release assets CDN}"

EKSA_RELEASE_MANIFEST_URL="${EKSA_RELEASE_CDN}/releases/eks-a/manifest.yaml"
LATEST_EKSA_VERSION=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.latestVersion')
LATEST_EKSA_RELEASE_NUMBER=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.releases[] | select(.version == '\"$LATEST_EKSA_VERSION\"') | .number')
LATEST_BUNDLE_RELEASE_MANIFEST_URI=$(curl $EKSA_RELEASE_MANIFEST_URL | yq e '.spec.releases[] | select(.version == '\"$LATEST_EKSA_VERSION\"') | .bundleManifestUrl')
LATEST_EKSA_ADMIN_IMAGE_URI="$EKSA_RELEASE_CDN/releases/eks-a/$LATEST_EKSA_RELEASE_NUMBER/artifacts/eks-a-admin-ami/$LATEST_EKSA_VERSION/eks-anywhere-admin-ami-$LATEST_EKSA_VERSION-eks-a-$LATEST_EKSA_RELEASE_NUMBER.raw"

NOTIFICATION_SUBJECT="New release of EKS Anywhere - version $LATEST_EKSA_VERSION"

SNOW_NODE_IMAGES_JSON_ARRAY='[]'
readarray bundles < <(curl $LATEST_BUNDLE_RELEASE_MANIFEST_URI | yq e -o=j -I=0 ".spec.versionsBundles[]" -)
for bundle in "${bundles[@]}"; do
  short_kube_version=$(echo $bundle | yq ".kubeVersion" -)
  full_kube_version=$(echo $bundle | yq ".eksD.kubeVersion" -)
  eksd_release_name=$(echo $bundle | yq ".eksD.name" -)
  node_image_os=$(echo $bundle | yq ".eksD.raw.bottlerocket.os" -)
  node_image_os_name=$(echo $bundle | yq ".eksD.raw.bottlerocket.osName" -)
  node_image_uri=$(echo $bundle | yq ".eksD.raw.bottlerocket.uri" -)
  node_image_sha256=$(echo $bundle | yq ".eksD.raw.bottlerocket.sha256" -)
  node_image_sha512=$(echo $bundle | yq ".eksD.raw.bottlerocket.sha512" -)
  if [[ "$node_image_uri" != "null" ]]; then
    node_image_entry=$(jq -n \
      --arg short_kube_version $short_kube_version \
      --arg full_kube_version $full_kube_version \
      --arg eksd_release_name $eksd_release_name \
      --arg node_image_os $node_image_os \
      --arg node_image_os_name $node_image_os_name \
      --arg node_image_uri $node_image_uri \
      --arg node_image_sha256 $node_image_sha256 \
      --arg node_image_sha512 $node_image_sha512 \
      '{"kubeChannel": $short_kube_version, "kubeVersion": $full_kube_version, "eksdReleaseName": $eksd_release_name, "bottlerocket": {"uri": $node_image_uri, "os": $node_image_os, "osName": $node_image_os_name, "sha256": $node_image_sha256, "sha512": $node_image_sha256}}'
    )
    SNOW_NODE_IMAGES_JSON_ARRAY=$(echo $SNOW_NODE_IMAGES_JSON_ARRAY | jq --argjson node_image_entry "$node_image_entry" '. += [$node_image_entry]')
  fi
done

EKSA_ADMIN_IMAGE_JSON_OBJECT=$(jq -n \
  --arg admin_image_uri $LATEST_EKSA_ADMIN_IMAGE_URI \
  --arg admin_image_os "linux" \
  --arg admin_image_os_name "amazonlinux2" \
  '{"uri": $admin_image_uri, "os": $admin_image_os, "osName": $admin_image_os_name}'
)

NOTIFICATION_MESSAGE=$(jq -n \
  --arg eksa_release_version $LATEST_EKSA_VERSION \
  --argjson eksa_admin_image "$EKSA_ADMIN_IMAGE_JSON_OBJECT" \
  --argjson snow_node_images "$SNOW_NODE_IMAGES_JSON_ARRAY" \
  '{"releaseVersion": $eksa_release_version, "eksaAdminImage": {"amazonlinux2": $eksa_admin_image}, "snowNodeImages": $snow_node_images}'
)

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
notification may not have been published"
  exit 1
fi
