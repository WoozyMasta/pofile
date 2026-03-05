# pofile

`pofile` is a Go module for gettext text catalogs: `.po` and `.pot`.

It is built for predictable file processing in build tools and localization
pipelines, without pulling in runtime gettext behavior.

It provides:

* semantic parse API (`Parse`, `ParseReader`, `ParseFile`, `ParseDir`)
* lossless parse API (`ParseDocument*`) with source positions
* parser diagnostics (`ParseCatalogWithDiagnostics`)
* formatter/writer (`Format`, `WriteFile`, `FormatDocument`)
* semantic model (`Catalog`, `Message`)
* lossless model (`Document`, `Entry`, `Comment`, `Header`)
* index for fast key lookup in lossless model (`NewIndex`, `EntryKey`)
* lint/validation (`LintDocumentWithOptions`, `ValidateDocument`)
* merge helper (`MergeTemplate`)

Scope:

* text `.po/.pot` only
* no binary `.mo`
* no runtime gettext engine

## Install

```bash
go get github.com/woozymasta/pofile
```

## Quick Example

```go
package main

import (
    "log"

    "github.com/woozymasta/pofile"
)

func main() {
    catalog, err := pofile.ParseFile("l18n/russian.po")
    if err != nil {
        log.Fatal(err)
    }

    catalog.UpsertMessage("UI_OK", "OK", "Ок")
    catalog.SetHeader("Last-Translator", "team@example.com")

    if err := pofile.WriteFile("l18n/russian.po", catalog); err != nil {
        log.Fatal(err)
    }
}
```

## Diagnostics Example

```go
catalog, diagnostics, err := pofile.ParseCatalogWithDiagnostics(
    data,
    pofile.ParseOptions{AllowInvalid: true},
)
if err != nil {
    return err
}
_ = catalog
for _, d := range diagnostics {
    // d.Code, d.Severity, d.Position, d.Span
}
```

## Lint Rules

`LintDocument` and `LintDocumentWithOptions` emit diagnostics with codes:

* `PO2000`: input document is nil
* `PO2001`: duplicate `domain/context/msgid` entry
* `PO2002`: `msgstr[n]` exists, but `msgid_plural` is missing
* `PO2003`: translations exist, but `Language` header is empty
* `PO2004`: entry has empty `msgid`
* `PO2005`: printf-like placeholders mismatch between source and translation
* `PO2006`: `msgid_plural` exists, but `msgstr[n>0]` is missing

Severity policy:

* `basic` mode: non-critical checks are warnings
* `strict` mode: warnings are upgraded to errors
