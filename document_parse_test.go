// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"testing"
)

func TestParseDocumentWithDiagnostics(t *testing.T) {
	t.Parallel()

	document, diagnostics, err := ParseDocumentWithOptions(
		readFixture(t, "parse/domain_plural.po"),
		ParseOptions{},
	)
	if err != nil {
		t.Fatalf("ParseDocumentWithOptions error: %v", err)
	}
	if hasErrorDiagnostics(diagnostics) {
		t.Fatalf("unexpected error diagnostics: %+v", diagnostics)
	}

	if got := document.HeaderValue("Language"); got != "ru" {
		t.Fatalf("Language header = %q, want %q", got, "ru")
	}
	if len(document.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(document.Entries))
	}

	entry := document.Entries[0]
	if entry.Context != "key.ctx" {
		t.Fatalf("context = %q, want %q", entry.Context, "key.ctx")
	}
	if entry.IDPlural != "Many files" {
		t.Fatalf("id_plural = %q, want %q", entry.IDPlural, "Many files")
	}
	if entry.Translations[1] != "Много файлов" {
		t.Fatalf("msgstr[1] = %q, want %q", entry.Translations[1], "Много файлов")
	}
	if entry.PreviousID != "Old value" {
		t.Fatalf("previous_id = %q, want %q", entry.PreviousID, "Old value")
	}
	if len(entry.Flags) != 2 {
		t.Fatalf("flags len = %d, want 2", len(entry.Flags))
	}
	if len(entry.References) != 1 {
		t.Fatalf("references len = %d, want 1", len(entry.References))
	}
	if entry.Domain != "ui" {
		t.Fatalf("domain = %q, want %q", entry.Domain, "ui")
	}
}

func TestParseDocumentSyntaxErrorPosition(t *testing.T) {
	t.Parallel()

	_, diagnostics, err := ParseDocumentWithOptions(
		readFixture(t, "parse/syntax_missing_quote.po"),
		ParseOptions{AllowInvalid: true},
	)
	if err != nil {
		t.Fatalf("ParseDocumentWithOptions AllowInvalid error: %v", err)
	}
	if !hasErrorDiagnostics(diagnostics) {
		t.Fatalf("expected error diagnostics, got: %+v", diagnostics)
	}

	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.Code != diagCodeParseMissingQuote {
			continue
		}
		found = true
		if diagnostic.Position.Line != 5 {
			t.Fatalf("line = %d, want 5", diagnostic.Position.Line)
		}
		if diagnostic.Span.StartOffset < 0 {
			t.Fatalf("span start offset = %d, want >= 0", diagnostic.Span.StartOffset)
		}
		if diagnostic.Span.EndOffset <= diagnostic.Span.StartOffset {
			t.Fatalf(
				"invalid span offsets: start=%d end=%d",
				diagnostic.Span.StartOffset,
				diagnostic.Span.EndOffset,
			)
		}
	}
	if !found {
		t.Fatalf("missing %s diagnostic in %+v", diagCodeParseMissingQuote, diagnostics)
	}
}

func TestParseDocumentObsoleteEntry(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "parse/obsolete.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if len(document.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(document.Entries))
	}
	if !document.Entries[0].Obsolete {
		t.Fatal("entry.Obsolete = false, want true")
	}
}

func TestParseWithDiagnosticsCatalog(t *testing.T) {
	t.Parallel()

	catalog, diagnostics, err := ParseCatalogWithDiagnostics(
		readFixture(t, "parse/domain_plural.po"),
		ParseOptions{},
	)
	if err != nil {
		t.Fatalf("ParseCatalogWithDiagnostics error: %v", err)
	}
	if hasErrorDiagnostics(diagnostics) {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if got := catalog.TranslationNInDomain("ui", "key.ctx", "One file", 1); got != "Много файлов" {
		t.Fatalf("translation = %q, want %q", got, "Много файлов")
	}
}
