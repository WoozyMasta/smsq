<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.2.0][] - 2026-10-31

### Added

* `udp` command for the Steam Master Server Query Protocol (MSQP),
  the legacy server-list protocol.
* `grpc` command for server queries through an authenticated
  Steam Connection Manager (CM) session over protobuf/WebSocket.
  Supports QR and Steam Guard login with optional token reuse.
* Discovery with reusable plans for large Web API and CM result sets.
* Selectable output formats: table, JSON, YAML, CSV, Markdown, and HTML.
  Output columns can be selected.

### Changed

* Filter, query, and output options are shared across protocol commands;
  `rest` remains the default.
* Default output changed from JSON to a table;
  the default result limit changed from 100 to 25.
* Short options changed: `-t` now sets the timeout, and `-c` selects columns.

### Removed

* `appid=...` is no longer accepted as a filter; use `--app-id`.
* `--table` and `--column-width` are no longer available;
  use `--format` and `--columns`.

[0.2.0]: https://github.com/WoozyMasta/smsq/compare/v0.1.0...v0.2.0

## [0.1.0][] - 2025-07-30

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/smsq/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
