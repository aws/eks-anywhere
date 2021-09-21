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
set -o pipefail

GOLANG_VERSION="${1?Specify fourth argument - golang version}"
BINARY_NAME="${2?Specify fifth argument - binary name}"
IMAGE_REPO="${3?Specify sixth argument - ecr image repo}"
IMAGE_TAG="${4?Specify seventh argument - ecr image tag}"

BIN_ROOT="_output/bin"
BIN_PATH=$BIN_ROOT/$BINARY_NAME

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "$REPO_ROOT/scripts/common.sh"

KUSTOMIZE_BIN="${REPO_ROOT}/_output/kustomize-bin"

function build::install::kustomize(){
  curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
  mv kustomize $KUSTOMIZE_BIN
  export PATH=$KUSTOMIZE_BIN:$PATH
}

function build::eks-anywhere-cluster-controller::create_binaries(){
  platform=${1}
  OS="$(cut -d '/' -f1 <<< ${platform})"
  ARCH="$(cut -d '/' -f2 <<< ${platform})"
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH make build-cluster-controller-binaries
  mkdir -p ../${BIN_PATH}/${OS}-${ARCH}/
  mv bin/* ../${BIN_PATH}/${OS}-${ARCH}/
}

function build::eks-anywhere-cluster-controller::manifests(){
  MANIFEST_IMAGE_NAME="$(yq e '.images[] | select(.name == "controller") | .newName' ./config/prod/kustomization.yaml)"
  MANIFEST_IMAGE_TAG="$(yq e '.images[] | select(.name == "controller") | .newTag' ./config/prod/kustomization.yaml)"
  MANIFEST_IMAGE_NAME_OVERRIDE="${IMAGE_REPO}/eks-anywhere-cluster-controller"
  MANIFEST_IMAGE_TAG_OVERRIDE=${IMAGE_TAG}

  KUBE_RBAC_PROXY_IMAGE_NAME="$(yq e '.images[] | select(.name == "*kube-rbac-proxy") | .newName' ./config/prod/kustomization.yaml)"
  KUBE_RBAC_PROXY_IMAGE_TAG="$(yq e '.images[] | select(.name == "*kube-rbac-proxy") | .newTag' ./config/prod/kustomization.yaml)"
  KUBE_RBAC_PROXY_IMAGE_NAME_OVERRIDE="${IMAGE_REPO}/brancz/kube-rbac-proxy"
  KUBE_RBAC_PROXY_IMAGE_TAG_OVERRIDE="latest"

  sed -i "s,${MANIFEST_IMAGE_NAME},${MANIFEST_IMAGE_NAME_OVERRIDE}," ./config/prod/kustomization.yaml
  sed -i "s,${MANIFEST_IMAGE_TAG},${MANIFEST_IMAGE_TAG_OVERRIDE}," ./config/prod/kustomization.yaml
  sed -i "s,${KUBE_RBAC_PROXY_IMAGE_NAME},${KUBE_RBAC_PROXY_IMAGE_NAME_OVERRIDE}," ./config/prod/kustomization.yaml
  sed -i "s,${KUBE_RBAC_PROXY_IMAGE_TAG},${KUBE_RBAC_PROXY_IMAGE_TAG_OVERRIDE}," ./config/prod/kustomization.yaml

  mkdir -p _output/manifests/cluster-controller
  make release-manifests RELEASE_DIR=.
  cp eksa-components.yaml "_output/manifests/cluster-controller/"
}

function build::eks-anywhere-cluster-controller::binaries(){
  cd $REPO_ROOT
  mkdir -p $BIN_PATH
  mkdir -p $KUSTOMIZE_BIN
  build::install::kustomize
  build::common::use_go_version $GOLANG_VERSION
  go mod vendor
  build::eks-anywhere-cluster-controller::create_binaries "linux/amd64"
  build::eks-anywhere-cluster-controller::manifests
  build::gather_licenses $REPO_ROOT/_output "./controllers"
}

build::eks-anywhere-cluster-controller::binaries
