# Shell to use for running scripts
export SHELL := /bin/bash
# Get docker path or an empty string
DOCKER := $(shell command -v docker)

DOCKER_IMAGE = objectrocket/sensu-backend
# allow builds without tags
IMAGE_VERSION ?= latest

# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif


.PHONY: test
test:
	@./build.sh unit

.PHONY: integration
integration:
	@./build.sh integration

.PHONY: clean
clean:
	@go clean


docker-build: deps-development
	docker build --build-arg APPVERSION=$(IMAGE_VERSION) -t $(DOCKER_IMAGE):$(IMAGE_VERSION) .

docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(IMAGE_VERSION)
