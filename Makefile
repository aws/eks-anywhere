export INTEGRATION_TEST_UBUNTU_AMI_ID?=integration_test_ami
export INTEGRATION_TEST_STORAGE_BUCKET?=integration_test_storage_bucket
export INTEGRATION_TEST_INSTANCE_PROFILE?=integration_test_instance_profile
export INTEGRATION_TEST_MAX_INSTANCE_AGE?=86400
export INTEGRATION_TEST_SUBNET_ID?=integration_test_subnet_id
export INTEGRATION_TEST_INSTANCE_TAG?=integration_test_instance_tag
export JOB_ID?=${PROW_JOB_ID}
GO_TEST ?= go test
GIT_VERSION?=$(shell git describe --tag)

RELEASE_MANIFEST_URL?=https://dev-release-prod-pdx.s3.us-west-2.amazonaws.com/eks-a-release.yaml
DEV_GIT_VERSION:=v0.0.0-dev

BIN_DIR := bin
TOOLS_BIN_DIR := hack/tools/bin

KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize
KUSTOMIZE_VERSION := 4.2.0

KUBEBUILDER := $(TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_VERSION := v3.1.0

CONTROLLER_GEN_BIN := contoller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)

# This removes the compile dependency on C libraries from github.com/containers/storage which is imported by github.com/replicatedhq/troubleshoot
BUILD_TAGS := exclude_graphdriver_btrfs exclude_graphdriver_devicemapper

GO_ARCH:=$(shell go env GOARCH)
GO_OS:=$(shell go env GOOS)

DOCKER_E2E_TEST := TestDockerKubernetes121SimpleFlow 

.PHONY: build
build: eks-a eks-a-tool lint unit-test ## Generate binaries, run go lint and unit tests

.PHONY: release
release: eks-a-release unit-test ## Generate release binary and run unit tests

.PHONY: eks-a-binary
eks-a-binary: ALL_LINKER_FLAGS := $(LINKER_FLAGS) -X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(GIT_VERSION) -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=$(RELEASE_MANIFEST_URL)
eks-a-binary: LINKER_FLAGS_ARG := -ldflags "$(ALL_LINKER_FLAGS)"
eks-a-binary: BUILD_TAGS_ARG := -tags "$(BUILD_TAGS)"
eks-a-binary: OUTPUT_FILE ?= bin/eksctl-anywhere
eks-a-binary: 
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build $(BUILD_TAGS_ARG) $(LINKER_FLAGS_ARG) -o $(OUTPUT_FILE) github.com/aws/eks-anywhere/cmd/eks-a

.PHONY: eks-a-embed-config
eks-a-embed-config: ## Build a dev release version of eks-a with embed cluster spec config
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) RELEASE_MANIFEST_URL=embed:///config/releases.yaml BUILD_TAGS='$(BUILD_TAGS) spec_embed_config'

.PHONY: eks-a
eks-a: ## Build a dev release version of eks-a
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION)

.PHONY: eks-a-release
eks-a-release: ## Generate a release binary
	$(MAKE) eks-a-binary GO_OS=linux GO_ARCH=amd64 LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'

.PHONY: eks-a-cross-platform
eks-a-cross-platform: ## Generate binaries for Linux and MacOS
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/eksctl-anywhere
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/eksctl-anywhere

.PHONY: eks-a-release-cross-platform
eks-a-release-cross-platform: ## Generate binaries for Linux and MacOS
	$(MAKE) eks-a-binary GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/eksctl-anywhere LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'
	$(MAKE) eks-a-binary GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/eksctl-anywhere LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(KUSTOMIZE): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR) && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | sh -s $(KUSTOMIZE_VERSION)

$(KUBEBUILDER): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR) && curl -L -o kubebuilder https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(GO_OS)/$(GO_ARCH)
	chmod +x $(KUBEBUILDER)

$(CONTROLLER_GEN): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR); go build -tags=tools -o $(CONTROLLER_GEN_BIN) sigs.k8s.io/controller-tools/cmd/controller-gen 

.PHONY: lint
lint: bin/golangci-lint ## Run golangci-lint
	bin/golangci-lint run

bin/golangci-lint: ## Download golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.39.0

.PHONY: build-cross-platform
build-cross-platform: eks-a-cross-platform

.PHONY: eks-a-tool
eks-a-tool: ## Build eks-a-tool
	go build -o bin/eks-a-tool github.com/aws/eks-anywhere/cmd/eks-a-tool

.PHONY: eks-a-cluster-controller
eks-a-cluster-controller: ## Build eks-a-cluster-controller
	go build -ldflags "-s -w -buildid='' -extldflags -static" -o bin/manager ./controllers

# This target will copy LICENSE file from root to the release submodule
# when fetching licenses for cluster-controller
.PHONY: copy-license-cluster-controller
copy-license-cluster-controller: SHELL := /bin/bash
copy-license-cluster-controller:
	source scripts/attribution_helpers.sh && build::fix_licenses

# Build target invoked by build-tooling repo script
.PHONY: build-cluster-controller
build-cluster-controller: eks-a-cluster-controller copy-license-cluster-controller

.PHONY: generate-attribution
generate-attribution: GOLANG_VERSION ?= "1.16"
generate-attribution:
	scripts/make_attribution.sh $(GOLANG_VERSION)

.PHONY: update-attribution-files
update-attribution-files: generate-attribution
	scripts/create_pr.sh

.PHONY: clean
clean: ## Clean up resources created by make targets
	rm -rf ./bin/*
	rm -rf ./pkg/executables/cluster-name/
	rm -rf ./pkg/providers/vsphere/test/
	find . -depth -name 'folderWriter*' -exec rm -rf {} \;
	rm -rf ./controllers/bin/*
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
	$(GO_TEST) ./... -cover -tags "$(BUILD_TAGS)"

.PHONY: capd-test
capd-test: e2e ## Run default e2e capd test locally
	./bin/e2e.test -test.v -test.run $(DOCKER_E2E_TEST)

.PHONY: docker-e2e-test
docker-e2e-test: e2e ## Run docker integration test in new ec2 instance
	scripts/e2e_test_docker.sh $(DOCKER_E2E_TEST)

.PHONY: e2e-cleanup
e2e-cleanup: e2e ## Clean up resources generated by e2e tests
	scripts/e2e_cleanup.sh

.PHONY: vsphere-e2e-test
vsphere-e2e-test: build-cross-platform ## Run vsphere integration tests
	go run cmd/integration_test/main.go e2e vsphere \
		-a ${INTEGRATION_TEST_UBUNTU_AMI_ID} \
		-s ${INTEGRATION_TEST_STORAGE_BUCKET} \
		-j ${JOB_ID} \
		-i ${INTEGRATION_TEST_INSTANCE_PROFILE} \
		-n ${INTEGRATION_TEST_SUBNET_ID} \
		-d false

.PHONY: build-e2e-test-binaries
build-e2e-test-binaries: build build-cross-platform ## Build e2e binaries
	go build -o bin/eks-a-e2e github.com/aws/eks-anywhere/cmd/integration_test

.PHONY: capd-test-all
capd-test-all: capd-test-118 capd-test capd-test-120

.PHONY: capd-test-%
capd-test-%: e2e ## Run CAPD tests
	./bin/e2e.test -test.v -test.run TestDockerKubernetes$*SimpleFlow

.PHONY: mocks
mocks: ## Generate mocks
	go get github.com/golang/mock/mockgen@v1.5.0
	${GOPATH}/bin/mockgen -destination=pkg/providers/mocks/providers.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers" Provider,DatacenterConfig,MachineConfig
	${GOPATH}/bin/mockgen -destination=pkg/executables/mocks/executables.go -package=mocks "github.com/aws/eks-anywhere/pkg/executables" Executable
	${GOPATH}/bin/mockgen -destination=pkg/providers/aws/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/aws" ProviderClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/docker/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/docker" ProviderClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/vsphere" ProviderGovcClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/filewriter/mocks/filewriter.go -package=mocks "github.com/aws/eks-anywhere/pkg/filewriter" FileWriter
	${GOPATH}/bin/mockgen -destination=pkg/clustermanager/mocks/client_and_networking.go -package=mocks "github.com/aws/eks-anywhere/pkg/clustermanager" ClusterClient,Networking
	${GOPATH}/bin/mockgen -destination=pkg/addonmanager/addonclients/mocks/fluxaddonclient.go -package=mocks "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients" Flux
	${GOPATH}/bin/mockgen -destination=pkg/task/mocks/task.go -package=mocks "github.com/aws/eks-anywhere/pkg/task" Task
	${GOPATH}/bin/mockgen -destination=pkg/bootstrapper/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/bootstrapper" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/cluster/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/cluster" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/workflows/interfaces/mocks/clients.go -package=mocks "github.com/aws/eks-anywhere/pkg/workflows/interfaces" Bootstrapper,ClusterManager,AddonManager,Validator
	${GOPATH}/bin/mockgen -destination=pkg/git/providers/github/mocks/github.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/providers/github" GitProviderClient,GithubProviderClient
	${GOPATH}/bin/mockgen -destination=pkg/git/mocks/git.go -package=mocks "github.com/aws/eks-anywhere/pkg/git" Provider
	${GOPATH}/bin/mockgen -destination=pkg/git/gogithub/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogithub" Client
	${GOPATH}/bin/mockgen -destination=pkg/git/gogit/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogit" GoGitClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/docker.go -package=mocks "github.com/aws/eks-anywhere/pkg/validations" DockerExecutable
	${GOPATH}/bin/mockgen -destination=controllers/controllers/resource/mocks/resource.go -package=mocks "github.com/aws/eks-anywhere/controllers/controllers/resource" ResourceFetcher,ResourceUpdater
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/templates/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/templates/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/tags/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/tags/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/upgradevalidations/mocks/upgradevalidations.go -package=mocks -source "pkg/validations/upgradevalidations/upgradevalidations.go" ValidationsKubectlClient

.PHONY: verify-mocks
verify-mocks: mocks ## Verify if mocks need to be updated
	$(eval DIFF=$(shell git diff --raw -- '*.go' | wc -c))
	if [[ $(DIFF) != 0 ]]; then \
		echo "Detected out of date mocks"; \
		exit 1;\
	fi

.PHONY: e2e
e2e: eks-a-release-cross-platform integration-test-binary ## Build integration tests
	$(MAKE) e2e-tests-binary E2E_TAGS=e2e

.PHONY: conformance-tests
conformance-tests: eks-a-release-cross-platform integration-test-binary ## Build e2e conformance tests
	$(MAKE) e2e-tests-binary E2E_TAGS=conformance_e2e

.PHONY: e2e-tests-binary
e2e-tests-binary:
	go test ./test/e2e -c -o bin/e2e.test -tags "$(E2E_TAGS)" -ldflags "-X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(DEV_GIT_VERSION)"

.PHONY: integration-test-binary
integration-test-binary:
	go build -o bin/test github.com/aws/eks-anywhere/cmd/integration_test

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate-manifests
generate-manifests: ## Generate manifests e.g. CRD, RBAC etc.
	$(MAKE) generate-core-manifests

.PHONY: generate-core-manifests
generate-core-manifests: $(CONTROLLER_GEN) ## Generate manifests for the core provider e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./pkg/api/... \
		paths=./controllers/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook

GOLANG_VERSION := 1.16.2
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
	DOCKER_BUILDKIT=1 docker build --build-arg ARCH=$(ARCH) --build-arg ldflags="$(LDFLAGS)" . -t CONTROLLER_IMG_TAGGED -f build/Dockerfile

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push CONTROLLER_IMG_TAGGED

.PHONY: docker-pull-prerequisites
docker-pull-prerequisites:
	docker pull docker.io/docker/dockerfile:1.1-experimental
	docker pull docker.io/library/golang:$(GOLANG_VERSION)
	docker pull gcr.io/distroless/static:latest

## TODO update release folder
RELEASE_DIR := config/manifest

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) generate-manifests $(RELEASE_DIR) ## Builds the manifests to publish with a release
	# Build core-components.
	$(KUSTOMIZE) build config/prod > $(RELEASE_DIR)/eksa-components.yaml

.PHONY: e2e-cleanup-postsubmit # Temporary target to make postsubmit cleanup job work.
e2e-cleanup-postsubmit:
	go run cmd/integration_test/main.go e2e cleanup \
		-s ${INTEGRATION_TEST_STORAGE_BUCKET} \
		-a ${INTEGRATION_TEST_MAX_INSTANCE_AGE} \
		-t ${INTEGRATION_TEST_INSTANCE_TAG} \
		-d false

.PHONY: run-controller # Run eksa controller from local repo with tilt
run-controller:
	tilt up --file controllers/Tiltfile	
