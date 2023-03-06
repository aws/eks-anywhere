SHELL := /bin/bash
GIT_TAG?=$(shell git describe --tag | cut -d'-' -f1)
GITHUB_USER?=$(shell git config --get remote.origin.url|sed -e 's/.*://' -e 's,/.*,,')
SONOBUOY_OSARCH?=darwin_amd64
TARGET_RELEASE?=1.24
TARGET_DIR=k8s-conformance/v$(TARGET_RELEASE)/eks-a
K8S_VERSION=v1.24.6
YOUR_EMAIL=$(shell git config --get user.email)

.PHONY: default
default: clean k8s-conformance sonobuoy sonobuoy-run sonobuoy-results generate git-commit

k8s-conformance:
	git clone git@github.com:$(GITHUB_USER)/k8s-conformance.git
	cd k8s-conformance ; git remote add upstream https://github.com/cncf/k8s-conformance.git
	cd k8s-conformance ; git fetch upstream
	cd k8s-conformance ; git checkout master
	cd k8s-conformance ; git reset --hard upstream/master
	cd k8s-conformance ; git branch -D v$(TARGET_RELEASE) || true
	cd k8s-conformance ; git checkout -b v$(TARGET_RELEASE)
	mkdir -p $(TARGET_DIR)
	rm -f $(TARGET_DIR)/*

sonobuoy: SONOBUOY=https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.56.11/sonobuoy_0.56.11_$(SONOBUOY_OSARCH).tar.gz
sonobuoy:
	wget -qO- ${SONOBUOY} |tar -xz sonobuoy
	chmod 755 sonobuoy

.PHONY: sonobuoy-run
sonobuoy-run:
	./sonobuoy run --mode=certified-conformance --wait --kube-conformance-image registry.k8s.io/conformance:$(K8S_VERSION)

.PHONY: sonobuoy-results
sonobuoy-results:
	build/lib/get_results.sh $(TARGET_DIR)

.PHONY: generate
generate:
	sed -e "s/{{k8s_version}}/$(K8S_VERSION)/" -e "s/{{your_email}}/$(YOUR_EMAIL)/" <template/PRODUCT.yaml >$(TARGET_DIR)/PRODUCT.yaml
	build/lib/generate_readme.sh $(TARGET_RELEASE) $(K8S_VERSION) $(TARGET_DIR)

.PHONY: git-commit
git-commit:
	cd k8s-conformance ; git add v$(TARGET_RELEASE)/eks-a/*
	cd k8s-conformance ; git commit -a -m "AWS EKS Anywhere v$(TARGET_RELEASE) conformance" --signoff
	cd k8s-conformance ; git push

clean:
	rm -rf k8s-conformance
	rm -f ./sonobuoy
