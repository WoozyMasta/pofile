// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"testing"
)

func TestParseReaderDomainPluralToCatalog(t *testing.T) {
	t.Parallel()

	catalog, err := Parse(readFixture(t, "parse/domain_plural.po"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	message := catalog.FindMessageInDomain("ui", "key.ctx", "One file")
	if message == nil {
		t.Fatal("FindMessageInDomain returned nil")
	}
	if !message.IsPlural() {
		t.Fatal("IsPlural = false, want true")
	}
	if message.IDPlural != "Many files" {
		t.Fatalf("IDPlural = %q, want %q", message.IDPlural, "Many files")
	}
	if got := catalog.TranslationNInDomain("ui", "key.ctx", "One file", 1); got != "Много файлов" {
		t.Fatalf("msgstr[1] = %q, want %q", got, "Много файлов")
	}
	if !message.HasFlag("fuzzy") {
		t.Fatal("HasFlag(fuzzy) = false, want true")
	}
}

func TestCatalogRoundTripPluralMetadata(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	catalog.SetHeader("Language", "ru")
	catalog.Language = "ru"

	message := catalog.UpsertMessageInDomain("ui", "key", "One file", "Один файл")
	message.IDPlural = "Many files"
	message.SetTranslationAt(1, "Много файлов")
	message.Flags = []string{"fuzzy", "c-format"}
	message.References = []string{"app/file.cpp:10"}
	message.PreviousID = "Old one"
	message.PreviousIDPlural = "Old many"

	data, err := FormatWithOptions(catalog, &WriteOptions{
		Mode: WriteModeCanonical,
	})
	if err != nil {
		t.Fatalf("FormatWithOptions error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := parsed.FindMessageInDomain("ui", "key", "One file")
	if got == nil {
		t.Fatal("parsed message is nil")
	}
	if got.IDPlural != "Many files" {
		t.Fatalf("IDPlural = %q, want %q", got.IDPlural, "Many files")
	}
	if got.TranslationAt(1) != "Много файлов" {
		t.Fatalf("msgstr[1] = %q, want %q", got.TranslationAt(1), "Много файлов")
	}
	if !got.HasFlag("fuzzy") {
		t.Fatal("HasFlag(fuzzy) = false, want true")
	}
	if len(got.References) == 0 || got.References[0] != "app/file.cpp:10" {
		t.Fatalf("references = %v, want app/file.cpp:10", got.References)
	}
	if got.PreviousID != "Old one" || got.PreviousIDPlural != "Old many" {
		t.Fatalf(
			"previous fields mismatch: id=%q id_plural=%q",
			got.PreviousID,
			got.PreviousIDPlural,
		)
	}
}
