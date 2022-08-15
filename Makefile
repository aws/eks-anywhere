export INTEGRATION_TEST_UBUNTU_AMI_ID?=integration_test_ami
export INTEGRATION_TEST_STORAGE_BUCKET?=integration_test_storage_bucket
export INTEGRATION_TEST_INSTANCE_PROFILE?=integration_test_instance_profile
export INTEGRATION_TEST_MAX_INSTANCE_AGE?=86400
export INTEGRATION_TEST_SUBNET_ID?=integration_test_subnet_id
export INTEGRATION_TEST_INSTANCE_TAG?=integration_test_instance_tag
export JOB_ID?=${PROW_JOB_ID}

LOCAL_ENV=.env

# Include local .env file for exporting e2e test env vars locally
# FYI, This file is git ignored
ifneq ("$(wildcard $(LOCAL_ENV))","")
include $(LOCAL_ENV)
export $(shell sed 's/=.*//' $(LOCAL_ENV))
endif

SHELL := /bin/bash

ARTIFACTS_BUCKET?=my-s3-bucket
GIT_VERSION?=$(shell git describe --tag)
GIT_TAG?=$(shell git tag -l --sort -v:refname | head -1)
GOLANG_VERSION?="1.17"
GO_VERSION ?= $(shell source ./scripts/common.sh && build::common::get_go_path $(GOLANG_VERSION))
GO ?= $(GO_VERSION)/go
GO_TEST ?= $(GO) test
# A regular expression defining what packages to exclude from the unit-test recipe.
UNIT_TEST_PACKAGE_EXCLUSION_REGEX ?=mocks$

## ensure local execution uses the 'main' branch bundle
BRANCH_NAME?=main
ifneq ($(PULL_BASE_REF),)
	BRANCH_NAME=$(PULL_BASE_REF)
endif
ifeq (,$(findstring $(BRANCH_NAME),main))
## use the branch-specific bundle manifest if the branch is not 'main'
DEV_GIT_VERSION:=v0.0.0-dev-${BRANCH_NAME}
BUNDLE_MANIFEST_URL?=https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/${BRANCH_NAME}/bundle-release.yaml
RELEASE_MANIFEST_URL?=https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/${BRANCH_NAME}/eks-a-release.yaml
LATEST=$(BRANCH_NAME)
$(info    Using branch-specific BUNDLE_MANIFEST_URL $(BUNDLE_MANIFEST_URL) and RELEASE_MANIFEST_URL $(RELEASE_MANIFEST_URL))
else
## use the standard bundle manifest if the branch is 'main'
DEV_GIT_VERSION:=v0.0.0-dev
BUNDLE_MANIFEST_URL?=https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/bundle-release.yaml
RELEASE_MANIFEST_URL?=https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/eks-a-release.yaml
$(info    Using standard BUNDLE_MANIFEST_URL $(BUNDLE_MANIFEST_URL) and RELEASE_MANIFEST_URL $(RELEASE_MANIFEST_URL))
LATEST=latest
endif

CUSTOM_GIT_VERSION:=v0.0.0-custom

AWS_ACCOUNT_ID?=$(shell aws sts get-caller-identity --query Account --output text)
AWS_REGION=us-west-2

BIN_DIR := bin
TOOLS_BIN_DIR := hack/tools/bin

OUTPUT_DIR := _output
OUTPUT_BIN_DIR := ${OUTPUT_DIR}/bin

KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize
KUSTOMIZE_VERSION := 4.2.0

KUSTOMIZE_OUTPUT_BIN_DIR="${OUTPUT_DIR}/kustomize-bin/"

KUBEBUILDER := $(TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_VERSION := v3.1.0

BUILD_LIB := build/lib
BUILDKIT := $(BUILD_LIB)/buildkit.sh

CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)

SETUP_ENVTEST_BIN := setup-envtest
SETUP_ENVTEST := $(TOOLS_BIN_DIR)/$(SETUP_ENVTEST_BIN)

BINARY_NAME=eks-anywhere-cluster-controller
ifdef CODEBUILD_SRC_DIR
	TAR_PATH?=$(CODEBUILD_SRC_DIR)/$(PROJECT_PATH)/$(CODEBUILD_BUILD_NUMBER)-$(CODEBUILD_RESOLVED_SOURCE_VERSION)/artifacts
else
	TAR_PATH?="$(OUTPUT_DIR)/tar"
endif

BASE_REPO?=public.ecr.aws/eks-distro-build-tooling
CLUSTER_CONTROLLER_BASE_IMAGE_NAME?=eks-distro-minimal-base
CLUSTER_CONTROLLER_BASE_TAG?=$(shell cat manager/EKS_DISTRO_MINIMAL_BASE_TAG_FILE)
CLUSTER_CONTROLLER_BASE_IMAGE?=$(BASE_REPO)/$(CLUSTER_CONTROLLER_BASE_IMAGE_NAME):$(CLUSTER_CONTROLLER_BASE_TAG)
DOCKERFILE_FOLDER = ./manager/docker/linux/eks-anywhere-cluster-controller

IMAGE_REPO?=$(if $(AWS_ACCOUNT_ID),$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com,localhost:5000)
IMAGE_TAG?=$(GIT_TAG)-$(shell git rev-parse HEAD)
CLUSTER_CONTROLLER_IMAGE_NAME=eks-anywhere-cluster-controller
CLUSTER_CONTROLLER_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):$(IMAGE_TAG)
CLUSTER_CONTROLLER_LATEST_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):$(LATEST)

# Branch builds should look at the current branch latest image for cache as well as main branch latest for cache to cover the cases
# where its the first build from a new release branch
IMAGE_IMPORT_CACHE = type=registry,ref=$(CLUSTER_CONTROLLER_LATEST_IMAGE) type=registry,ref=$(subst $(LATEST),latest,$(CLUSTER_CONTROLLER_LATEST_IMAGE))

MANIFEST_IMAGE_NAME_OVERRIDE="$(IMAGE_REPO)/eks-anywhere-cluster-controller"
MANIFEST_IMAGE_TAG_OVERRIDE=${IMAGE_TAG}
KUSTOMIZATION_CONFIG=./config/prod/kustomization.yaml

CONTROLLER_MANIFEST_OUTPUT_DIR=$(OUTPUT_DIR)/manifests/cluster-controller

BUILD_TAGS :=
BUILD_FLAGS?=

GO_ARCH:=$(shell go env GOARCH)
GO_OS:=$(shell go env GOOS)

BINARY_DEPS_DIR = $(OUTPUT_DIR)/dependencies
CLUSTER_CONTROLLER_PLATFORMS ?= linux-amd64 linux-arm64
CREATE_CLUSTER_CONTROLLER_BINARIES := $(foreach platform,$(CLUSTER_CONTROLLER_PLATFORMS),create-cluster-controller-binary-$(platform))
FETCH_BINARIES_TARGETS = eksa/vmware/govmomi eksa/helm/helm
ORGANIZE_BINARIES_TARGETS = $(addsuffix /eks-a-tools,$(addprefix $(BINARY_DEPS_DIR)/linux-,amd64 arm64))

EKS_A_PLATFORMS ?= linux-amd64 linux-arm64 darwin-arm64 darwin-amd64
EKS_A_CROSS_PLATFORMS := $(foreach platform,$(EKS_A_PLATFORMS),eks-a-cross-platform-$(platform))
E2E_CROSS_PLATFORMS := $(foreach platform,$(EKS_A_PLATFORMS),e2e-cross-platform-$(platform))
EKS_A_RELEASE_CROSS_PLATFORMS := $(foreach platform,$(EKS_A_PLATFORMS),eks-a-release-cross-platform-$(platform))

DOCKER_E2E_TEST := TestDockerKubernetes121SimpleFlow
LOCAL_E2E_TESTS ?= $(DOCKER_E2E_TEST)

export KUBEBUILDER_ENVTEST_KUBERNETES_VERSION ?= 1.21.x

UNAME := $(shell uname -s)

.PHONY: default
default: build lint

.PHONY: build
build: eks-a eks-a-tool unit-test ## Generate binaries and run unit tests

.PHONY: release
release: eks-a-release unit-test ## Generate release binary and run unit tests

.PHONY: eks-a-binary
eks-a-binary: ALL_LINKER_FLAGS := $(LINKER_FLAGS) -X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(GIT_VERSION) -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=$(RELEASE_MANIFEST_URL) -X github.com/aws/eks-anywhere/pkg/manifests/releases.manifestURL=$(RELEASE_MANIFEST_URL) -s -w -buildid='' -extldflags -static
eks-a-binary: LINKER_FLAGS_ARG := -ldflags "$(ALL_LINKER_FLAGS)"
eks-a-binary: BUILD_TAGS_ARG := -tags "$(BUILD_TAGS)"
eks-a-binary: OUTPUT_FILE ?= bin/eksctl-anywhere
eks-a-binary:
	CGO_ENABLED=0 GOOS=$(GO_OS) GOARCH=$(GO_ARCH) $(GO) build $(BUILD_TAGS_ARG) $(LINKER_FLAGS_ARG) $(BUILD_FLAGS) -o $(OUTPUT_FILE) github.com/aws/eks-anywhere/cmd/eksctl-anywhere

.PHONY: eks-a-embed-config
eks-a-embed-config: ## Build a dev release version of eks-a with embed cluster spec config
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) RELEASE_MANIFEST_URL=embed:///config/releases.yaml BUILD_TAGS='$(BUILD_TAGS) spec_embed_config'

.PHONY: eks-a-cross-platform-embed-latest-config
eks-a-cross-platform-embed-latest-config: ## Build cross platform dev release versions of eks-a with the latest bundle-release.yaml embedded in cluster spec config
	curl -L $(BUNDLE_MANIFEST_URL) --output pkg/cluster/config/bundle-release.yaml
	$(MAKE) eks-a-embed-config GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/amd64/eksctl-anywhere
	$(MAKE) eks-a-embed-config GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/amd64/eksctl-anywhere
	$(MAKE) eks-a-embed-config GO_OS=darwin GO_ARCH=arm64 OUTPUT_FILE=bin/darwin/arm64/eksctl-anywhere
	$(MAKE) eks-a-embed-config GO_OS=linux GO_ARCH=arm64 OUTPUT_FILE=bin/linux/arm64/eksctl-anywhere
	rm pkg/cluster/config/bundle-release.yaml

.PHONY: eks-a-custom-embed-config
eks-a-custom-embed-config:
	$(MAKE) eks-a-binary GIT_VERSION=$(CUSTOM_GIT_VERSION) RELEASE_MANIFEST_URL=embed:///config/releases.yaml LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true' BUILD_TAGS='$(BUILD_TAGS) spec_embed_config'

.PHONY: eks-a-cross-platform-custom-embed-latest-config
eks-a-cross-platform-custom-embed-latest-config: ## Build custom binary with latest dev release bundle that embeds config and builds it as a release binary for all os/arch
	curl -L $(BUNDLE_MANIFEST_URL) --output pkg/cluster/config/bundle-release.yaml
	$(MAKE) eks-a-custom-embed-config GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/amd64/eksctl-anywhere
	$(MAKE) eks-a-custom-embed-config GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/amd64/eksctl-anywhere
	$(MAKE) eks-a-custom-embed-config GO_OS=darwin GO_ARCH=arm64 OUTPUT_FILE=bin/darwin/arm64/eksctl-anywhere
	$(MAKE) eks-a-custom-embed-config GO_OS=linux GO_ARCH=arm64 OUTPUT_FILE=bin/linux/arm64/eksctl-anywhere
	rm pkg/cluster/config/bundle-release.yaml

.PHONY: eks-a-custom-release-zip
eks-a-custom-release-zip: eks-a-cross-platform-custom-embed-latest-config ## Build from linux/amd64
	rm -f bin/eksctl-anywhere.zip ## Remove previous zip
	zip -j bin/eksctl-anywhere.zip bin/linux/amd64/eksctl-anywhere

.PHONY: eks-a-cross-platform-custom-release-zip
eks-a-cross-platform-custom-release-zip: eks-a-cross-platform-custom-embed-latest-config
	rm -f bin/eksctl-anywhere-darwin-amd64.zip bin/eksctl-anywhere-linux-amd64.zip bin/eksctl-anywhere-darwin-arm64.zip bin/eksctl-anywhere-linux-arm64.zip
	zip -j bin/eksctl-anywhere-darwin-amd64.zip bin/darwin/amd64/eksctl-anywhere
	zip -j bin/eksctl-anywhere-linux-amd64.zip bin/linux/amd64/eksctl-anywhere
	zip -j bin/eksctl-anywhere-darwin-arm64.zip bin/darwin/arm64/eksctl-anywhere
	zip -j bin/eksctl-anywhere-linux-arm64.zip bin/linux/arm64/eksctl-anywhere

.PHONY: eks-a
eks-a: ## Build a dev release version of eks-a
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION)

.PHONY: eks-a-release
eks-a-release: ## Generate a release binary
	$(MAKE) eks-a-binary GO_OS=linux GO_ARCH=amd64 LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'

.PHONY: eks-a-cross-platform
eks-a-cross-platform: $(EKS_A_CROSS_PLATFORMS)

.PHONY: eks-a-cross-platform-%
eks-a-cross-platform-%: ## Generate binaries for Linux and MacOS
eks-a-cross-platform-%: GO_OS = $(firstword $(subst -, ,$*))
eks-a-cross-platform-%: GO_ARCH = $(lastword $(subst -, ,$*))
eks-a-cross-platform-%:
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) GO_OS=$(GO_OS) GO_ARCH=$(GO_ARCH) OUTPUT_FILE=bin/$(GO_OS)/$(GO_ARCH)/eksctl-anywhere

.PHONY: eks-a-release-cross-platform
eks-a-release-cross-platform: $(EKS_A_RELEASE_CROSS_PLATFORMS)

.PHONY: eks-a-release-cross-platform-%
eks-a-release-cross-platform-%: ## Generate binaries for Linux and MacOS
eks-a-release-cross-platform-%: GO_OS = $(firstword $(subst -, ,$*))
eks-a-release-cross-platform-%: GO_ARCH = $(lastword $(subst -, ,$*))
eks-a-release-cross-platform-%:
	$(MAKE) eks-a-binary GIT_VERSION=$(GIT_VERSION) GO_OS=$(GO_OS) GO_ARCH=$(GO_ARCH) OUTPUT_FILE=bin/$(GO_OS)/$(GO_ARCH)/eksctl-anywhere LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true' BUILD_FLAGS='-trimpath'

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

$(OUTPUT_BIN_DIR): $(OUTPUT_DIR)
	mkdir -p $(OUTPUT_BIN_DIR)

$(KUSTOMIZE_OUTPUT_BIN_DIR): $(OUTPUT_DIR)
	mkdir -p $(KUSTOMIZE_OUTPUT_BIN_DIR)

$(CONTROLLER_MANIFEST_OUTPUT_DIR):
	mkdir -p $(CONTROLLER_MANIFEST_OUTPUT_DIR)

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(KUSTOMIZE): $(TOOLS_BIN_DIR) $(KUSTOMIZE_OUTPUT_BIN_DIR)
	-rm $(TOOLS_BIN_DIR)/kustomize
	cd $(TOOLS_BIN_DIR) && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s $(KUSTOMIZE_VERSION)
	cp $(TOOLS_BIN_DIR)/kustomize $(KUSTOMIZE_OUTPUT_BIN_DIR)

$(KUBEBUILDER): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR) && curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/$(KUBEBUILDER_VERSION)/kubebuilder_$(GO_OS)_$(GO_ARCH)
	chmod +x $(KUBEBUILDER)

$(CONTROLLER_GEN): $(TOOLS_BIN_DIR)
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1)

$(SETUP_ENVTEST): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR); $(GO) build -tags=tools -o $(SETUP_ENVTEST_BIN) sigs.k8s.io/controller-runtime/tools/setup-envtest

.PHONY: lint
lint: bin/golangci-lint ## Run golangci-lint
	bin/golangci-lint run

bin/golangci-lint: ## Download golangci-lint
bin/golangci-lint: GOLANGCI_LINT_VERSION?=$(shell cat .github/workflows/golangci-lint.yml | yq e '.jobs.golangci.steps[] | select(.name == "golangci-lint") .with.version' -)
bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

.PHONY: build-cross-platform
build-cross-platform: eks-a-cross-platform

.PHONY: eks-a-tool
eks-a-tool: ## Build eks-a-tool
	$(GO) build -o bin/eks-a-tool github.com/aws/eks-anywhere/cmd/eks-a-tool

.PHONY: eks-a-cluster-controller
eks-a-cluster-controller: ## Build eks-a-cluster-controller
	$(GO) build -ldflags "-s -w -buildid='' -extldflags -static" -o bin/manager ./manager

# This target will copy LICENSE file from root to the release submodule
# when fetching licenses for cluster-controller
.PHONY: copy-license-cluster-controller
copy-license-cluster-controller:
copy-license-cluster-controller:
	source scripts/attribution_helpers.sh && build::fix_licenses

.PHONY: build-cluster-controller-binaries
build-cluster-controller-binaries: eks-a-cluster-controller copy-license-cluster-controller

.PHONY: build-cluster-controller
build-cluster-controller: cluster-controller-local-images cluster-controller-tarballs

.PHONY: release-cluster-controller
release-cluster-controller: cluster-controller-images cluster-controller-tarballs

.PHONY: release-upload-cluster-controller
release-upload-cluster-controller: release-cluster-controller upload-artifacts

.PHONY: upload-artifacts
upload-artifacts:
	manager/build/upload_artifacts.sh $(TAR_PATH) $(ARTIFACTS_BUCKET) $(PROJECT_PATH) $(CODEBUILD_BUILD_NUMBER) $(CODEBUILD_RESOLVED_SOURCE_VERSION) $(LATEST)

.PHONY: create-cluster-controller-binaries
create-cluster-controller-binaries: $(CREATE_CLUSTER_CONTROLLER_BINARIES)

create-cluster-controller-binary-%:
	CGO_ENABLED=0 GOOS=$(firstword $(subst -, ,$*)) GOARCH=$(lastword $(subst -, ,$*)) $(MAKE) build-cluster-controller-binaries
	mkdir -p $(OUTPUT_BIN_DIR)/$(BINARY_NAME)/$*/
	mv bin/manager $(OUTPUT_BIN_DIR)/$(BINARY_NAME)/$*/

.PHONY: cluster-controller-binaries
cluster-controller-binaries: $(OUTPUT_BIN_DIR)
	mkdir -p $(OUTPUT_BIN_DIR)/$(BINARY_NAME)
	$(GO) mod vendor
	$(MAKE) create-cluster-controller-binaries
	$(MAKE) update-kustomization-yaml
	$(MAKE) release-manifests RELEASE_DIR=.
	source ./scripts/common.sh && build::gather_licenses $(OUTPUT_DIR) "./manager"

.PHONY: cluster-controller-tarballs
cluster-controller-tarballs:  cluster-controller-binaries
	manager/build/create_tarballs.sh $(BINARY_NAME) $(GIT_TAG) $(TAR_PATH)

.PHONY: cluster-controller-local-images
cluster-controller-local-images: IMAGE_PLATFORMS = linux/amd64
cluster-controller-local-images: IMAGE_OUTPUT_TYPE = oci
cluster-controller-local-images: IMAGE_OUTPUT = dest=/tmp/$(CLUSTER_CONTROLLER_IMAGE_NAME).tar
cluster-controller-local-images: cluster-controller-binaries $(ORGANIZE_BINARIES_TARGETS)
	$(BUILDCTL)

.PHONY: cluster-controller-images
cluster-controller-images: IMAGE_PLATFORMS = linux/amd64,linux/arm64
cluster-controller-images: IMAGE_OUTPUT_TYPE = image
cluster-controller-images: IMAGE_OUTPUT = push=true
cluster-controller-images: cluster-controller-binaries $(ORGANIZE_BINARIES_TARGETS)
	$(BUILDCTL)


.PHONY: generate-attribution
generate-attribution:
	scripts/make_attribution.sh $(GOLANG_VERSION)

.PHONY: update-attribution-files
update-attribution-files: generate-attribution
	scripts/create_pr.sh

.PHONY: generate-checksums
generate-checksums:
	scripts/generate_checksum.sh

.PHONY: update-brew-formula
update-brew-formula:
	scripts/brew_formula_pr.sh

.PHONY: clean
clean: ## Clean up resources created by make targets
	rm -rf ./bin/*
	rm -rf ./pkg/executables/cluster-name/
	rm -rf ./pkg/providers/vsphere/test/
ifeq ($(UNAME), Darwin)
	  find -E . -depth -type d -regex '.*\/Test.*-[0-9]{9}\/.*' -exec rm -rf {} \;
else
	  find . -depth -type d -regextype posix-egrep -regex '.*\/Test.*-[0-9]{9}\/.*' -exec rm -rf {} \;
endif
	rm -rf ./manager/bin/*
	rm -rf ./hack/tools/bin
	rm -rf vendor
	rm -rf GIT_TAG
	rm -rf _output
	make -C docs clean

#
# Generate zz_generated.deepcopy.go
#
generate: $(CONTROLLER_GEN)  ## Generate zz_generated.deepcopy.go
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: test
test: unit-test capd-test  ## Run unit and capd tests

.PHONY: unit-test
unit-test: ## Run unit tests
unit-test: $(SETUP_ENVTEST) 
unit-test: KUBEBUILDER_ASSETS ?= $(shell $(SETUP_ENVTEST) use --use-env -p path --arch amd64 $(KUBEBUILDER_ENVTEST_KUBERNETES_VERSION))
unit-test:
	KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" $(GO_TEST) $$($(GO) list ./... | grep -vE "$(UNIT_TEST_PACKAGE_EXCLUSION_REGEX)") -cover -tags "$(BUILD_TAGS)" $(GO_TEST_FLAGS)

.PHONY: coverage-unit-test
coverage-unit-test: COVER_PROFILE?=coverage.html 
coverage-unit-test:
	$(MAKE) unit-test GO_TEST_FLAGS="-coverprofile=$(COVER_PROFILE) -covermode=atomic"

.PHONY: local-e2e
local-e2e: e2e ## Run e2e test's locally
	./bin/e2e.test -test.v -test.run $(LOCAL_E2E_TESTS) $(GO_TEST_FLAGS)

.PHONY: capd-test
capd-test: ## Run default e2e capd test locally
	$(MAKE) local-e2e LOCAL_E2E_TESTS=$(DOCKER_E2E_TEST)

.PHONY: docker-e2e-test
docker-e2e-test: e2e ## Run docker integration test in new ec2 instance
	scripts/e2e_test_docker.sh $(DOCKER_E2E_TEST) $(BRANCH_NAME)

.PHONY: e2e-cleanup
e2e-cleanup: e2e ## Clean up resources generated by e2e tests
	scripts/e2e_cleanup.sh

.PHONY: capd-test-all
capd-test-all: capd-test capd-test-120

.PHONY: capd-test-%
capd-test-%: e2e ## Run CAPD tests
	./bin/e2e.test -test.v -test.run TestDockerKubernetes$*SimpleFlow


PACKAGES_E2E_TESTS ?= TestVSphereKubernetes121PackagesInstallSimpleFlow
ifeq ($(PACKAGES_E2E_TESTS),all)
PACKAGES_E2E_TESTS='TestCPackages.*'
endif
packages-e2e-test: e2e ## Run Curated Packages tests
	./bin/e2e.test -test.v -test.run $(PACKAGES_E2E_TESTS)

.PHONY: mocks
mocks: export PATH := $(GO_VERSION):$(PATH)
mocks: ## Generate mocks
	$(GO) install github.com/golang/mock/mockgen@v1.6.0
	${GOPATH}/bin/mockgen -destination=controllers/mocks/snow_machineconfig_controller.go -package=mocks -source "controllers/snow_machineconfig_controller.go"
	${GOPATH}/bin/mockgen -destination=pkg/providers/mocks/providers.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers" Provider,DatacenterConfig,MachineConfig
	${GOPATH}/bin/mockgen -destination=pkg/executables/mocks/executables.go -package=mocks "github.com/aws/eks-anywhere/pkg/executables" Executable,DockerClient,DockerContainer
	${GOPATH}/bin/mockgen -destination=pkg/providers/docker/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/docker" ProviderClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/tinkerbell/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell" ProviderKubectlClient,SSHAuthKeyGenerator
	${GOPATH}/bin/mockgen -destination=pkg/providers/cloudstack/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/cloudstack" ProviderCmkClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/vsphere" ProviderGovcClient,ProviderKubectlClient,ClusterResourceSetManager
	${GOPATH}/bin/mockgen -destination=pkg/filewriter/mocks/filewriter.go -package=mocks "github.com/aws/eks-anywhere/pkg/filewriter" FileWriter
	${GOPATH}/bin/mockgen -destination=pkg/clustermanager/mocks/client_and_networking.go -package=mocks "github.com/aws/eks-anywhere/pkg/clustermanager" ClusterClient,Networking,AwsIamAuth
	${GOPATH}/bin/mockgen -destination=pkg/gitops/flux/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/gitops/flux" FluxClient,KubeClient,GitOpsFluxClient,GitClient,Templater
	${GOPATH}/bin/mockgen -destination=pkg/task/mocks/task.go -package=mocks "github.com/aws/eks-anywhere/pkg/task" Task
	${GOPATH}/bin/mockgen -destination=pkg/bootstrapper/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/bootstrapper" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/cluster/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/cluster" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/git/providers/github/mocks/github.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/providers/github" GithubClient
	${GOPATH}/bin/mockgen -destination=pkg/git/mocks/git.go -package=mocks "github.com/aws/eks-anywhere/pkg/git" Client,ProviderClient
	${GOPATH}/bin/mockgen -destination=pkg/workflows/interfaces/mocks/clients.go -package=mocks "github.com/aws/eks-anywhere/pkg/workflows/interfaces" Bootstrapper,ClusterManager,GitOpsManager,Validator,CAPIManager,EksdInstaller,EksdUpgrader,PackageInstaller
	${GOPATH}/bin/mockgen -destination=pkg/git/gogithub/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogithub" Client
	${GOPATH}/bin/mockgen -destination=pkg/git/gitclient/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gitclient" GoGit
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/docker.go -package=mocks "github.com/aws/eks-anywhere/pkg/validations" DockerExecutable
	${GOPATH}/bin/mockgen -destination=controllers/resource/mocks/resource.go -package=mocks "github.com/aws/eks-anywhere/controllers/resource" ResourceFetcher,ResourceUpdater
	${GOPATH}/bin/mockgen -destination=controllers/resource/mocks/reader.go -package=mocks "sigs.k8s.io/controller-runtime/pkg/client" Reader
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/templates/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/templates/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/tags/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/tags/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/kubectl.go -package=mocks -source "pkg/validations/kubectl.go" KubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/tls.go -package=mocks -source "pkg/validations/tls.go" TlsValidator
	${GOPATH}/bin/mockgen -destination=pkg/diagnostics/interfaces/mocks/diagnostics.go -package=mocks -source "pkg/diagnostics/interfaces.go" DiagnosticBundle,AnalyzerFactory,CollectorFactory,BundleClient
	${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/capiclient.go -package=mocks -source "pkg/clusterapi/manager.go" CAPIClient,KubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/client.go -package=mocks -source "pkg/clusterapi/resourceset_manager.go" Client
	${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/fetch.go -package=mocks -source "pkg/clusterapi/fetch.go"
	${GOPATH}/bin/mockgen -destination=pkg/crypto/mocks/crypto.go -package=mocks -source "pkg/crypto/certificategen.go" CertificateGenerator
	${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/clients.go -package=mocks -source "pkg/networking/cilium/client.go"
	${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/helm.go -package=mocks -source "pkg/networking/cilium/templater.go"
	${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/upgrader.go -package=mocks -source "pkg/networking/cilium/upgrader.go"
	${GOPATH}/bin/mockgen -destination=pkg/networking/kindnetd/mocks/client.go -package=mocks -source "pkg/networking/kindnetd/upgrader.go"
	${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/cilium.go -package=mocks -source "pkg/networking/cilium/cilium.go"
	${GOPATH}/bin/mockgen -destination=pkg/networkutils/mocks/client.go -package=mocks -source "pkg/networkutils/netclient.go" NetClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/tinkerbell/hardware/mocks/translate.go -package=mocks -source "pkg/providers/tinkerbell/hardware/translate.go" MachineReader,MachineWriter,MachineValidator
	${GOPATH}/bin/mockgen -destination=pkg/providers/tinkerbell/stack/mocks/stack.go -package=mocks -source "pkg/providers/tinkerbell/stack/stack.go" Docker,Helm,StackInstaller
	${GOPATH}/bin/mockgen -destination=pkg/docker/mocks/mocks.go -package=mocks -source "pkg/docker/mover.go"
	${GOPATH}/bin/mockgen -destination=internal/test/mocks/reader.go -package=mocks -source "internal/test/reader.go"
	${GOPATH}/bin/mockgen -destination=cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks/download.go -package=mocks -source "cmd/eksctl-anywhere/cmd/internal/commands/artifacts/download.go"
	${GOPATH}/bin/mockgen -destination=cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks/import.go -package=mocks -source "cmd/eksctl-anywhere/cmd/internal/commands/artifacts/import.go"
	${GOPATH}/bin/mockgen -destination=cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks/import_tools_image.go -package=mocks -source "cmd/eksctl-anywhere/cmd/internal/commands/artifacts/import_tools_image.go"
	${GOPATH}/bin/mockgen -destination=pkg/helm/mocks/download.go -package=mocks -source "pkg/helm/download.go"
	${GOPATH}/bin/mockgen -destination=pkg/aws/mocks/ec2.go -package=mocks -source "pkg/aws/ec2.go"
	${GOPATH}/bin/mockgen -destination=pkg/providers/snow/mocks/aws.go -package=mocks -source "pkg/providers/snow/aws.go"
	${GOPATH}/bin/mockgen -destination=pkg/providers/snow/mocks/defaults.go -package=mocks -source "pkg/providers/snow/defaults.go"
	${GOPATH}/bin/mockgen -destination=pkg/providers/snow/mocks/client.go -package=mocks -source "pkg/providers/snow/snow.go"
	${GOPATH}/bin/mockgen -destination=pkg/eksd/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/eksd" EksdInstallerClient
	${GOPATH}/bin/mockgen -destination=pkg/curatedpackages/mocks/kubectlrunner.go -package=mocks -source "pkg/curatedpackages/kubectlrunner.go" KubectlRunner
	${GOPATH}/bin/mockgen -destination=pkg/curatedpackages/mocks/packageinstaller.go -package=mocks -source "pkg/curatedpackages/packageinstaller.go" PackageController PackageHandler
	${GOPATH}/bin/mockgen -destination=pkg/curatedpackages/mocks/reader.go -package=mocks -source "pkg/curatedpackages/bundle.go" Reader BundleRegistry
	${GOPATH}/bin/mockgen -destination=pkg/curatedpackages/mocks/bundlemanager.go -package=mocks -source "pkg/curatedpackages/bundlemanager.go" Manager
	${GOPATH}/bin/mockgen -destination=pkg/clients/kubernetes/mocks/kubectl.go -package=mocks -source "pkg/clients/kubernetes/unauth.go"
	${GOPATH}/bin/mockgen -destination=pkg/clients/kubernetes/mocks/kubeconfig.go -package=mocks -source "pkg/clients/kubernetes/kubeconfig.go"
	${GOPATH}/bin/mockgen -destination=pkg/curatedpackages/mocks/installer.go -package=mocks -source "pkg/curatedpackages/packagecontrollerclient.go" ChartInstaller
	${GOPATH}/bin/mockgen -destination=pkg/cluster/mocks/client_builder.go -package=mocks -source "pkg/cluster/client_builder.go"
	${GOPATH}/bin/mockgen -destination=controllers/mocks/factory.go -package=mocks "github.com/aws/eks-anywhere/controllers" Manager

.PHONY: verify-mocks
verify-mocks: mocks ## Verify if mocks need to be updated
	$(eval DIFF=$(shell git diff --raw -- '*.go' | wc -c))
	if [[ $(DIFF) != 0 ]]; then \
		echo "Detected out of date mocks"; \
		exit 1;\
	fi

.PHONY: e2e-cross-platform
e2e-cross-platform: $(E2E_CROSS_PLATFORMS)

.PHONY: e2e-cross-platform-%
e2e-cross-platform-%: ## Generate binaries for Linux and MacOS
e2e-cross-platform-%: GO_OS = $(firstword $(subst -, ,$*))
e2e-cross-platform-%: GO_ARCH = $(lastword $(subst -, ,$*))
e2e-cross-platform-%:
	$(MAKE) e2e-tests-binary E2E_TAGS=e2e GIT_VERSION=$(DEV_GIT_VERSION) GO_OS=$(GO_OS) GO_ARCH=$(GO_ARCH) E2E_OUTPUT_FILE=bin/$(GO_OS)/$(GO_ARCH)/e2e.test

.PHONY: e2e
e2e: eks-a-e2e integration-test-binary ## Build integration tests
	$(MAKE) e2e-tests-binary E2E_TAGS=e2e

.PHONY: eks-a-e2e
eks-a-e2e:
	if [ "$(CODEBUILD_CI)" = "true" ]; then \
		if [[ "$(CODEBUILD_BUILD_ID)" =~ "aws-staging-eks-a-build" ]]; then \
			make eks-a-release-cross-platform GIT_VERSION=$(shell cat release/triggers/eks-a-release/development/RELEASE_VERSION) RELEASE_MANIFEST_URL=https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml; \
			make eks-a-release GIT_VERSION=$(DEV_GIT_VERSION); \
			scripts/get_bundle.sh; \
		else \
			make check-eksa-components-override; \
			make eks-a-cross-platform; \
			make eks-a; \
		fi \
	else \
		make check-eksa-components-override; \
		make eks-a; \
	fi

.PHONY: e2e-tests-binary
e2e-tests-binary: E2E_OUTPUT_FILE ?= bin/e2e.test
e2e-tests-binary:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) $(GO) test ./test/e2e -c -o bin/e2e.test -tags "$(E2E_TAGS)" -ldflags "-X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(DEV_GIT_VERSION) -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=$(RELEASE_MANIFEST_URL) -X github.com/aws/eks-anywhere/pkg/manifests/releases.manifestURL=$(RELEASE_MANIFEST_URL)"

.PHONY: integration-test-binary
integration-test-binary:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) $(GO) build -o bin/test github.com/aws/eks-anywhere/cmd/integration_test

.PHONY: conformance
conformance:
	$(MAKE) e2e-tests-binary E2E_TAGS=conformance_e2e
	./bin/e2e.test -test.v -test.run 'TestVSphereKubernetes121ThreeWorkersConformanc.*'

.PHONY: conformance-tests
conformance-tests: eks-a-e2e integration-test-binary ## Build e2e conformance tests
	$(MAKE) e2e-tests-binary E2E_TAGS=conformance_e2e

.PHONY: check-eksa-components-override
check-eksa-components-override:
	scripts/eksa_components_override.sh $(BUNDLE_MANIFEST_URL)

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: update-kustomization-yaml
update-kustomization-yaml:
	yq e ".images[] |= select(.name == \"controller\") |= .newName = \"${MANIFEST_IMAGE_NAME_OVERRIDE}\"" -i $(KUSTOMIZATION_CONFIG)
	yq e ".images[] |= select(.name == \"controller\") |= .newTag = \"${MANIFEST_IMAGE_TAG_OVERRIDE}\"" -i $(KUSTOMIZATION_CONFIG)

.PHONY: generate-manifests
generate-manifests: ## Generate manifests e.g. CRD, RBAC etc.
	$(MAKE) generate-core-manifests

.PHONY: generate-core-manifests
generate-core-manifests: $(CONTROLLER_GEN) ## Generate manifests for the core provider e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./pkg/api/... \
		paths=./controllers/... \
		paths=./manager/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook

REGISTRY ?= public.ecr.aws/a2k4d8v8
IMAGE_NAME ?= eksa-cluster-controller
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)

TAG ?= dev
ARCH ?= amd64

CONTROLLER_IMG_TAGGED ?= $(CONTROLLER_IMG)-$(ARCH):$(TAG)

LDFLAGS := $(shell hack/version.sh)

.PHONY: docker-build
docker-build:
	$(MAKE) ARCH=$(ARCH) docker-build-core

.PHONY: docker-build-core
docker-build-core: docker-pull-prerequisites ## Build the docker image for controller-manager
	DOCKER_BUILDKIT=1 docker build --build-arg ARCH=$(ARCH) --build-arg ldflags="$(LDFLAGS)" . -t $(CONTROLLER_IMG_TAGGED) -f build/Dockerfile

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG_TAGGED)

.PHONY: docker-pull-prerequisites
docker-pull-prerequisites:
	docker pull docker.io/docker/dockerfile:1.1-experimental
	docker pull docker.io/library/golang:$(GOLANG_VERSION)
	docker pull gcr.io/distroless/static:latest

## TODO update release folder
RELEASE_DIR := config/manifest
RELEASE_MANIFEST_TARGET ?= eksa-components.yaml

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) generate-manifests $(RELEASE_DIR) $(CONTROLLER_MANIFEST_OUTPUT_DIR) ## Builds the manifests to publish with a release
	# Build core-components.
	$(KUSTOMIZE) build config/prod > $(RELEASE_DIR)/$(RELEASE_MANIFEST_TARGET)
	cp $(RELEASE_DIR)/$(RELEASE_MANIFEST_TARGET) $(CONTROLLER_MANIFEST_OUTPUT_DIR)

.PHONY: run-controller # Run eksa controller from local repo with tilt
run-controller: $(KUSTOMIZE)
	tilt up --file manager/Tiltfile

# go-get-tool will 'go get' any package $2 and install it to $1.
# originally copied from kubebuilder
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
BIN_PATH=$$(realpath $$(dirname $(1))) ;\
PKG_BIN_NAME=$$(echo "$(2)" | sed 's,^.*/\(.*\)@v.*$$,\1,') ;\
BIN_NAME=$$(basename $(1)) ;\
echo "Install dir $$BIN_PATH" ;\
echo "Downloading $(2)" ;\
GOBIN=$$BIN_PATH go install $(2) ;\
[[ $$PKG_BIN_NAME == $$BIN_NAME ]] || mv -f $$BIN_PATH/$$PKG_BIN_NAME $$BIN_PATH/$$BIN_NAME ;\
}
endef

define BUILDCTL
	$(BUILDKIT) \
		build \
		--frontend dockerfile.v0 \
		--opt platform=$(IMAGE_PLATFORMS) \
		--opt build-arg:BASE_IMAGE=$(CLUSTER_CONTROLLER_BASE_IMAGE) \
		--progress plain \
		--local dockerfile=$(DOCKERFILE_FOLDER) \
		--local context=. \
		--output type=$(IMAGE_OUTPUT_TYPE),oci-mediatypes=true,\"name=$(CLUSTER_CONTROLLER_IMAGE),$(CLUSTER_CONTROLLER_LATEST_IMAGE)\",$(IMAGE_OUTPUT) \
		$(if $(filter push=true,$(IMAGE_OUTPUT)),--export-cache type=inline,) \
		$(foreach IMPORT_CACHE,$(IMAGE_IMPORT_CACHE),--import-cache $(IMPORT_CACHE))

endef 


## Fetch Binary Targets
define FULL_FETCH_BINARIES_TARGETS
	$(addprefix $(BINARY_DEPS_DIR)/linux-amd64/, $(1)) $(addprefix $(BINARY_DEPS_DIR)/linux-arm64/, $(1))
endef

$(ORGANIZE_BINARIES_TARGETS): ARTIFACTS_BUCKET = s3://projectbuildpipeline-857-pipelineoutputartifactsb-10ajmk30khe3f
$(ORGANIZE_BINARIES_TARGETS): $(call FULL_FETCH_BINARIES_TARGETS, $(FETCH_BINARIES_TARGETS))
	$(BUILD_LIB)/organize_binaries.sh $(BINARY_DEPS_DIR) $(lastword $(subst -, ,$(@D)))

$(BINARY_DEPS_DIR)/linux-%:
	$(BUILD_LIB)/fetch_binaries.sh $(BINARY_DEPS_DIR) $* $(ARTIFACTS_BUCKET) $(LATEST)

# Do not binary deps as intermediate files
ifneq ($(FETCH_BINARIES_TARGETS),)
.SECONDARY: $(call FULL_FETCH_BINARIES_TARGETS, $(FETCH_BINARIES_TARGETS))
endif
