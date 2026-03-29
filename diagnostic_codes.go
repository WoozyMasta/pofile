// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"github.com/woozymasta/lintkit/lint"
)

const (
	// LintModule is stable lint module namespace for pofile rules.
	LintModule = "pofile"
)

const (
	// StageParse marks parser diagnostics.
	StageParse lint.Stage = "parse"

	// StageLint marks semantic lint diagnostics.
	StageLint lint.Stage = "lint"
)

const (
	// CodeParseMissingQuote reports malformed quoted value syntax.
	CodeParseMissingQuote lint.Code = 1001

	// CodeParseUnknownLine reports unknown PO directive line.
	CodeParseUnknownLine lint.Code = 1002

	// CodeParseBadContinuation reports continuation without active section.
	CodeParseBadContinuation lint.Code = 1003

	// CodeParseBadHeader reports malformed header entry.
	CodeParseBadHeader lint.Code = 1004

	// CodeParseBadMsgStrIndex reports invalid msgstr[n] index.
	CodeParseBadMsgStrIndex lint.Code = 1005

	// CodeParseMissingMsgID reports entry with missing msgid.
	CodeParseMissingMsgID lint.Code = 1006
)

const (
	// CodeLintDuplicateEntry reports duplicate domain/context/msgid.
	CodeLintDuplicateEntry lint.Code = 2001

	// CodeLintEntryWithoutID reports entry with empty msgid.
	CodeLintEntryWithoutID lint.Code = 2002

	// CodeLintPluralShape reports missing msgid_plural with msgstr[n].
	CodeLintPluralShape lint.Code = 2003

	// CodeLintPluralMissing reports missing msgstr[n>0] for plural entry.
	CodeLintPluralMissing lint.Code = 2004

	// CodeLintPluralIndexGap reports missing plural translation index in range.
	CodeLintPluralIndexGap lint.Code = 2005

	// CodeLintPluralFormsMismatch reports mismatch with Plural-Forms nplurals.
	CodeLintPluralFormsMismatch lint.Code = 2006

	// CodeLintDuplicateHeader reports duplicate header key entries.
	CodeLintDuplicateHeader lint.Code = 2007

	// CodeLintMissingLanguage reports missing Language header.
	CodeLintMissingLanguage lint.Code = 2008

	// CodeLintPrintfMismatch reports placeholder mismatch in translation.
	CodeLintPrintfMismatch lint.Code = 2009

	// CodeLintEmptyTranslation reports empty translation text.
	CodeLintEmptyTranslation lint.Code = 2010
)

var diagnosticCodeCatalogHandle = lint.NewCodeCatalogHandle(
	lint.CodeCatalogConfig{
		Module:            LintModule,
		CodePrefix:        "PO",
		ModuleName:        "PO File",
		ModuleDescription: "Lint rules for Gettext PO parsing and semantic checks.",
		ScopeDescriptions: map[lint.Stage]string{
			StageParse: "PO parser diagnostics.",
			StageLint:  "PO semantic lint diagnostics.",
		},
	},
	diagnosticCatalog,
)

// getDiagnosticCodeCatalog returns lazy-initialized code catalog helper.
func getDiagnosticCodeCatalog() (lint.CodeCatalog, error) {
	return diagnosticCodeCatalogHandle.Catalog()
}

// DiagnosticRuleSpec converts one diagnostic spec into lint rule metadata.
func DiagnosticRuleSpec(spec lint.CodeSpec) (lint.RuleSpec, error) {
	return diagnosticCodeCatalogHandle.RuleSpec(spec)
}

// LintRuleID returns lint rule ID mapped from stable pofile diagnostic code.
func LintRuleID(code lint.Code) string {
	return diagnosticCodeCatalogHandle.RuleIDOrUnknown(code)
}

// DiagnosticCatalog returns stable diagnostics metadata list.
func DiagnosticCatalog() []lint.CodeSpec {
	return diagnosticCodeCatalogHandle.CodeSpecs()
}

// DiagnosticByCode returns diagnostic metadata for code.
func DiagnosticByCode(code lint.Code) (lint.CodeSpec, bool) {
	return diagnosticCodeCatalogHandle.ByCode(code)
}

// LintRuleSpecs returns deterministic lint rule specs from diagnostics catalog.
func LintRuleSpecs() []lint.RuleSpec {
	return diagnosticCodeCatalogHandle.RuleSpecs()
}
