.DEFAULT_GOAL := help

export GO111MODULE := off
export PATH := $(CURDIR)/.go-packages/bin:$(PATH)

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
	go test -cover $$(go list ./... | grep -v vendor)

.PHONY:testv
testv: ## Run all tests with verbose outputing.
	go test -v -cover $$(go list ./... | grep -v vendor)

.PHONY: installtools
installtools: ## Install dev tools
	GOPATH=$(CURDIR)/.go-packages && \
      go get -u github.com/golang/dep/cmd/dep && \
      go get -u github.com/mitchellh/gox && \
      go get -u github.com/axw/gocov/gocov && \
      go get -u gopkg.in/matm/v1/gocov-html

.PHONY:deps
deps: ## Install dependences.
	PATH=$(CURDIR)/.go-packages/bin:${PATH} && dep ensure

.PHONY:resetdeps
resetdeps: ## reset dependences.
	rm -rf Gopkg.*
	rm -rf vendor
	dep init
	dep ensure
