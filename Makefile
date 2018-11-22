.DEFAULT_GOAL := help

# This is a magic code to output help message at default
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY:help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY:dev
dev: ## Build dev binary
	@bash -c $(CURDIR)/build/scripts/dev.sh

.PHONY:dist
dist: ## Build dist binaries
	@bash -c $(CURDIR)/build/scripts/dist.sh

.PHONY:packaging
packaging: ## Create packages (now support RPM only)
	@bash -c $(CURDIR)/build/scripts/packaging.sh

.PHONY:clean
clean: ## Clean the built binaries.
	@bash -c $(CURDIR)/build/scripts/clean.sh

.PHONY:fmt
fmt:
	go fmt $$(go list ./... | grep -v vendor)

.PHONY:test
test: ## Run all tests
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos7" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:centos6" && bash -c $(CURDIR)/test/test.sh
	@export DOCKER_IMAGE="kohkimakimoto/golang:debian8" && bash -c $(CURDIR)/test/test.sh

.PHONY:deps
deps: ## Install dependences.
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/mitchellh/gox
	dep ensure

.PHONY:resetdeps
resetdeps: ## reset dependences.
	rm -rf Gopkg.*
	rm -rf vendor
	dep init
	dep ensure
