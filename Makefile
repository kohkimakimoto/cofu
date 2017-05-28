.PHONY:help dev dist packaging fmt test deps
.DEFAULT_GOAL := help

# This is a magic code to output help message at default
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Build dev binary
	@bash -c $(CURDIR)/build/scripts/dev.sh

dist: ## Build dist binaries
	@bash -c $(CURDIR)/build/scripts/dist.sh

packaging: ## Create packages (now support RPM only)
	@bash -c $(CURDIR)/build/scripts/packaging.sh

clean: ## Clean the built binaries.
	@bash -c $(CURDIR)/build/scripts/clean.sh

fmt:
	go fmt $$(go list ./... | grep -v vendor)

test: ## Run all tests
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos7" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos6" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:debian8" && bash -c $(CURDIR)/test/test.sh

deps: ## Install dependences by using glide
	glide install

