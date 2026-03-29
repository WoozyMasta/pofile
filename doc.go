// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

/*
Package pofile provides primitives for gettext text catalogs (.po/.pot).

The package exposes two data layers:
  - Catalog for semantic operations (upsert/find/delete, headers, merge)
  - Document for lossless workflows (comments, ordering, source positions)

Use ParseFile or ParseDir for semantic flows. Use ParseDocumentFile and
ParseDocumentDir when you need source-preserving round-trip behavior.
Use ParseCatalogWithDiagnostics for tolerant parse with structured findings.
Lint and parse diagnostics use stable machine-readable codes.
The module includes lintkit integration for diagnostics metadata and
rule-provider registration.

Scope is intentionally narrow:
  - text .po/.pot only
  - no binary .mo support
  - no runtime gettext lookup engine
*/
package pofile
