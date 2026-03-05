// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import "testing"

func TestLintDocument(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{
			{Key: "X-Generator", Value: "test"},
		},
		Entries: []*Entry{
			{
				ID: "same",
				Translations: map[int]string{
					0: "A",
				},
			},
			{
				ID: "same",
				Translations: map[int]string{
					0: "B",
				},
			},
			{
				ID: "plural-no-id",
				Translations: map[int]string{
					1: "X",
				},
			},
		},
	}

	diagnostics := LintDocument(document)
	if len(diagnostics) == 0 {
		t.Fatal("LintDocument returned no diagnostics")
	}

	var (
		hasDuplicate bool
		hasPlural    bool
		hasLanguage  bool
	)
	for _, diagnostic := range diagnostics {
		if diagnostic.Span.EndOffset < diagnostic.Span.StartOffset {
			t.Fatalf(
				"invalid span offsets: start=%d end=%d",
				diagnostic.Span.StartOffset,
				diagnostic.Span.EndOffset,
			)
		}

		switch diagnostic.Code {
		case diagCodeLintDuplicateEntry:
			hasDuplicate = true
		case diagCodeLintPluralShape:
			hasPlural = true
		case diagCodeLintMissingLanguage:
			hasLanguage = true
		}
	}

	if !hasDuplicate {
		t.Fatalf("missing %s diagnostic", diagCodeLintDuplicateEntry)
	}
	if !hasPlural {
		t.Fatalf("missing %s diagnostic", diagCodeLintPluralShape)
	}
	if !hasLanguage {
		t.Fatalf("missing %s diagnostic", diagCodeLintMissingLanguage)
	}
}

func TestLintDocumentWithOptions(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "lint/placeholder_mismatch.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}

	diagnostics := LintDocumentWithOptions(document, &LintOptions{
		CheckPlaceholders: boolPtr(false),
	})
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == diagCodeLintPrintfMismatch {
			t.Fatalf("did not expect %s with placeholders disabled", diagCodeLintPrintfMismatch)
		}
	}
}

func TestValidateDocumentStrict(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "lint/placeholder_mismatch.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}

	if err := ValidateDocument(document); err == nil {
		t.Fatal("ValidateDocument error = nil, want strict lint error")
	}
}

// boolPtr returns pointer to bool literal.
func boolPtr(value bool) *bool {
	return &value
}

func TestLintDocumentFromFixture(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "lint/placeholder_mismatch.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}

	diagnostics := LintDocument(document)
	if len(diagnostics) == 0 {
		t.Fatal("LintDocument returned no diagnostics")
	}

	var (
		hasPrintfMismatch bool
		hasPluralMissing  bool
	)
	for _, diagnostic := range diagnostics {
		switch diagnostic.Code {
		case diagCodeLintPrintfMismatch:
			hasPrintfMismatch = true
		case diagCodeLintPluralMissing:
			hasPluralMissing = true
		}
	}

	if !hasPrintfMismatch {
		t.Fatalf("missing %s diagnostic", diagCodeLintPrintfMismatch)
	}
	if !hasPluralMissing {
		t.Fatalf("missing %s diagnostic", diagCodeLintPluralMissing)
	}
}
