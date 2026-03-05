// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	diagCodeLintDuplicateEntry  = "PO2001"
	diagCodeLintPluralShape     = "PO2002"
	diagCodeLintMissingLanguage = "PO2003"
	diagCodeLintEntryWithoutID  = "PO2004"
	diagCodeLintPrintfMismatch  = "PO2005"
	diagCodeLintPluralMissing   = "PO2006"
)

// LintMode defines lint strictness profile.
type LintMode string

const (
	// LintModeBasic keeps non-critical findings as warnings.
	LintModeBasic LintMode = "basic"

	// LintModeStrict upgrades warnings to errors.
	LintModeStrict LintMode = "strict"
)

// LintOptions controls lint checks and strictness.
type LintOptions struct {
	// CheckLanguageHeader enables Language header check for translated files.
	CheckLanguageHeader *bool `json:"check_language_header,omitempty" yaml:"check_language_header,omitempty"`

	// CheckPluralShape enables plural consistency checks.
	CheckPluralShape *bool `json:"check_plural_shape,omitempty" yaml:"check_plural_shape,omitempty"`

	// CheckPlaceholders enables printf-like placeholder checks.
	CheckPlaceholders *bool `json:"check_placeholders,omitempty" yaml:"check_placeholders,omitempty"`

	// Mode controls warning/error severity policy.
	Mode LintMode `json:"mode,omitempty" yaml:"mode,omitempty"`
}

type lintSettings struct {
	Mode                LintMode
	CheckLanguageHeader bool
	CheckPluralShape    bool
	CheckPlaceholders   bool
}

// LintDocument runs basic semantic checks and returns diagnostics.
func LintDocument(document *Document) []Diagnostic {
	return LintDocumentWithOptions(document, nil)
}

// LintDocumentWithOptions runs lint with selected checks.
func LintDocumentWithOptions(
	document *Document,
	options *LintOptions,
) []Diagnostic {
	settings := normalizeLintOptions(options)
	if document == nil {
		return []Diagnostic{
			newDiagnostic(
				SeverityError,
				"PO2000",
				"document is nil",
				Position{},
			),
		}
	}

	diagnostics := make([]Diagnostic, 0)
	seen := make(map[string]Position, len(document.Entries))
	hasAnyTranslation := false

	for _, entry := range document.Entries {
		if entry == nil {
			continue
		}
		if entry.ID == "" {
			diagnostics = append(
				diagnostics,
				newDiagnostic(
					SeverityError,
					diagCodeLintEntryWithoutID,
					"entry has empty msgid",
					entry.Position,
				),
			)
			continue
		}

		key := entry.Domain + "\x00" + entry.Context + "\x00" + entry.ID
		if first, ok := seen[key]; ok {
			diagnostics = append(
				diagnostics,
				newDiagnostic(
					SeverityError,
					diagCodeLintDuplicateEntry,
					"duplicate domain/context/msgid entry (first at line "+
						strconv.Itoa(first.Line)+")",
					entry.Position,
				),
			)
		} else {
			seen[key] = entry.Position
		}

		for _, value := range entry.Translations {
			if value != "" {
				hasAnyTranslation = true
				break
			}
		}

		hasPluralIndex := false
		for index := range entry.Translations {
			if index > 0 {
				hasPluralIndex = true
				break
			}
		}
		if settings.CheckPluralShape && entry.IDPlural == "" && hasPluralIndex {
			diagnostics = append(
				diagnostics,
				newDiagnostic(
					lintWarningSeverity(settings),
					diagCodeLintPluralShape,
					"entry has msgstr[n] but no msgid_plural",
					entry.Position,
				),
			)
		}
		if settings.CheckPluralShape && entry.IDPlural != "" && !hasPluralIndex {
			diagnostics = append(
				diagnostics,
				newDiagnostic(
					lintWarningSeverity(settings),
					diagCodeLintPluralMissing,
					"entry has msgid_plural but no msgstr[n>0]",
					entry.Position,
				),
			)
		}

		if settings.CheckPlaceholders {
			diagnostics = append(
				diagnostics,
				lintPrintfPlaceholders(entry, settings)...,
			)
		}
	}

	if settings.CheckLanguageHeader &&
		hasAnyTranslation &&
		document.HeaderValue("Language") == "" {
		diagnostics = append(
			diagnostics,
			newDiagnostic(
				lintWarningSeverity(settings),
				diagCodeLintMissingLanguage,
				"document has translations but Language header is empty",
				Position{Line: 1, Column: 1, Offset: 0},
			),
		)
	}

	return diagnostics
}

// ValidateDocument runs strict lint and returns one aggregated error.
func ValidateDocument(document *Document) error {
	diagnostics := LintDocumentWithOptions(document, &LintOptions{
		Mode: LintModeStrict,
	})
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity != SeverityError {
			continue
		}

		return fmt.Errorf(
			"%s at line %d:%d: %s",
			diagnostic.Code,
			diagnostic.Position.Line,
			diagnostic.Position.Column,
			diagnostic.Message,
		)
	}

	return nil
}

// lintPrintfPlaceholders checks printf-like placeholder compatibility.
func lintPrintfPlaceholders(entry *Entry, options lintSettings) []Diagnostic {
	if entry == nil {
		return nil
	}

	diagnostics := make([]Diagnostic, 0)
	for index, translation := range entry.Translations {
		source := entry.ID
		if index > 0 && entry.IDPlural != "" {
			source = entry.IDPlural
		}

		if !hasPrintfLikeTokens(source) && !hasPrintfLikeTokens(translation) {
			continue
		}

		sourceSig := placeholderSignature(source)
		targetSig := placeholderSignature(translation)
		if signaturesEqual(sourceSig, targetSig) {
			continue
		}

		diagnostics = append(
			diagnostics,
			newDiagnostic(
				lintWarningSeverity(options),
				diagCodeLintPrintfMismatch,
				"printf-like placeholders mismatch for msgstr["+
					strconv.Itoa(index)+"]",
				entry.Position,
			),
		)
	}

	return diagnostics
}

// hasPrintfLikeTokens checks whether text looks like it has printf placeholders.
func hasPrintfLikeTokens(text string) bool {
	return strings.Contains(text, "%")
}

// placeholderSignature builds placeholder multiset by conversion verb.
func placeholderSignature(text string) map[rune]int {
	signature := make(map[rune]int)
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		if runes[i] != '%' {
			continue
		}
		if i+1 < len(runes) && runes[i+1] == '%' {
			i++
			continue
		}

		verb, consumed := parsePrintfVerb(runes[i+1:])
		if consumed == 0 {
			continue
		}

		signature[unicode.ToLower(verb)]++
		i += consumed
	}

	return signature
}

// parsePrintfVerb parses one placeholder conversion verb from tail after '%'.
func parsePrintfVerb(tail []rune) (rune, int) {
	for index, r := range tail {
		if unicode.IsLetter(r) {
			return r, index + 1
		}
	}

	return 0, 0
}

// signaturesEqual compares placeholder signatures.
func signaturesEqual(left, right map[rune]int) bool {
	if len(left) != len(right) {
		return false
	}
	for token, leftCount := range left {
		if right[token] != leftCount {
			return false
		}
	}

	return true
}

// normalizeLintOptions fills defaults for lint behavior.
func normalizeLintOptions(options *LintOptions) lintSettings {
	if options == nil {
		return lintSettings{
			Mode:                LintModeBasic,
			CheckLanguageHeader: true,
			CheckPluralShape:    true,
			CheckPlaceholders:   true,
		}
	}

	out := lintSettings{
		Mode:                options.Mode,
		CheckLanguageHeader: boolOption(options.CheckLanguageHeader, true),
		CheckPluralShape:    boolOption(options.CheckPluralShape, true),
		CheckPlaceholders:   boolOption(options.CheckPlaceholders, true),
	}
	if out.Mode == "" {
		out.Mode = LintModeBasic
	}

	return out
}

// lintWarningSeverity maps warning checks to severity by mode.
func lintWarningSeverity(options lintSettings) Severity {
	if options.Mode == LintModeStrict {
		return SeverityError
	}

	return SeverityWarning
}

// boolOption resolves optional bool with default value.
func boolOption(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}
