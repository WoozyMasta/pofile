// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"strings"
	"testing"
)

const benchmarkInput = `msgid ""
msgstr ""
"Project-Id-Version: Demo\n"
"Language: ru\n"
"MIME-Version: 1.0\n"

#, notranslate
msgctxt "UI_BUTTON_OK"
msgid "OK"
msgstr "Ок"

msgctxt "UI_BUTTON_CANCEL"
msgid "Cancel"
msgstr ""
`

// BenchmarkParseReader benchmarks parse flow from reader.
func BenchmarkParseReader(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := ParseReader(strings.NewReader(benchmarkInput))
		if err != nil {
			b.Fatalf("ParseReader error: %v", err)
		}
	}
}

// BenchmarkFormat benchmarks write/format flow.
func BenchmarkFormat(b *testing.B) {
	catalog, err := ParseReader(strings.NewReader(benchmarkInput))
	if err != nil {
		b.Fatalf("ParseReader setup error: %v", err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_, err := Format(catalog)
		if err != nil {
			b.Fatalf("Format error: %v", err)
		}
	}
}

// BenchmarkMergeTemplate benchmarks top-level preprocess flow.
func BenchmarkMergeTemplate(b *testing.B) {
	template, err := ParseReader(strings.NewReader(benchmarkInput))
	if err != nil {
		b.Fatalf("ParseReader setup error: %v", err)
	}
	existing := template.Clone()
	existing.UpsertMessage("UI_BUTTON_CANCEL", "Cancel", "Отмена")
	existing.SetHeader("Language", "ru")
	existing.Language = "ru"

	b.ReportAllocs()
	for b.Loop() {
		_, err := MergeTemplate(template, existing)
		if err != nil {
			b.Fatalf("MergeTemplate error: %v", err)
		}
	}
}

// BenchmarkParseAndLint benchmarks parse + lint flow.
func BenchmarkParseAndLint(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		document, err := ParseDocument([]byte(benchmarkInput))
		if err != nil {
			b.Fatalf("ParseDocument error: %v", err)
		}
		if _, err := LintDocument(document); err != nil {
			b.Fatalf("LintDocument error: %v", err)
		}
	}
}
