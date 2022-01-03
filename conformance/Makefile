SHELL := /bin/bash
GIT_TAG?=$(shell git describe --tag | cut -d'-' -f1)
GITHUB_USER?=$(shell git config --get remote.origin.url|sed -e 's/.*://' -e 's,/.*,,')
SONOBUOY_OSARCH?=darwin_amd64
TARGET_RELEASE?=1.21
TARGET_DIR=k8s-conformance/v$(TARGET_RELEASE)/eks-a
K8S_VERSION=v1.21.5
RESULTS_DIR=./results

.PHONY: default
default: k8s-conformance sonobuoy-download sonobuoy-run sonobuoy-results generate

k8s-conformance:
	git clone git@github.com:$(GITHUB_USER)/k8s-conformance.git
	cd k8s-conformance ; git remote add upstream https://github.com/cncf/k8s-conformance.git
	cd k8s-conformance ; git fetch upstream
	cd k8s-conformance ; git checkout master
	cd k8s-conformance ; git reset --hard upstream/master
	cd k8s-conformance ; git branch -D v$(TARGET_RELEASE) || true
	cd k8s-conformance ; git checkout -b v$(TARGET_RELEASE)

sonobuoy-download: SONOBUOY=https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.50.0/sonobuoy_0.50.0_$(SONOBUOY_OSARCH).tar.gz
sonobuoy-download:
	wget -qO- ${SONOBUOY} |tar -xz sonobuoy
	chmod 755 sonobuoy

sonobuoy-run:
	./sonobuoy run --mode=certified-conformance --wait --kube-conformance-image k8s.gcr.io/conformance:$(K8S_VERSION)
	build/lib/get_results.sh $(TARGET_DIR)

sonobuoy-results:
	build/lib/get_results.sh $(TARGET_DIR)

.PHONY: generate
generate:
	mkdir -p $(TARGET_DIR)
	rm -f $(TARGET_DIR)/*
	sed -e "s/{{k8s_version}}/$(K8S_VERSION)/" <template/PRODUCT.yaml >$(TARGET_DIR)/PRODUCT.yaml
	build/lib/generate_readme.sh $(TARGET_RELEASE) $(K8S_VERSION) $(TARGET_DIR)

clean:
	rm -rf k8s-conformance
	rm -f ./sonobuoy
	rm -f $(RESULTS_DIR)
