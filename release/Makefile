REPO_ROOT=$(shell git rev-parse --show-toplevel)

AWS_ACCOUNT_ID?=$(shell aws sts get-caller-identity --query Account --output text)
AWS_REGION?=us-west-2

RELEASE_ENVIRONMENT?=development
RELEASE_VERSION?=$(shell cat triggers/eks-a-release/$(RELEASE_ENVIRONMENT)/RELEASE_VERSION)
BUNDLE_NUMBER?=$(shell cat triggers/bundle-release/$(RELEASE_ENVIRONMENT)/BUNDLE_NUMBER)
RELEASE_NUMBER?=$(shell cat triggers/eks-a-release/$(RELEASE_ENVIRONMENT)/RELEASE_NUMBER)
CLI_MIN_VERSION?=$(shell cat triggers/bundle-release/$(RELEASE_ENVIRONMENT)/CLI_MIN_VERSION)
CLI_MAX_VERSION?=$(shell cat triggers/bundle-release/$(RELEASE_ENVIRONMENT)/CLI_MAX_VERSION)
SOURCE_BUCKET?=source-bucket
RELEASE_BUCKET?=release-bucket
SOURCE_CONTAINER_REGISTRY?=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
RELEASE_CONTAINER_REGISTRY?=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
CDN?=https://$(RELEASE_BUCKET)
CLI_REPO_URL?=https://github.com/aws/eks-anywhere.git
BUILD_REPO_URL?=https://github.com/aws/eks-anywhere-build-tooling.git
BUILD_REPO_BRANCH_NAME?=main
CLI_REPO_BRANCH_NAME?=main
ifeq ($(CI),true)
BUILD_REPO_BRANCH_NAME=$(PULL_BASE_REF)
CLI_REPO_BRANCH_NAME=$(PULL_BASE_REF)
endif
ifeq ($(CODEBUILD_CI),)
DRY_RUN=true
SOURCE_BUCKET=projectbuildpipeline-857-pipelineoutputartifactsb-10ajmk30khe3f
else
DRY_RUN=false
endif
DATE_YYYYMMDD=$(shell date "+%F")
WEEKLY?=false
WEEKLY_RELEASES_URL_PREFIX=https://dev-release-assets.eks-anywhere.model-rocket.aws.dev/weekly-releases

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setup Go
GOLANG_VERSION?="1.21"
GO_VERSION ?= $(shell source $(REPO_ROOT)/scripts/common.sh && build::common::get_go_path $(GOLANG_VERSION))
GO ?= $(GO_VERSION)/go

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

TOOLS_BIN_DIR := ../hack/tools/bin
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

clean: ## Cleanup resources created by make targets
	rm -rf bin $(HOME)/eks-a-source downloaded-artifacts

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	## Copy only the bundles crd to root config kustomization folder
	cp -f config/crd/bases/anywhere.eks.amazonaws.com_bundles.yaml ../config/crd/bases/
	cp -f config/crd/bases/anywhere.eks.amazonaws.com_eksareleases.yaml ../config/crd/bases/
	@echo "Make sure to update the test bundle with \`make update-bundle-golden-files\`"

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	$(GO) fmt ./...

vet: ## Run go vet against code.
	$(GO) vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.2/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); $(GO) test ./... -coverprofile cover.out

##@ Build

build: clean fmt vet ## Build release binary.
	$(GO) build -C cli -o $(REPO_ROOT)/release/bin/eks-anywhere-release github.com/aws/eks-anywhere/release/cli

.PHONY: lint
lint: bin/golangci-lint ## Run golangci-lint
	bin/golangci-lint run

bin/golangci-lint: ## Download golangci-lint
bin/golangci-lint: GOLANGCI_LINT_VERSION?=$(shell cat $(REPO_ROOT)/.github/workflows/golangci-lint.yml | sed -n -e 's/^\s*version: //p')
bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)

unit-test: ## Run go test against code.
	$(GO) test -count=1 ./...

update-bundle-golden-files: ## Updates testdata files located under pkg/test/testdata
	$(GO) test -C cli -count=1 ./pkg/operations -update
	$(eval DIFF_LINE_COUNT=$(shell git diff cli/pkg/test/testdata | wc -l))
	@if [[ $(DIFF_LINE_COUNT) != 0 ]]; then \
		printf "\n\033[33mWarning:\033[0m Testdata files have been updated! Ensure that these changes were intentional.\n"; \
	fi

run: manifests generate fmt vet ## Run a controller from your host.
	$(GO) run ./main.go

##@ Release

dev-release: build ## Perform a dev release of EKS-A bundles and CLI manifests
	scripts/release.sh $(REPO_ROOT)/release/downloaded-artifacts $(SOURCE_BUCKET) $(RELEASE_BUCKET) $(CDN) $(SOURCE_CONTAINER_REGISTRY) $(RELEASE_CONTAINER_REGISTRY) $(BUILD_REPO_URL) $(CLI_REPO_URL) $(BUILD_REPO_BRANCH_NAME) $(CLI_REPO_BRANCH_NAME) $(DRY_RUN) $(WEEKLY)

bundle-release: build ## Perform EKS-A versioned bundles release
	scripts/bundle-release.sh $(REPO_ROOT)/release/downloaded-artifacts $(SOURCE_BUCKET) $(RELEASE_BUCKET) $(CDN) $(BUNDLE_NUMBER) $(CLI_MIN_VERSION) $(CLI_MAX_VERSION) $(SOURCE_CONTAINER_REGISTRY) $(RELEASE_CONTAINER_REGISTRY) $(RELEASE_ENVIRONMENT) $(BUILD_REPO_BRANCH_NAME) $(CLI_REPO_BRANCH_NAME) $(BUILD_REPO_URL) $(CLI_REPO_URL)

eks-a-release: build ## Perform EKS-A CLI release
	scripts/eks-a-release.sh $(RELEASE_VERSION) $(REPO_ROOT)/release/downloaded-artifacts $(SOURCE_BUCKET) $(RELEASE_BUCKET) $(CDN) $(BUNDLE_NUMBER) $(RELEASE_NUMBER) $(RELEASE_ENVIRONMENT) $(CLI_REPO_BRANCH_NAME) $(BUILD_REPO_URL) $(CLI_REPO_URL)

github-bundle-release:
	scripts/github-bundle-release.sh $(REPO_ROOT)/release/downloaded-artifacts $(WEEKLY_RELEASES_URL_PREFIX)

announce-eks-a-release: EKSA_RELEASE_MANIFEST_URL?=https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml
announce-eks-a-release:
	scripts/announce_eksa_release.sh $(EKSA_RELEASE_MANIFEST_URL)

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
