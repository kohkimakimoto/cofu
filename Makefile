.PHONY:help dev dist packaging fmt test testv deps
.DEFAULT_GOAL := help

# This is a magic code to output help message at default
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Build dev binary
	@bash -c $(CURDIR)/build/dev.sh

dist: ## Build dist binaries
	@bash -c $(CURDIR)/build/dist.sh

packaging: ## Create packages (now support RPM only)
	@bash -c $(CURDIR)/build/packaging.sh

clean: ## Clean the built binaries.
	@bash -c $(CURDIR)/build/clean.sh

fmt:
	go fmt $$(go list ./... | grep -v vendor)

test:
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos7" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos6" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:debian8" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:debian7" && bash -c $(CURDIR)/test/test.sh

testv:
	@export GOTEST_FLAGS="-cover -timeout=360s -v" && export DOCKER_IMAGE="kohkimakimoto/golang:centos7" && bash -c $(CURDIR)/test/test.sh
	@export GOTEST_FLAGS="-cover -timeout=360s -v" && export DOCKER_IMAGE="kohkimakimoto/golang:centos6" && bash -c $(CURDIR)/test/test.sh
	@export GOTEST_FLAGS="-cover -timeout=360s -v" && export DOCKER_IMAGE="kohkimakimoto/golang:debian8" && bash -c $(CURDIR)/test/test.sh
	@export GOTEST_FLAGS="-cover -timeout=360s -v" && export DOCKER_IMAGE="kohkimakimoto/golang:debian7" && bash -c $(CURDIR)/test/test.sh

testone:
	@export GOTEST_FLAGS="-cover -timeout=360s -v" && export DOCKER_IMAGE="kohkimakimoto/golang:centos7" && bash -c $(CURDIR)/test/test.sh

deps: ## Install dependences by using glide
	glide install

