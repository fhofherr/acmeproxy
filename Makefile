# This is a self-documenting Makefile.
# See https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.DEFAULT_GOAL := all

GO := go
GOLINT := golint
GOLANGCI_LINT := golangci-lint
DOCKER_COMPOSE := docker-compose

HOST_IP := $(shell $(GO) run scripts/dev/hostip/main.go)

BIN_DIR := bin
BIN_DIR_LOCAL := $(BIN_DIR)/local
BIN_DIR_LINUX_AMD64 := $(BIN_DIR)/linux/amd64
BINARY_NAME := acmeproxy

SCRIPTS_DIR := scripts

PEBBLE_DIR := .pebble
COVERAGE_FILE := .coverage.out

GO_PACKAGES := $(shell $(GO) list ./... | grep -v scripts | tr "\n" ",")
GO_FILES := $(shell find . -iname '*.go' -not -path "./$(PEBBLE_DIR)/*" -not -path "./$(SCRIPTS_DIR)/*")


.PHONY: all
all: documentation lint test build

.PHONY: test
test: $(COVERAGE_FILE) ## Execute all tests and show a coverage summary

$(COVERAGE_FILE): $(GO_FILES)
# Using -coverprofile together with -coverpkg works since Go 1.10. Thus
# there is no need for some complicated concatenating of coverprofiles.
# We still use an explicit list of packages for coverpkg, since using all
# would instrument dependencies as well.
#
# See https://golang.org/doc/go1.10#test
	@$(GO) test -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) -coverpkg=$(GO_PACKAGES) ./... 2> /dev/null

.PHONY: coverage
coverage: $(COVERAGE_FILE) ## Compute and display the current total code coverage
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n1

.PHONY: coverageHTML
coverageHTML: $(COVERAGE_FILE) ## Create HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: race
race: ## Execute all tests with race detector enabled
	$(GO) test -race ./...

.PHONY: test-update
test-update: ## Execute all tests that have a -update flag defined.
	$(GO) test ./pkg/certutil -update

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

$(BIN_DIR_LOCAL):
	mkdir -p $@

$(BIN_DIR_LINUX_AMD64):
	mkdir -p $@

$(BIN_DIR_LOCAL)/$(BINARY_NAME): $(GO_FILES) | $(BIN_DIR_LOCAL)
	$(GO) build -o $@

$(BIN_DIR_LINUX_AMD64)/$(BINARY_NAME): $(GO_FILES) | $(BIN_DIR_LINUX_AMD64)
	GOOS=linux GOARCH=amd64 $(GO) build -o $@

.PHONY: build-local
build-local: $(BIN_DIR_LOCAL)/$(BINARY_NAME)

# TODO see if we can use make file rules once we need to crosscompile for other
#      architectures.
.PHONY: build-linux-amd64
build-linux-amd64: $(BIN_DIR_LINUX_AMD64)/$(BINARY_NAME)

.PHONY: build
build: build-local build-linux-amd64

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(PEBBLE_DIR)
	rm -rf $(COVERAGE_FILE)

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
