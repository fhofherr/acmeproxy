# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep
a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

* A `github.com/fhofherr/acmeproxy/pkg/acmetest` package containing
  a type `Pebble`. It represents an instance of the
  [pebble](https://github.com/letsencrypt/pebble) test server.
* A local test environment using `docker-compose`. It runs `pebble` and
  allows to execute any integration tests locally.

[Unreleased]: https://github.com/fhofherr/leproxy
