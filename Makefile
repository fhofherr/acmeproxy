# This is a self-documenting Makefile.
# See https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.DEFAULT_GOAL := all

GO := go
GOLINT := golint
GOLANGCI_LINT := golangci-lint
DOCKER_COMPOSE := docker-compose

HOST_IP := $(shell $(GO) run scripts/dev/hostip/main.go)

GO_PACKAGES := $(shell $(GO) list ./... | grep -v scripts | tr "\n" ",")

PEBBLE_DIR := .pebble
COVERAGE_FILE := .coverage.out

.PHONY: all
all: documentation lint test

.PHONY: test
test: .$(COVERAGE_FILE) ## Execute all tests and show a coverage summary

$(COVERAGE_FILE):
# Using -coverprofile together with -coverpkg works since Go 1.10. Thus
# there is no need for some complicated concatenating of coverprofiles.
# We still use an explicit list of packages for coverpkg, since using all
# would instrument dependencies as well.
#
# See https://golang.org/doc/go1.10#test
	$(GO) test -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) -coverpkg=$(GO_PACKAGES) ./... 2> /dev/null

.PHONY: coverageHTML
coverageHTML: $(COVERAGE_FILE) ## Create HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: race
race: ## Execute all tests with race detector enabled
	$(GO) test -race ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run
	$(GOLINT) ./...

.PHONY: documentation
documentation: ## Update the documentation
	make -C doc/img svg

.PHONY: dev-env-up
dev-env-up: | $(PEBBLE_DIR) ## Start the local development environment
ifeq ($(strip $(HOST_IP)),)
	$(error HOST_IP has to be set)
endif
	HOST_IP=$(HOST_IP) $(DOCKER_COMPOSE) -f docker/docker-compose.dev.yml up --build --detach
	@echo
	@echo "***** Local development environment started *****"
	@echo
	@echo "Execute:"
	@echo
	@echo "\texport ACMEPROXY_PEBBLE_HOST=localhost"
	@echo "\texport ACMEPROXY_PEBBLE_TEST_CERT=$(PWD)/$(PEBBLE_DIR)/test/certs/pebble.minica.pem"
	@echo "\texport ACMEPROXY_PEBBLE_ACME_PORT=14000"
	@echo "\texport ACMEPROXY_PEBBLE_MGMT_PORT=15000"

.PHONY: dev-env-down
dev-env-down: ## Shut the local development environment down
ifeq ($(strip $(HOST_IP)),)
	$(error HOST_IP has to be set)
endif
	HOST_IP=$(HOST_IP) $(DOCKER_COMPOSE) -f docker/docker-compose.dev.yml down
	@echo
	@echo "***** Local development environment stopped *****"
	@echo

$(PEBBLE_DIR):
	git clone https://github.com/letsencrypt/pebble $@

.PHONY: clean
clean:
	rm -rf $(PEBBLE_DIR)
	rm -rf $(COVERAGE_FILE)

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
