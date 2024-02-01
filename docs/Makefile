DOCS_FOLDER=$(shell git rev-parse --show-toplevel)/docs
IMAGE_REGISTRY ?= public.ecr.aws/eks-distro-build-tooling
CONTAINER_IMAGE   = $(IMAGE_REGISTRY)/eks-a-hugo
DOCS_ARCHIVE      = public.zip
HTMLTEST_VERSION?=v0.16.0

# Detach can be used with the container-serve command to run the container in the background.
DETACH ?= false

# _FOCUS_OPTS describes the options for focusing the serve container (background vs foreground).
_FOCUS_OPTS := --interactive --tty
ifeq ($(DETACH),true)
	_FOCUS_OPTS := --detach
endif

.PHONY: help build release submodule zip deploy serve server container-build container-serve container-server clean

help: ## Show this help. Requires locally installed npm and hugo.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: build-docs htmltest

build-docs: submodule ## Generate the static site into the /public folder
	npm install
	hugo --environment production --cleanDestinationDir

release: build

submodule: ## Initialize the docsy theme submodule
	git submodule update --init --recursive

submodule-apply-patches: submodule ## Apply submodule patches
	cd themes/docsy; git am ../../patches/0001-Make-docsy-js-files-static.patch; cd ../..

submodule-reset: ## Reset submodules to tracked commit
	git submodule update --init --recursive --force

htmltest: install-htmltest
	bin/htmltest -c $(DOCS_FOLDER)/.htmltest.yml -s $(DOCS_FOLDER)/public

install-htmltest:
	curl https://htmltest.wjdp.uk | bash -s $(HTMLTEST_VERSION)

zip: clean build ## Create zip file of assets to upload
	# move into public folder so zip file doesn't have a public folder
	# put the zip file up one directory in docs to be used by deploy target
	cd public && zip -r --quiet ../$(DOCS_ARCHIVE) ./

deploy: zip ## Deploy docs to Amplify
	bash ./deploy-docs.sh public.zip

serve: submodule build ## Boot the development server.
	hugo server --buildFuture --baseURL http://127.0.0.1

server: serve

container-build: submodule-apply-patches ## Build a container image for the preview of the website
	docker build -t $(CONTAINER_IMAGE) .
	$(MAKE) submodule-reset

container-serve: ## Boot the development server using container. Run `make container-build` before this.
	docker run \
		--rm $(_FOCUS_OPTS) \
		--volume $(CURDIR)/../:/src \
		--mount type=tmpfs,destination=/tmp,tmpfs-mode=01777 \
		--workdir /src/docs \
		-p 1313:1313 $(CONTAINER_IMAGE) server \
		--buildFuture \
		--buildDrafts \
		--bind 0.0.0.0 \
		--destination /tmp/hugo \
		--cleanDestinationDir \
		--baseURL http://0.0.0.0

container-server: container-serve

clean: ## Delete generated content. (requires sudo if using container-serve)
	rm -rf public resources node_modules public.zip

compute-checksum: ## Compute docs checksum
	sha256sum public/index.html > DOCS_CHECKSUM

upload-checksum: compute-checksum ## Upload the checksum to artifacts bucket
	aws s3 cp DOCS_CHECKSUM s3://$(ARTIFACTS_BUCKET) --acl public-read
