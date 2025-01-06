#!/usr/bin/env bash

# Usage: ./update_vsphere_creds.sh <CLUSTER_NAME> <VSPHERE_SERVER_NAME>
set -o errexit
set -o nounset
set -o pipefail

if [ "$#" -ne 2 ]; then
    echo "Usage: ./update_vsphere_creds.sh <CLUSTER_NAME> <VSPHERE_SERVER_NAME>"
    exit 1
fi

cluster_name=$1
vsphere_server_name=$2
password=$EKSA_VSPHERE_PASSWORD
username=$EKSA_VSPHERE_USERNAME
encoded_password="$(echo -n $password | base64)"
encoded_username="$(echo -n $username | base64)"

# Patch {CLUSTER_NAME}-vsphere-credentials in eksa-system
kubectl patch -n eksa-system secrets "$cluster_name-vsphere-credentials" --patch="{\"data\":{\"password\":\"$encoded_password\"}}"
last_applied=$(kubectl get secrets -n eksa-system "$cluster_name-vsphere-credentials" -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration}')
if [[ $last_applied ]]; then
  new_annotation=$(echo $last_applied | jq -c --arg password $encoded_password '.data.password=$password')
  kubectl annotate --overwrite -n eksa-system secrets "$cluster_name-vsphere-credentials" kubectl.kubernetes.io/last-applied-configuration=$new_annotation
fi

# Patch vsphere-credentials in eksa-system
kubectl patch -n eksa-system secrets vsphere-credentials --patch="{\"data\":{\"password\":\"$encoded_password\",\"passwordCP\":\"$encoded_password\"}}"
if [[ $(kubectl get secrets -n eksa-system vsphere-credentials -o jsonpath='{.data.passwordCSI}') ]]; then
  kubectl patch -n eksa-system secrets vsphere-credentials --patch="{\"data\":{\"passwordCSI\":\"$encoded_password\"}}"
fi
last_applied=$(kubectl get secrets -n eksa-system vsphere-credentials -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration}')
if [[ $last_applied ]]; then
  new_annotation=$(echo $last_applied | jq -c --arg password $encoded_password '.data.password=$password | .data.passwordCP=$password | if (.data | has("passwordCSI")) then .data.passwordCSI=$password else. end')
  kubectl annotate --overwrite -n eksa-system secrets vsphere-credentials kubectl.kubernetes.io/last-applied-configuration=$new_annotation
fi

# Patch {CLUSTER_NAME}-cloud-provider-vsphere-credentials in eksa-system
cloud_provider_vsphere_credential=$(cat <<-END
apiVersion: v1
kind: Secret
metadata:
  name: cloud-provider-vsphere-credentials
  namespace: kube-system
data:
  $vsphere_server_name.password: $encoded_password
  $vsphere_server_name.username: $encoded_username
type: Opaque
END

)

encoded_cloud_provider_vsphere_credential=$(echo "$cloud_provider_vsphere_credential" | base64)

kubectl patch -n eksa-system secrets "$cluster_name-cloud-provider-vsphere-credentials" --patch="{\"data\":{\"data\":\"$encoded_cloud_provider_vsphere_credential\"}}"
last_applied=$(kubectl get secrets -n eksa-system "$cluster_name-cloud-provider-vsphere-credentials" -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration}')
if [[ $last_applied ]]; then
  new_annotation=$(echo $last_applied | jq -c --arg data $encoded_cloud_provider_vsphere_credential '.data.data=$data')
  kubectl annotate --overwrite -n eksa-system secrets "$cluster_name-cloud-provider-vsphere-credentials" kubectl.kubernetes.io/last-applied-configuration=$new_annotation
fi
