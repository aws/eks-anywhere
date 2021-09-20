BASE_DIRECTORY=$(shell git rev-parse --show-toplevel)
BUILD_LIB=$(BASE_DIRECTORY)/build/lib
ARTIFACTS_BUCKET?=my-s3-bucket
GIT_TAG=$(shell git describe --tag)
GOLANG_VERSION?="1.16"

BINARY_NAME=eks-anywhere-cluster-controller
ifdef CODEBUILD_SRC_DIR
	TAR_PATH?=$(CODEBUILD_SRC_DIR)/$(PROJECT_PATH)/$(CODEBUILD_BUILD_NUMBER)-$(CODEBUILD_RESOLVED_SOURCE_VERSION)/artifacts
else
	TAR_PATH?="_output/tar"
endif

GOPROXY_DNS?=https://proxy.golang.org
export GOPROXY=$(GOPROXY_DNS)

AWS_ACCOUNT_ID?=$(shell aws sts get-caller-identity --query Account --output text)
AWS_REGION=us-west-2

MAKE_ROOT=$(shell cd "$(shell dirname "${BASH_SOURCE[0]}")" && pwd -P)
BASE_REPO?=public.ecr.aws/eks-distro-build-tooling
CLUSTER_CONTROLLER_BASE_IMAGE_NAME?=eks-distro-minimal-base
CLUSTER_CONTROLLER_BASE_TAG?=$(shell cat $(MAKE_ROOT)/EKS_DISTRO_MINIMAL_BASE_TAG_FILE)
CLUSTER_CONTROLLER_BASE_IMAGE?=$(BASE_REPO)/$(CLUSTER_CONTROLLER_BASE_IMAGE_NAME):$(CLUSTER_CONTROLLER_BASE_TAG)
COLLECTOR_BASE_IMAGE_NAME?=eks-distro-base
COLLECTOR_BASE_TAG?=$(shell cat $(MAKE_ROOT)/EKS_DISTRO_BASE_TAG_FILE)
COLLECTOR_BASE_IMAGE=$(BASE_REPO)/$(COLLECTOR_BASE_IMAGE_NAME):$(COLLECTOR_BASE_TAG)

IMAGE_REPO=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
IMAGE_TAG?=$(GIT_TAG)-$(shell git rev-parse HEAD)
CLUSTER_CONTROLLER_IMAGE_NAME=eks-anywhere-cluster-controller
CLUSTER_CONTROLLER_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):$(IMAGE_TAG)
CLUSTER_CONTROLLER_LATEST_IMAGE=$(IMAGE_REPO)/$(CLUSTER_CONTROLLER_IMAGE_NAME):latest
COLLECTOR_IMAGE_NAME=eks-anywhere-diagnostic-collector
COLLECTOR_IMAGE=$(IMAGE_REPO)/$(COLLECTOR_IMAGE_NAME):$(IMAGE_TAG)
COLLECTOR_LATEST_IMAGE=$(IMAGE_REPO)/$(COLLECTOR_IMAGE_NAME):latest

.PHONY: upload-artifacts
upload-artifacts:
	build/upload_artifacts.sh $(TAR_PATH) $(ARTIFACTS_BUCKET) $(PROJECT_PATH) $(CODEBUILD_BUILD_NUMBER) $(CODEBUILD_RESOLVED_SOURCE_VERSION)

.PHONY: binaries
binaries:
	build/create_binaries.sh $(GOLANG_VERSION) $(BINARY_NAME) $(IMAGE_REPO) $(IMAGE_TAG)

.PHONY: generate-attribution
generate-attribution:
	build/create_attribution.sh $(GOLANG_VERSION)

.PHONY: tarballs
tarballs:  binaries
	build/create_tarballs.sh $(BINARY_NAME) $(GIT_TAG) $(TAR_PATH)

.PHONY: local-images
local-images: binaries
	$(BUILD_LIB)/buildkit.sh \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(CLUSTER_CONTROLLER_BASE_IMAGE) \
		--local dockerfile=./docker/linux/eks-anywhere-cluster-controller \
		--local context=. \
		--output type=oci,oci-mediatypes=true,\"name=$(CLUSTER_CONTROLLER_IMAGE),$(CLUSTER_CONTROLLER_LATEST_IMAGE)\",dest=/tmp/eks-anywhere-cluster-controller.tar
	$(BUILD_LIB)/buildkit.sh \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(COLLECTOR_BASE_IMAGE) \
		--local dockerfile=./docker/linux/eksa-diagnostic-collector \
		--local context=. \
		--output type=oci,oci-mediatypes=true,\"name=$(COLLECTOR_IMAGE),$(COLLECTOR_LATEST_IMAGE)\",dest=/tmp/eksa-diagnostic-collector.tar

.PHONY: images
images: binaries
	$(BUILD_LIB)/buildkit.sh \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(CLUSTER_CONTROLLER_BASE_IMAGE) \
		--local dockerfile=./docker/linux/eks-anywhere-cluster-controller \
		--local context=. \
		--output type=image,oci-mediatypes=true,\"name=$(CLUSTER_CONTROLLER_IMAGE),$(CLUSTER_CONTROLLER_LATEST_IMAGE)\",push=true
	$(BUILD_LIB)/buildkit.sh \
		build \
		--frontend dockerfile.v0 \
		--opt platform=linux/amd64 \
		--opt build-arg:BASE_IMAGE=$(COLLECTOR_BASE_IMAGE) \
		--local dockerfile=./docker/linux/eksa-diagnostic-collector \
		--local context=. \
		--output type=image,oci-mediatypes=true,\"name=$(COLLECTOR_IMAGE),$(COLLECTOR_LATEST_IMAGE)\",push=true

.PHONY: build
build: local-images tarballs
	echo "Done"

.PHONY: release
release: images tarballs
	echo "Done"

.PHONY: release-upload
release-upload: release upload-artifacts
	echo "Done upload"

.PHONY: all
all: release

.PHONY: clean
clean:
	rm -rf "_output"
