// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import "github.com/woozymasta/lintkit/lint"

// withDescription attaches optional documentation text to one catalog spec.
func withDescription(spec lint.CodeSpec, description string) lint.CodeSpec {
	spec.Description = description

	return spec
}

// diagnosticCatalog stores stable diagnostics metadata table.
var diagnosticCatalog = []lint.CodeSpec{
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseMissingQuote,
			StageParse,
			"directive value must be quoted",
		),
		"PO directives such as `msgid`, `msgstr`, `msgctxt`, and `msgid_plural` "+
			"must use quoted string value on the same logical line.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseUnknownLine,
			StageParse,
			"unknown PO directive or malformed line",
		),
		"Parser found non-empty line that is not recognized as PO keyword, "+
			"comment, or valid continuation content.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseBadContinuation,
			StageParse,
			"continuation string is outside active field",
		),
		"String continuation line must follow active field (`msgid`, `msgstr`, "+
			"`msgctxt`, or `msgid_plural`) and cannot appear standalone.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseBadHeader,
			StageParse,
			"`header` metadata line is malformed",
		),
		"`header` metadata line must follow supported syntax with valid key/value "+
			"shape for this parser.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseBadMsgStrIndex,
			StageParse,
			"`msgstr[n]` index must be non-negative integer",
		),
		"Plural translation form must use integer index notation `msgstr[n]` "+
			"with non-negative numeric `n`.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeParseMissingMsgID,
			StageParse,
			"entry is missing required `msgid`",
		),
		"Every non-obsolete PO entry must contain source identifier `msgid`.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeLintDuplicateEntry,
			StageLint,
			"duplicate `domain/msgctxt/msgid` entry key",
		),
		"More than one entry has same `domain` + `msgctxt` + `msgid` key. Keep "+
			"single canonical entry per key.",
	),
	withDescription(
		lint.ErrorCodeSpec(
			CodeLintEntryWithoutID,
			StageLint,
			"entry has empty `msgid` value",
		),
		"Entry is present but source key is empty. Remove broken entry or fill "+
			"valid `msgid` value.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintPluralShape,
			StageLint,
			"`msgstr[n]` present but `msgid_plural` is missing",
		),
		"Plural translations require plural source form. Add `msgid_plural` when "+
			"`msgstr[1]` or higher indexes are present.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintPluralMissing,
			StageLint,
			"`msgid_plural` present but plural `msgstr[n]` is incomplete",
		),
		"Plural source form exists but translated plural slots are incomplete. Add "+
			"at least `msgstr[1]` and other required indexes for locale rules.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintPluralIndexGap,
			StageLint,
			"plural indexes have gap",
		),
		"Plural entry has missing `msgstr[n]` between existing indexes. Use "+
			"continuous index range without holes.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintPluralFormsMismatch,
			StageLint,
			"`msgstr[n]` count mismatches `Plural-Forms`",
		),
		"`Plural-Forms` header declares `nplurals`, but entry translations do not "+
			"match expected plural slot count.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintDuplicateHeader,
			StageLint,
			"duplicate header key",
		),
		"Header contains same key more than once. Keep one canonical key/value "+
			"entry to avoid parser-dependent behavior.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintMissingLanguage,
			StageLint,
			"translations exist but `Language` header is empty",
		),
		"File contains translated strings but header does not declare target "+
			"language.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintPrintfMismatch,
			StageLint,
			"printf placeholders mismatch",
		),
		"Set of printf-style verbs differs between source and translated string. "+
			"Runtime formatting may fail or substitute wrong arguments.",
	),
	withDescription(
		lint.WarningCodeSpec(
			CodeLintEmptyTranslation,
			StageLint,
			"entry has empty translation text",
		),
		"Translated entry contains empty `msgstr` value. Enable this check only "+
			"when empty translations should be treated as potential issues.",
	),
}
