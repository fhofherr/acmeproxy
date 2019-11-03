# This is a self-documenting Makefile.
# See https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# Get the directory which contains this Makefile.
# See: http://timmurphy.org/2015/09/27/how-to-get-a-makefile-directory-path/
PRJ_DIR := $(patsubst %/,%,$(dir $(realpath $(firstword $(MAKEFILE_LIST)))))


GIT := $(shell command -v git 2> /dev/null)
GO := $(shell command -v go 2> /dev/null)
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
GOLINT := $(shell command -v golint 2> /dev/null)
PODMAN := $(shell command -v podman 2> /dev/null)
PROTOC := $(shell command -v protoc 2> /dev/null)
SED := $(shell command -v sed 2> /dev/null)

BIN_DIR := bin
BINARY_NAME := acmeproxy
TARGET_ARCHITECTURES := local linux/amd64
BINARY_FILES:=$(addsuffix /$(BINARY_NAME),$(addprefix $(BIN_DIR)/,$(TARGET_ARCHITECTURES)))

SCRIPTS_DIR := scripts

GO_PACKAGES := $(shell $(GO) list ./... | grep -v scripts | tr "\n" ",")
GO_FILES := $(shell find . -iname '*.go' -not -path "./$(PEBBLE_DIR)/*" -not -path "./$(SCRIPTS_DIR)/*" -not -path "*.pb.go")

.DEFAULT_GOAL := all
.PHONY: all
all: documentation lint test build

# -----------------------------------------------------------------------------
#
# Tests
#
# -----------------------------------------------------------------------------
PEBBLE_DIR := .pebble
ACMEPROXY_PEBBLE_DIR := $(PRJ_DIR)/$(PEBBLE_DIR)
COVERAGE_FILE := .coverage.out

.PHONY: test
test: coverage ## Execute all tests and show a coverage summary

.PHONY: coverage
coverage: $(COVERAGE_FILE) ## Compute and display the current total code coverage
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n1

.PHONY: coverageHTML
coverageHTML: $(COVERAGE_FILE) ## Create HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE)

$(COVERAGE_FILE): $(GO_FILES) $(PEBBLE_DIR)/pebble $(PEBBLE_DIR)/pebble-challtestsrv
# Using -coverprofile together with -coverpkg works since Go 1.10. Thus
# there is no need for some complicated concatenating of coverprofiles.
# We still use an explicit list of packages for coverpkg, since using all
# would instrument dependencies as well.
#
# See https://golang.org/doc/go1.10#test
	@ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test -race -covermode=atomic -coverprofile=$@ -coverpkg=$(GO_PACKAGES) ./... 2> /dev/null
	@$(SED) -i.bak '/\.pb\.go/d' $@
	@$(SED) -i.bak '/testing\.go/d' $@
	@$(SED) -i.bak '/\/testsupport\//d' $@

$(PEBBLE_DIR):
	$(GIT) clone https://github.com/letsencrypt/pebble $@

$(PEBBLE_DIR)/pebble: | $(PEBBLE_DIR)
	cd $(PEBBLE_DIR); $(GO) build -o pebble ./cmd/pebble

$(PEBBLE_DIR)/pebble-challtestsrv: | $(PEBBLE_DIR)
	cd $(PEBBLE_DIR); $(GO) build -o pebble-challtestsrv ./cmd/pebble-challtestsrv

.PHONY: pebble
pebble: $(PEBBLE_DIR)/pebble $(PEBBLE_DIR)/pebble-challtestsrv

.PHONY: test-update
test-update: $(PEBBLE_DIR)/pebble $(PEBBLE_DIR)/pebble-challtestsrv ## Execute all tests that have an -update flag defined.
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/acme -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/acme/acmeclient -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/api/grpcapi -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/certutil -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/db -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/db/internal/dbrecords -update

# -----------------------------------------------------------------------------
#
# Build
#
# -----------------------------------------------------------------------------

.PHONY: build-local
build-local: $(BIN_DIR)/local/$(BINARY_NAME) ## Build a binary for the local machine only

.PHONY: build
build: $(BINARY_FILES) ## Build all binary files

$(BIN_DIR)/%/$(BINARY_NAME): $(GO_FILES)
	@$(GO) run ./$(SCRIPTS_DIR)/xbuild $(XBUILD_FLAGS) $@

# -----------------------------------------------------------------------------
#
# Build Docker image
#
# -----------------------------------------------------------------------------
.PHONY: build-image
build-image: ## Build a Docker image and tag it as acmeproxy:latest
	$(PODMAN) build --tag acmeproxy:latest .


# -----------------------------------------------------------------------------
#
# Code Linters
#
# -----------------------------------------------------------------------------
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run
	$(GOLINT) ./...

# -----------------------------------------------------------------------------
#
# Documentation
#
# -----------------------------------------------------------------------------
.PHONY: documentation
documentation: ## Update the documentation
	make -C doc/img svg

# -----------------------------------------------------------------------------
#
# Cleanups
#
# -----------------------------------------------------------------------------
.PHONY: clean
clean: ## Remove all intermediate directories and files
	rm -rf $(BIN_DIR)
	rm -rf $(PEBBLE_DIR)
	rm -rf $(COVERAGE_FILE)
	rm -rf $(COVERAGE_FILE).bak

# -----------------------------------------------------------------------------
#
# Protocol Buffers
#
# The targets below are not added as dependencies to any of the other targets.
# All *.pb.go files are checked into version controll to ensure the module can
# be built even if protoc is not installed. Therefore those targets should only
# be called explicitly if needed.
#
# -----------------------------------------------------------------------------
PROTOBUF_SRC_FILES := $(shell find . -iname '*.proto')
PROTOBUF_GO_FILES := $(patsubst %.proto,%.pb.go,$(PROTOBUF_SRC_FILES))

pkg/api/grpcapi/internal/pb/%.pb.go: pkg/api/grpcapi/internal/pb/%.proto
	$(PROTOC) -I=$(PRJ_DIR) --go_out=plugins=grpc:$(PRJ_DIR) $<

pkg/db/internal/dbrecords/%.pb.go: pkg/db/internal/dbrecords/%.proto
	$(PROTOC) -I=$(PRJ_DIR) --go_out=$(PRJ_DIR) $<

.PHONY: pb
pb: $(PROTOBUF_GO_FILES) ## Generate all *.pb.go files

.PHONY: pb-clean
pb-clean: ## Remove all *.pb.go files
	find . -iname '*.pb.go' -delete

# -----------------------------------------------------------------------------
#
# Help
#
# -----------------------------------------------------------------------------
.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


