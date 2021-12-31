SHELL := /bin/bash
GIT_TAG?=$(shell git describe --tag | cut -d'-' -f1)
GITHUB_USER?=$(shell git config --get remote.origin.url|sed -e 's/.*://' -e 's,/.*,,')
SONOBUOY_OSARCH?=darwin_amd64

.PHONY: default
default: k8s-conformance sonobuoy v1.21

k8s-conformance:
	git clone git@github.com:$(GITHUB_USER)/k8s-conformance.git
	cd k8s-conformance ; git remote add upstream https://github.com/cncf/k8s-conformance.git
	cd k8s-conformance ; git fetch upstream
	cd k8s-conformance ; git checkout master
	cd k8s-conformance ; git merge upstream/master

sonobuoy: SONOBUOY=https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.50.0/sonobuoy_0.50.0_$(SONOBUOY_OSARCH).tar.gz
sonobuoy:
	wget -qO- ${SONOBUOY} |tar -xz sonobuoy
	chmod 755 sonobuoy

.PHONY: v1.21
v1.21: TARGET_DIR=k8s-conformance/v1.21/eks-a
v1.21: K8S_VERSION=v1.21.2
v1.21: EKSA_K8S_VERSION=1.21
v1.21: CONFORMANCE_VERSION=v1.21.2
v1.21:
	mkdir -p $(TARGET_DIR)
	sed -e "s/{{k8s_version}}/$(K8S_VERSION)/" <template/PRODUCT.yaml >$(TARGET_DIR)/PRODUCT.yaml
	build/lib/generate_readme.sh $(EKSA_K8S_VERSION) $(K8S_VERSION) $(TARGET_DIR)
