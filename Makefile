# This is a self-documenting Makefile.
# See https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# Get the directory which contains this Makefile.
# See: http://timmurphy.org/2015/09/27/how-to-get-a-makefile-directory-path/
PRJ_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := all

GO := go
GOPATH := $(shell $(GO) env GOPATH | tr -d "\n\t ")
GOBIN := $(GOPATH)/bin
GOLINT := $(GOBIN)/golint
GOLANGCI_LINT := $(GOBIN)/golangci-lint
PROTOC := protoc
SED := sed

BIN_DIR := bin
BINARY_NAME := acmeproxy
TARGET_ARCHITECTURES := local linux/amd64
BINARY_FILES:=$(addsuffix /$(BINARY_NAME),$(addprefix $(BIN_DIR)/,$(TARGET_ARCHITECTURES)))

SCRIPTS_DIR := scripts

PEBBLE_DIR := .pebble
# PRJ_DIR ends with /. We therefore omit it here.
ACMEPROXY_PEBBLE_DIR := $(PRJ_DIR)$(PEBBLE_DIR)
COVERAGE_FILE := .coverage.out

GO_PACKAGES := $(shell $(GO) list ./... | grep -v scripts | tr "\n" ",")
GO_FILES := $(shell find . -iname '*.go' -not -path "./$(PEBBLE_DIR)/*" -not -path "./$(SCRIPTS_DIR)/*" -not -path "*.pb.go")

# -----------------------------------------------------------------------------

.PHONY: all
all: documentation lint test build

.PHONY: test
test: coverage ## Execute all tests and show a coverage summary

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

.PHONY: coverage
coverage: $(COVERAGE_FILE) ## Compute and display the current total code coverage
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n1

.PHONY: coverageHTML
coverageHTML: $(COVERAGE_FILE) ## Create HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: test-update
test-update: $(PEBBLE_DIR)/pebble $(PEBBLE_DIR)/pebble-challtestsrv ## Execute all tests that have an -update flag defined.
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/certutil -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/db/internal/dbrecords -update
	ACMEPROXY_PEBBLE_DIR=$(ACMEPROXY_PEBBLE_DIR) $(GO) test ./pkg/db/db -update

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run
	$(GOLINT) ./...

.PHONY: documentation
documentation: ## Update the documentation
	make -C doc/img svg

$(PEBBLE_DIR):
	git clone https://github.com/letsencrypt/pebble $@

$(PEBBLE_DIR)/pebble: | $(PEBBLE_DIR)
	cd $(PEBBLE_DIR); $(GO) build -o pebble ./cmd/pebble

$(PEBBLE_DIR)/pebble-challtestsrv: | $(PEBBLE_DIR)
	cd $(PEBBLE_DIR); $(GO) build -o pebble-challtestsrv ./cmd/pebble-challtestsrv

.PHONY: pebble
pebble: $(PEBBLE_DIR)/pebble $(PEBBLE_DIR)/pebble-challtestsrv

# We use an inline shell script to make this rule easier to write.
# See https://stackoverflow.com/a/29085684/86967
$(BIN_DIR)/%/$(BINARY_NAME): $(GO_FILES)
	@{ \
	set -e ;\
	GOBUILD="$(GO) build -o $@"; \
	if [ "$*" != "local" ]; then \
		goos="$$(echo $* | cut -d'/' -f1)"; \
		goarch="$$(echo $* | cut -d'/' -f2)"; \
		GOBUILD="GOOS=$$goos GOARCH=$$goarch $$GOBUILD"; \
	fi; \
	echo $$GOBUILD; \
	eval $$GOBUILD; \
	}

.PHONY: build-local ## Build a binary for the local machine only
build-local: $(BIN_DIR)/local/$(BINARY_NAME)

.PHONY: build
build: $(BINARY_FILES) ## Build all binary files

.PHONY: clean
clean: ## Remove all intermediate directories and files
	rm -rf $(BIN_DIR)
	rm -rf $(PEBBLE_DIR)
	rm -rf $(COVERAGE_FILE)
	rm -rf $(COVERAGE_FILE).bak

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


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

%.pb.go: %.proto
	$(PROTOC) -I=$(PWD) --go_out=$(PWD) $<

.PHONY: pb
pb: $(PROTOBUF_GO_FILES) ## Generate all *.pb.go files

.PHONY: pb-clean
pb-clean: ## Remove all *.pb.go files
	find . -iname '*.pb.go' -delete
