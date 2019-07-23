# This is a self-documenting Makefile.
# See https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.DEFAULT_GOAL := all

GO := go
GOLINT := golint
GOLANGCI_LINT := golangci-lint
DOCKER_COMPOSE := docker-compose

HOST_IP := $(shell $(GO) run scripts/dev/hostip/main.go)

.PHONY: all
all: documentation test

.PHONY: test
test: ## Execute all tests and show a coverage summary
	$(GO) test -coverprofile=coverage.out ./...

.PHONY: coverageHTML
coverageHTML: test ## Create HTML coverage report
	$(GO) tool cover -html=coverage.out

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
dev-env-up: | .pebble ## Start the local development environment
ifeq ($(strip $(HOST_IP)),)
	$(error HOST_IP has to be set)
endif
	HOST_IP=$(HOST_IP) $(DOCKER_COMPOSE) -f docker/docker-compose.dev.yml up --build --detach
	@echo
	@echo "***** Local development environment started *****"
	@echo

.PHONY: dev-env-down
dev-env-down: ## Shut the local development environment down
ifeq ($(strip $(HOST_IP)),)
	$(error HOST_IP has to be set)
endif
	HOST_IP=$(HOST_IP) $(DOCKER_COMPOSE) -f docker/docker-compose.dev.yml down
	@echo
	@echo "***** Local development environment stopped *****"
	@echo

.pebble:
	git clone https://github.com/letsencrypt/pebble $@

.PHONY: clean
clean:
	rm -rf .pebble

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
