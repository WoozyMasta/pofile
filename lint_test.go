// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

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

	diagnostics, err := LintDocument(document)
	if err != nil {
		t.Fatalf("LintDocument error: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("LintDocument returned no diagnostics")
	}

	var (
		hasDuplicate bool
		hasPlural    bool
		hasLanguage  bool
	)
	for _, diagnostic := range diagnostics {
		if diagnostic.End.Offset < diagnostic.Start.Offset {
			t.Fatalf(
				"invalid span offsets: start=%d end=%d",
				diagnostic.Start.Offset,
				diagnostic.End.Offset,
			)
		}

		switch parseDiagnosticCode(diagnostic) {
		case CodeLintDuplicateEntry:
			hasDuplicate = true
		case CodeLintPluralShape:
			hasPlural = true
		case CodeLintMissingLanguage:
			hasLanguage = true
		}
	}

	if !hasDuplicate {
		t.Fatalf("missing %d diagnostic", CodeLintDuplicateEntry)
	}
	if !hasPlural {
		t.Fatalf("missing %d diagnostic", CodeLintPluralShape)
	}
	if !hasLanguage {
		t.Fatalf("missing %d diagnostic", CodeLintMissingLanguage)
	}
}

func TestLintDocumentWithOptions(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "lint/placeholder_mismatch.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}

	diagnostics, lintErr := LintDocumentWithOptions(document, &LintOptions{
		CheckPlaceholders: boolPtr(false),
	})
	if lintErr != nil {
		t.Fatalf("LintDocumentWithOptions error: %v", lintErr)
	}
	for _, diagnostic := range diagnostics {
		if parseDiagnosticCode(diagnostic) == CodeLintPrintfMismatch {
			t.Fatalf(
				"did not expect %d with placeholders disabled",
				CodeLintPrintfMismatch,
			)
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

	diagnostics, lintErr := LintDocument(document)
	if lintErr != nil {
		t.Fatalf("LintDocument error: %v", lintErr)
	}
	if len(diagnostics) == 0 {
		t.Fatal("LintDocument returned no diagnostics")
	}

	var (
		hasPrintfMismatch bool
		hasPluralMissing  bool
	)
	for _, diagnostic := range diagnostics {
		switch parseDiagnosticCode(diagnostic) {
		case CodeLintPrintfMismatch:
			hasPrintfMismatch = true
		case CodeLintPluralMissing:
			hasPluralMissing = true
		}
	}

	if !hasPrintfMismatch {
		t.Fatalf("missing %d diagnostic", CodeLintPrintfMismatch)
	}
	if !hasPluralMissing {
		t.Fatalf("missing %d diagnostic", CodeLintPluralMissing)
	}
}

func TestLintDocumentNilDocument(t *testing.T) {
	t.Parallel()

	_, err := LintDocument(nil)
	if err == nil {
		t.Fatal("LintDocument(nil) error = nil, want ErrNilDocument")
	}
	if err != ErrNilDocument {
		t.Fatalf("LintDocument(nil) error = %v, want %v", err, ErrNilDocument)
	}
}

func TestLintDocumentDetectsDuplicateHeader(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{
			{Key: "Language", Value: "ru", Position: Position{Line: 1, Column: 1}},
			{Key: "language", Value: "en", Position: Position{Line: 2, Column: 1}},
		},
		Entries: []*Entry{},
	}

	diagnostics, err := LintDocument(document)
	if err != nil {
		t.Fatalf("LintDocument error: %v", err)
	}
	if !hasDiagnosticCode(diagnostics, CodeLintDuplicateHeader) {
		t.Fatalf("missing %d diagnostic", CodeLintDuplicateHeader)
	}
}

func TestLintDocumentDetectsPluralIndexGap(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{},
		Entries: []*Entry{
			{
				ID:       "id",
				IDPlural: "ids",
				Translations: map[int]string{
					0: "one",
					2: "many",
				},
			},
		},
	}

	diagnostics, err := LintDocument(document)
	if err != nil {
		t.Fatalf("LintDocument error: %v", err)
	}
	if !hasDiagnosticCode(diagnostics, CodeLintPluralIndexGap) {
		t.Fatalf("missing %d diagnostic", CodeLintPluralIndexGap)
	}
}

func TestLintDocumentDetectsPluralFormsMismatch(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{
			{
				Key:   "Plural-Forms",
				Value: "nplurals=3; plural=(n%10==1 && n%100!=11 ? 0 : 1);",
			},
		},
		Entries: []*Entry{
			{
				ID:       "id",
				IDPlural: "ids",
				Translations: map[int]string{
					0: "one",
					1: "few",
				},
			},
		},
	}

	diagnostics, err := LintDocument(document)
	if err != nil {
		t.Fatalf("LintDocument error: %v", err)
	}
	if !hasDiagnosticCode(diagnostics, CodeLintPluralFormsMismatch) {
		t.Fatalf("missing %d diagnostic", CodeLintPluralFormsMismatch)
	}
}

func TestLintDocumentEmptyTranslationDisabledByDefault(t *testing.T) {
	t.Parallel()

	document := &Document{
		Entries: []*Entry{
			{
				ID: "id",
				Translations: map[int]string{
					0: "",
				},
			},
		},
	}

	diagnostics, err := LintDocument(document)
	if err != nil {
		t.Fatalf("LintDocument error: %v", err)
	}
	if hasDiagnosticCode(diagnostics, CodeLintEmptyTranslation) {
		t.Fatalf("unexpected %d diagnostic", CodeLintEmptyTranslation)
	}
}

func TestLintDocumentEmptyTranslationEnabled(t *testing.T) {
	t.Parallel()

	document := &Document{
		Entries: []*Entry{
			{
				ID: "id",
				Translations: map[int]string{
					0: "",
				},
			},
		},
	}

	diagnostics, err := LintDocumentWithOptions(document, &LintOptions{
		CheckEmptyTranslations: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("LintDocumentWithOptions error: %v", err)
	}
	if !hasDiagnosticCode(diagnostics, CodeLintEmptyTranslation) {
		t.Fatalf("missing %d diagnostic", CodeLintEmptyTranslation)
	}
}

// hasDiagnosticCode reports whether one code exists in diagnostics slice.
func hasDiagnosticCode(diagnostics []lint.Diagnostic, code lint.Code) bool {
	for _, diagnostic := range diagnostics {
		if parseDiagnosticCode(diagnostic) == code {
			return true
		}
	}

	return false
}

// parseDiagnosticCode parses one exported diagnostic code token.
func parseDiagnosticCode(diagnostic lint.Diagnostic) lint.Code {
	code, ok := lint.ParsePublicCode(diagnostic.Code)
	if !ok {
		return 0
	}

	return code
}
