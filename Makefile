export INTEGRATION_TEST_UBUNTU_AMI_ID?=integration_test_ami
export INTEGRATION_TEST_STORAGE_BUCKET?=integration_test_storage_bucket
export INTEGRATION_TEST_INSTANCE_PROFILE?=integration_test_instance_profile
export INTEGRATION_TEST_MAX_INSTANCE_AGE?=86400
export INTEGRATION_TEST_SUBNET_ID?=integration_test_subnet_id
export INTEGRATION_TEST_INSTANCE_TAG?=integration_test_instance_tag
export JOB_ID?=${PROW_JOB_ID}
GO_TEST ?= go test
ARTIFACTS_BUCKET?=my-s3-bucket
GIT_VERSION?=$(shell git describe --tag)
GIT_TAG?=$(shell git describe --tag | cut -d'-' -f1)
GOLANG_VERSION?="1.16"

RELEASE_MANIFEST_URL?=https://dev-release-prod-pdx.s3.us-west-2.amazonaws.com/eks-a-release.yaml
BUNDLE_MANIFEST_URL?=https://dev-release-prod-pdx.s3.us-west-2.amazonaws.com/bundle-release.yaml
DEV_GIT_VERSION:=v0.0.0-dev

AWS_ACCOUNT_ID?=$(shell aws sts get-caller-identity --query Account --output text)
AWS_REGION=us-west-2

BIN_DIR := bin
TOOLS_BIN_DIR := hack/tools/bin

KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize
KUSTOMIZE_VERSION := 4.2.0

KUBEBUILDER := $(TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_VERSION := v3.1.0

BUILD_LIB := build/lib
BUILDKIT := $(BUILD_LIB)/buildkit.sh

CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)

BINARY_NAME=eks-anywhere-cluster-controller
ifdef CODEBUILD_SRC_DIR
	TAR_PATH?=$(CODEBUILD_SRC_DIR)/$(PROJECT_PATH)/$(CODEBUILD_BUILD_NUMBER)-$(CODEBUILD_RESOLVED_SOURCE_VERSION)/artifacts
else
	TAR_PATH?="_output/tar"
endif

BASE_REPO?=public.ecr.aws/eks-distro-build-tooling
CLUSTER_CONTROLLER_BASE_IMAGE_NAME?=eks-distro-minimal-base
CLUSTER_CONTROLLER_BASE_TAG?=$(shell cat controllers/EKS_DISTRO_MINIMAL_BASE_TAG_FILE)
CLUSTER_CONTROLLER_BASE_IMAGE?=$(BASE_REPO)/$(CLUSTER_CONTROLLER_BASE_IMAGE_NAME):$(CLUSTER_CONTROLLER_BASE_TAG)

IMAGE_REPO=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
IMAGE_TAG?=$(GIT_TAG)-$(shell git rev-parse HEAD)
CLUSTER_CONTROLLER_IMAGE_NAME=eks-anywhere-cluster-controller
CLUSTER_CONTROLLER_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):$(IMAGE_TAG)
CLUSTER_CONTROLLER_LATEST_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):latest

# This removes the compile dependency on C libraries from github.com/containers/storage which is imported by github.com/replicatedhq/troubleshoot
BUILD_TAGS := exclude_graphdriver_btrfs exclude_graphdriver_devicemapper

GO_ARCH:=$(shell go env GOARCH)
GO_OS:=$(shell go env GOOS)

DOCKER_E2E_TEST := TestDockerKubernetes121SimpleFlow

.PHONY: default
default: build lint

.PHONY: build
build: eks-a eks-a-tool unit-test ## Generate binaries and run unit tests

.PHONY: release
release: eks-a-release unit-test ## Generate release binary and run unit tests

.PHONY: eks-a-binary
eks-a-binary: ALL_LINKER_FLAGS := $(LINKER_FLAGS) -X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(GIT_VERSION) -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=$(RELEASE_MANIFEST_URL)
eks-a-binary: LINKER_FLAGS_ARG := -ldflags "$(ALL_LINKER_FLAGS)"
eks-a-binary: BUILD_TAGS_ARG := -tags "$(BUILD_TAGS)"
eks-a-binary: OUTPUT_FILE ?= bin/eksctl-anywhere
eks-a-binary:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build $(BUILD_TAGS_ARG) $(LINKER_FLAGS_ARG) -o $(OUTPUT_FILE) github.com/aws/eks-anywhere/cmd/eksctl-anywhere

.PHONY: eks-a-embed-config
eks-a-embed-config: ## Build a dev release version of eks-a with embed cluster spec config
	$(MAKE) eks-a-binary GIT_VERSION=$(DEV_GIT_VERSION) RELEASE_MANIFEST_URL=embed:///config/releases.yaml BUILD_TAGS='$(BUILD_TAGS) spec_embed_config'

.PHONY: eks-a-cross-platform-embed-latest-config
eks-a-cross-platform-embed-latest-config: ## Build cross platform dev release versions of eks-a with the latest bundle-release.yaml embedded in cluster spec config
	curl -L $(BUNDLE_MANIFEST_URL) --output pkg/cluster/config/bundle-release.yaml
	$(MAKE) eks-a-embed-config GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/eksctl-anywhere
	$(MAKE) eks-a-embed-config GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/eksctl-anywhere
	rm pkg/cluster/config/bundle-release.yaml

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
	$(MAKE) eks-a-binary GIT_VERSION=$(GIT_VERSION) GO_OS=darwin GO_ARCH=amd64 OUTPUT_FILE=bin/darwin/eksctl-anywhere LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'
	$(MAKE) eks-a-binary GIT_VERSION=$(GIT_VERSION) GO_OS=linux GO_ARCH=amd64 OUTPUT_FILE=bin/linux/eksctl-anywhere LINKER_FLAGS='-s -w -X github.com/aws/eks-anywhere/pkg/eksctl.enabled=true'

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(KUSTOMIZE): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR) && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s $(KUSTOMIZE_VERSION)

$(KUBEBUILDER): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR) && curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/$(KUBEBUILDER_VERSION)/kubebuilder_$(GO_OS)_$(GO_ARCH)
	chmod +x $(KUBEBUILDER)

$(CONTROLLER_GEN): $(TOOLS_BIN_DIR)
	cd $(TOOLS_BIN_DIR); go build -tags=tools -o $(CONTROLLER_GEN_BIN) sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: lint
lint: bin/golangci-lint ## Run golangci-lint
	bin/golangci-lint run

bin/golangci-lint: ## Download golangci-lint
bin/golangci-lint: GOLANGCI_LINT_VERSION?=$(shell cat .github/workflows/golangci-lint.yml | sed -n -e 's/^\s*version: //p')
bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

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
	controllers/build/upload_artifacts.sh $(TAR_PATH) $(ARTIFACTS_BUCKET) $(PROJECT_PATH) $(CODEBUILD_BUILD_NUMBER) $(CODEBUILD_RESOLVED_SOURCE_VERSION)

.PHONY: cluster-controller-binaries
cluster-controller-binaries:
	controllers/build/create_binaries.sh $(GOLANG_VERSION) $(BINARY_NAME) $(IMAGE_REPO) $(IMAGE_TAG)

.PHONY: cluster-controller-tarballs
cluster-controller-tarballs:  cluster-controller-binaries
	controllers/build/create_tarballs.sh $(BINARY_NAME) $(GIT_TAG) $(TAR_PATH)

.PHONY: cluster-controller-local-images
cluster-controller-local-images: cluster-controller-binaries
	$(BUILDKIT) \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(CLUSTER_CONTROLLER_BASE_IMAGE) \
		--local dockerfile=./controllers/docker/linux/eks-anywhere-cluster-controller \
		--local context=. \
		--output type=oci,oci-mediatypes=true,\"name=$(CLUSTER_CONTROLLER_IMAGE),$(CLUSTER_CONTROLLER_LATEST_IMAGE)\",dest=/tmp/eks-anywhere-cluster-controller.tar

.PHONY: cluster-controller-images
cluster-controller-images: cluster-controller-binaries
	$(BUILDKIT) \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(CLUSTER_CONTROLLER_BASE_IMAGE) \
		--local dockerfile=./controllers/docker/linux/eks-anywhere-cluster-controller \
		--local context=. \
		--output type=image,oci-mediatypes=true,\"name=$(CLUSTER_CONTROLLER_IMAGE),$(CLUSTER_CONTROLLER_LATEST_IMAGE)\",push=true

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

.PHONY: capd-test-all
capd-test-all: capd-test capd-test-120

.PHONY: capd-test-%
capd-test-%: e2e ## Run CAPD tests
	./bin/e2e.test -test.v -test.run TestDockerKubernetes$*SimpleFlow

.PHONY: mocks
mocks: ## Generate mocks
	go install github.com/golang/mock/mockgen@v1.5.0
	${GOPATH}/bin/mockgen -destination=pkg/providers/mocks/providers.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers" Provider,DatacenterConfig,MachineConfig
	${GOPATH}/bin/mockgen -destination=pkg/executables/mocks/executables.go -package=mocks "github.com/aws/eks-anywhere/pkg/executables" Executable
	${GOPATH}/bin/mockgen -destination=pkg/providers/docker/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/docker" ProviderClient,ProviderKubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/vsphere" ProviderGovcClient,ProviderKubectlClient,ClusterResourceSetManager
	${GOPATH}/bin/mockgen -destination=pkg/filewriter/mocks/filewriter.go -package=mocks "github.com/aws/eks-anywhere/pkg/filewriter" FileWriter
	${GOPATH}/bin/mockgen -destination=pkg/clustermanager/mocks/client_and_networking.go -package=mocks "github.com/aws/eks-anywhere/pkg/clustermanager" ClusterClient,Networking,AwsIamAuth
	${GOPATH}/bin/mockgen -destination=pkg/addonmanager/addonclients/mocks/fluxaddonclient.go -package=mocks "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients" Flux
	${GOPATH}/bin/mockgen -destination=pkg/task/mocks/task.go -package=mocks "github.com/aws/eks-anywhere/pkg/task" Task
	${GOPATH}/bin/mockgen -destination=pkg/bootstrapper/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/bootstrapper" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/cluster/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/cluster" ClusterClient
	${GOPATH}/bin/mockgen -destination=pkg/workflows/interfaces/mocks/clients.go -package=mocks "github.com/aws/eks-anywhere/pkg/workflows/interfaces" Bootstrapper,ClusterManager,AddonManager,Validator,CAPIManager
	${GOPATH}/bin/mockgen -destination=pkg/git/providers/github/mocks/github.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/providers/github" GitProviderClient,GithubProviderClient
	${GOPATH}/bin/mockgen -destination=pkg/git/mocks/git.go -package=mocks "github.com/aws/eks-anywhere/pkg/git" Provider
	${GOPATH}/bin/mockgen -destination=pkg/git/gogithub/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogithub" Client
	${GOPATH}/bin/mockgen -destination=pkg/git/gogit/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogit" GoGitClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/docker.go -package=mocks "github.com/aws/eks-anywhere/pkg/validations" DockerExecutable
	${GOPATH}/bin/mockgen -destination=controllers/controllers/resource/mocks/resource.go -package=mocks "github.com/aws/eks-anywhere/controllers/controllers/resource" ResourceFetcher,ResourceUpdater
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/templates/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/templates/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/tags/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/tags/factory.go" GovcClient
	${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/kubectl.go -package=mocks -source "pkg/validations/kubectl.go" KubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/diagnostics/interfaces/mocks/diagnostics.go -package=mocks -source "pkg/diagnostics/interfaces.go" DiagnosticBundle,AnalyzerFactory,CollectorFactory,BundleClient
	${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/capiclient.go -package=mocks -source "pkg/clusterapi/manager.go" CAPIClient,KubectlClient
	${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/client.go -package=mocks -source "pkg/clusterapi/resourceset_manager.go" Client
	${GOPATH}/bin/mockgen -destination=pkg/crypto/mocks/crypto.go -package=mocks -source "pkg/crypto/certificategen.go" CertificateGenerator

.PHONY: verify-mocks
verify-mocks: mocks ## Verify if mocks need to be updated
	$(eval DIFF=$(shell git diff --raw -- '*.go' | wc -c))
	if [[ $(DIFF) != 0 ]]; then \
		echo "Detected out of date mocks"; \
		exit 1;\
	fi

.PHONY: e2e
e2e: eks-a-e2e integration-test-binary ## Build integration tests
	$(MAKE) e2e-tests-binary E2E_TAGS=e2e

.PHONY: conformance
conformance:
	$(MAKE) e2e-tests-binary E2E_TAGS=conformance_e2e
	./bin/e2e.test -test.v -test.run 'TestVSphereKubernetes121ThreeWorkersConformanc.*'

.PHONY: conformance-tests
conformance-tests: eks-a-e2e integration-test-binary ## Build e2e conformance tests
	$(MAKE) e2e-tests-binary E2E_TAGS=conformance_e2e

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
e2e-tests-binary:
	go test ./test/e2e -c -o bin/e2e.test -tags "$(E2E_TAGS)" -ldflags "-X github.com/aws/eks-anywhere/pkg/version.gitVersion=$(DEV_GIT_VERSION) -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=$(RELEASE_MANIFEST_URL)"

.PHONY: integration-test-binary
integration-test-binary:
	go build -o bin/test github.com/aws/eks-anywhere/cmd/integration_test

.PHONY: check-eksa-components-override
check-eksa-components-override:
	scripts/eksa_components_override.sh $(BUNDLE_MANIFEST_URL)

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
RELEASE_MANIFEST_TARGET ?= eksa-components.yaml

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) generate-manifests $(RELEASE_DIR) ## Builds the manifests to publish with a release
	# Build core-components.
	$(KUSTOMIZE) build config/prod > $(RELEASE_DIR)/$(RELEASE_MANIFEST_TARGET)

.PHONY: run-controller # Run eksa controller from local repo with tilt
run-controller:
	tilt up --file controllers/Tiltfile
