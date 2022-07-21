#!/usr/bin/env bash

TEMP_FOLDER=tmp-kubebuilder-project
PATH_TO_KUBEBUILDER=hack/tools/bin/kubebuilder

setupFakeKubebuilderEnv() {
	mkdir -p $TEMP_FOLDER/hack
	ln -s $(pwd)/pkg/api $TEMP_FOLDER/api
	ln -s $(pwd)/controllers $TEMP_FOLDER/controllers
	ln -s $(pwd)/manager/main.go $TEMP_FOLDER/main.go
	ln -s $(pwd)/hack/boilerplate.go.txt $TEMP_FOLDER/hack/boilerplate.go.txt
	ln -s $(pwd)/PROJECT $TEMP_FOLDER/PROJECT
	cp hack/fake-Makefile $TEMP_FOLDER/Makefile
}

restoreKubebuilderEnv() {
	rm -rf $TEMP_FOLDER
}

if [ ! -f $PATH_TO_KUBEBUILDER ]; then
	echo "Kubebuilder is not installed. Run 'make hack/tools/bin/kubebuilder'"
	exit 1
fi

setupFakeKubebuilderEnv
(cd $TEMP_FOLDER && ./../$PATH_TO_KUBEBUILDER "$@")
restoreKubebuilderEnv
