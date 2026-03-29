// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/woozymasta/lintkit/lint"
)

const (
	// lintPublicCodePrefix is exported numeric code prefix for pofile diagnostics.
	lintPublicCodePrefix = "PO"
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

	// CheckEmptyTranslations enables warning for empty msgstr values.
	CheckEmptyTranslations *bool `json:"check_empty_translations,omitempty" yaml:"check_empty_translations,omitempty"`

	// Mode controls warning/error severity policy.
	Mode LintMode `json:"mode,omitempty" yaml:"mode,omitempty"`
}

type lintSettings struct {
	Mode                   LintMode
	CheckLanguageHeader    bool
	CheckPluralShape       bool
	CheckPlaceholders      bool
	CheckEmptyTranslations bool
}

// LintDocument runs semantic checks and returns lintkit diagnostics.
func LintDocument(document *Document) ([]lint.Diagnostic, error) {
	return LintDocumentWithOptions(document, nil)
}

// LintDocumentWithOptions runs lint with selected checks.
func LintDocumentWithOptions(
	document *Document,
	options *LintOptions,
) ([]lint.Diagnostic, error) {
	settings := normalizeLintOptions(options)
	if document == nil {
		return nil, ErrNilDocument
	}

	diagnostics := make([]lint.Diagnostic, 0)
	seenEntries := make(map[string]Position, len(document.Entries))
	seenHeaders := make(map[string]Position, len(document.Headers))
	hasAnyTranslation := false
	expectedPluralSlots, hasExpectedPluralSlots := parseNPlurals(
		document.HeaderValue("Plural-Forms"),
	)

	for _, header := range document.Headers {
		normalizedKey := normalizeHeaderKey(header.Key)
		if normalizedKey == "" {
			continue
		}

		if first, ok := seenHeaders[normalizedKey]; ok {
			diagnostics = append(
				diagnostics,
				newLintDiagnostic(
					lintWarningSeverity(settings),
					CodeLintDuplicateHeader,
					"duplicate header key (first at line "+
						strconv.Itoa(first.Line)+")",
					header.Position,
				),
			)

			continue
		}

		seenHeaders[normalizedKey] = header.Position
	}

	for _, entry := range document.Entries {
		if entry == nil {
			continue
		}
		if entry.ID == "" {
			diagnostics = append(
				diagnostics,
				newLintDiagnostic(
					lint.SeverityError,
					CodeLintEntryWithoutID,
					"entry has empty msgid",
					entry.Position,
				),
			)
			continue
		}

		key := entry.Domain + "\x00" + entry.Context + "\x00" + entry.ID
		if first, ok := seenEntries[key]; ok {
			diagnostics = append(
				diagnostics,
				newLintDiagnostic(
					lint.SeverityError,
					CodeLintDuplicateEntry,
					"duplicate domain/context/msgid entry (first at line "+
						strconv.Itoa(first.Line)+")",
					entry.Position,
				),
			)
		} else {
			seenEntries[key] = entry.Position
		}

		for _, value := range entry.Translations {
			if value != "" {
				hasAnyTranslation = true
				break
			}
		}

		if settings.CheckEmptyTranslations {
			for index, value := range entry.Translations {
				if value != "" {
					continue
				}

				diagnostics = append(
					diagnostics,
					newLintDiagnostic(
						lintWarningSeverity(settings),
						CodeLintEmptyTranslation,
						"empty translation text in msgstr["+
							strconv.Itoa(index)+"]",
						entry.Position,
					),
				)
			}
		}

		hasPluralIndex := false
		maxPluralIndex := -1
		for index := range entry.Translations {
			if index > 0 {
				hasPluralIndex = true
			}
			if index > maxPluralIndex {
				maxPluralIndex = index
			}
		}

		if settings.CheckPluralShape && entry.IDPlural == "" && hasPluralIndex {
			diagnostics = append(
				diagnostics,
				newLintDiagnostic(
					lintWarningSeverity(settings),
					CodeLintPluralShape,
					"entry has msgstr[n] but no msgid_plural",
					entry.Position,
				),
			)
		}

		if settings.CheckPluralShape && entry.IDPlural != "" && !hasPluralIndex {
			diagnostics = append(
				diagnostics,
				newLintDiagnostic(
					lintWarningSeverity(settings),
					CodeLintPluralMissing,
					"entry has msgid_plural but no msgstr[n>0]",
					entry.Position,
				),
			)
		}

		if settings.CheckPluralShape && entry.IDPlural != "" && maxPluralIndex > 0 {
			missingIndex := -1
			for index := 0; index <= maxPluralIndex; index++ {
				if _, ok := entry.Translations[index]; ok {
					continue
				}

				missingIndex = index
				break
			}

			if missingIndex >= 0 {
				diagnostics = append(
					diagnostics,
					newLintDiagnostic(
						lintWarningSeverity(settings),
						CodeLintPluralIndexGap,
						"plural index gap at msgstr["+
							strconv.Itoa(missingIndex)+"]",
						entry.Position,
					),
				)
			}
		}

		if settings.CheckPluralShape &&
			entry.IDPlural != "" &&
			hasExpectedPluralSlots &&
			hasPluralIndex {
			missingExpected := false
			for index := range expectedPluralSlots {
				if _, ok := entry.Translations[index]; ok {
					continue
				}

				missingExpected = true
				break
			}

			hasExtra := false
			for index := range entry.Translations {
				if index < expectedPluralSlots {
					continue
				}

				hasExtra = true
				break
			}

			if missingExpected || hasExtra {
				diagnostics = append(
					diagnostics,
					newLintDiagnostic(
						lintWarningSeverity(settings),
						CodeLintPluralFormsMismatch,
						"plural slots mismatch `Plural-Forms` nplurals="+
							strconv.Itoa(expectedPluralSlots),
						entry.Position,
					),
				)
			}
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
			newLintDiagnostic(
				lintWarningSeverity(settings),
				CodeLintMissingLanguage,
				"document has translations but Language header is empty",
				Position{Line: 1, Column: 1, Offset: 0},
			),
		)
	}

	return diagnostics, nil
}

// ValidateDocument runs strict lint and returns one aggregated error.
func ValidateDocument(document *Document) error {
	diagnostics, err := LintDocumentWithOptions(document, &LintOptions{
		Mode: LintModeStrict,
	})
	if err != nil {
		return err
	}

	return lint.ErrorFromDiagnostics(diagnostics, lint.SeverityError)
}

// lintPrintfPlaceholders checks printf-like placeholder compatibility.
func lintPrintfPlaceholders(
	entry *Entry,
	options lintSettings,
) []lint.Diagnostic {
	if entry == nil {
		return nil
	}

	diagnostics := make([]lint.Diagnostic, 0)
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
			newLintDiagnostic(
				lintWarningSeverity(options),
				CodeLintPrintfMismatch,
				"printf-like placeholders mismatch for msgstr["+
					strconv.Itoa(index)+"]",
				entry.Position,
			),
		)
	}

	return diagnostics
}

// newLintDiagnostic builds one lintkit diagnostic with stable code/rule metadata.
func newLintDiagnostic(
	severity lint.Severity,
	code lint.Code,
	message string,
	position Position,
) lint.Diagnostic {
	start := lintPosition(position)
	end := start
	end.Offset = start.Offset + 1

	return lint.Diagnostic{
		RuleID:   LintRuleID(code),
		Code:     publicLintCode(code),
		Severity: severity,
		Message:  message,
		Start:    start,
		End:      end,
	}
}

// lintPosition converts pofile position into lintkit position.
func lintPosition(position Position) lint.Position {
	return lint.Position{
		Line:   position.Line,
		Column: position.Column,
		Offset: position.Offset,
	}
}

// publicLintCode formats exported lint code token with module prefix.
func publicLintCode(code lint.Code) string {
	digits := lint.FormatCode(code)
	if digits == "" {
		return ""
	}

	return lintPublicCodePrefix + digits
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
			Mode:                   LintModeBasic,
			CheckLanguageHeader:    true,
			CheckPluralShape:       true,
			CheckPlaceholders:      true,
			CheckEmptyTranslations: false,
		}
	}

	out := lintSettings{
		Mode:                   options.Mode,
		CheckLanguageHeader:    boolOption(options.CheckLanguageHeader, true),
		CheckPluralShape:       boolOption(options.CheckPluralShape, true),
		CheckPlaceholders:      boolOption(options.CheckPlaceholders, true),
		CheckEmptyTranslations: boolOption(options.CheckEmptyTranslations, false),
	}
	if out.Mode == "" {
		out.Mode = LintModeBasic
	}

	return out
}

// lintWarningSeverity maps warning checks to severity by mode.
func lintWarningSeverity(options lintSettings) lint.Severity {
	if options.Mode == LintModeStrict {
		return lint.SeverityError
	}

	return lint.SeverityWarning
}

// boolOption resolves optional bool with default value.
func boolOption(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

// normalizeHeaderKey normalizes header keys for duplicate detection.
func normalizeHeaderKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

// parseNPlurals extracts nplurals count from Plural-Forms header value.
func parseNPlurals(value string) (int, bool) {
	if value == "" {
		return 0, false
	}

	parts := strings.SplitSeq(value, ";")
	for part := range parts {
		trimmed := strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(trimmed), "nplurals=") {
			continue
		}

		rawCount := strings.TrimSpace(trimmed[len("nplurals="):])
		count, err := strconv.Atoi(rawCount)
		if err != nil || count <= 0 {
			return 0, false
		}

		return count, true
	}

	return 0, false
}
