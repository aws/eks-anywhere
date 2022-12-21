ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

SOURCES := $(shell find ${ROOT_DIR} -name \*.puml)
DIAGRAMS := $(SOURCES:%.puml=%.png)

REGISTRY ?= public.ecr.aws/i0f3w2d9
PLANTUML_VERSION ?= 1.2022.6
IMAGE_NAME ?= eksa-plantuml
PLANTUML_IMG ?= $(REGISTRY)/$(IMAGE_NAME)
PLANTUML_IMG_TAGGED ?= $(PLANTUML_IMG):$(PLANTUML_VERSION)

.PHONY: diagrams
diagrams: $(DIAGRAMS)

%.png: %.puml
	echo "running"
	docker run \
		--rm \
		--volume ${ROOT_DIR}:/workdir$(DOCKER_VOL_OPTS) \
		--user $(shell id -u):$(shell id -g) \
		$(PLANTUML_IMG_TAGGED) \
		-v /workdir/$(shell echo '$^' | sed -e 's,.*designs/,,g')

.PHONY: plantuml-image
plantuml-image:
	docker build --build-arg PLANTUML_VERSION=$(PLANTUML_VERSION) -t $(PLANTUML_IMG_TAGGED) .

.PHONY: push-plantuml-image
push-plantuml-image: plantuml-image
	aws ecr-public create-repository --repository-name $(IMAGE_NAME) --region us-east-1 || true
	docker push $(PLANTUML_IMG_TAGGED)
