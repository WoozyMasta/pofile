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

## [0.2.0][] - 2026-03-29

### Added

* linting support in `pofile` via `lintkit`

### Changed

* lint API returns `[]lint.Diagnostic`:
  `LintDocument`, `LintDocumentWithOptions`, and
  `ParseCatalogWithDiagnostics`
* nil input in lint flow now returns `ErrNilDocument`

[0.2.0]: https://github.com/WoozyMasta/pofile/compare/v0.1.1...v0.2.0

## [0.1.1][] - 2026-03-05

### Added

* First public release

[0.1.1]: https://github.com/WoozyMasta/pofile/tree/v0.1.1

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
