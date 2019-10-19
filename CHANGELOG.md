# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep
a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2019-10-18

### Added

* The `acmeproxy serve` command which starts `acmeproxy` in server mode.
  The initial version of `acmeproxy` attempts to obtain a certificate
  for `www.example.com` from the configured Let's encrypt CA. This
  version of `acmeproxy` should thus **not** be started using the
  official Let's Encrypt production environment.
* `acmeproxy`'s tests start a local instance of
  [pebble](https://github.com/letsencrypt/pebble). This instance is used
  for integration testing against a "real" ACME CA.
* `Dockerfile` which allows to create a docker image for `acmeproxy`.
* The `scripts/test-env.sh` script which starts a local test environment
  for `acmeproxy`. Within this test environment `acmeproxy` is executing
  against pebble. Until `acmeproxy` stabilizes this should be the only
  way to start `acmeproxy`.

[Unreleased]: https://github.com/fhofherr/acmeproxy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/fhofherr/acmeproxy/releases/tag/v0.1.0
