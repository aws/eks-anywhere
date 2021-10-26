#!/usr/bin/env bash

setupFakeKubebuilderEnv() {
	ln -sr pkg/api api
	mv controllers controllers_tmp
	ln -sr controllers_tmp/controllers controllers
	ln -sr controllers_tmp/main.go main.go
}

restoreKubebuilderEnv() {
	rm main.go
	rm controllers
	mv controllers_tmp controllers
	rm api
}

setupFakeKubebuilderEnv
./hack/tools/bin/kubebuilder "$@"
restoreKubebuilderEnv
