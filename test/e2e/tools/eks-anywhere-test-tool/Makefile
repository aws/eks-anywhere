GO_ARCH:=$(shell go env GOARCH)
GO_OS:=$(shell go env GOOS)
BINARY_NAME:=bin/eks-a-test-tool

.PHONY: build-test-tool
build-test-tool:
	GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build -o $(BINARY_NAME) github.com/aws/eks-anywhere-test-tool

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/*

.PHONY: lint
lint: bin/golangci-lint ## Run golangci-lint
	bin/golangci-lint run

bin/golangci-lint: ## Download golangci-lint
bin/golangci-lint: GOLANGCI_LINT_VERSION?=$(shell cat .github/workflows/golangci-lint.yml | sed -n -e 's/^\s*version: //p')
bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)