// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"strings"
	"testing"
)

func TestFormatDocumentPreserve(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{
			{Key: "X-Generator", Value: "test"},
			{Key: "Language", Value: "ru"},
		},
		Entries: []*Entry{
			{
				Context: "b",
				ID:      "b",
				Translations: map[int]string{
					0: "Б",
				},
			},
			{
				Context: "a",
				ID:      "a",
				Translations: map[int]string{
					0: "А",
				},
			},
		},
	}

	data, err := FormatDocument(document, &WriteOptions{
		Mode: WriteModePreserve,
	})
	if err != nil {
		t.Fatalf("FormatDocument error: %v", err)
	}
	text := string(data)

	firstB := strings.Index(text, `msgctxt "b"`)
	firstA := strings.Index(text, `msgctxt "a"`)
	if !(firstB >= 0 && firstA > firstB) {
		t.Fatalf("entry order is not preserved:\n%s", text)
	}
}

func TestFormatDocumentCanonicalSort(t *testing.T) {
	t.Parallel()

	document := &Document{
		Headers: []Header{
			{Key: "X-Generator", Value: "test"},
			{Key: "Language", Value: "ru"},
		},
		Entries: []*Entry{
			{
				Context: "b",
				ID:      "b",
				Translations: map[int]string{
					0: "Б",
				},
			},
			{
				Context: "a",
				ID:      "a",
				Translations: map[int]string{
					0: "А",
				},
			},
		},
	}

	data, err := FormatDocument(document, &WriteOptions{
		Mode:        WriteModeCanonical,
		SortEntries: true,
		SortHeaders: true,
	})
	if err != nil {
		t.Fatalf("FormatDocument error: %v", err)
	}
	text := string(data)

	languagePos := strings.Index(text, `"Language: ru\n"`)
	generatorPos := strings.Index(text, `"X-Generator: test\n"`)
	if !(languagePos >= 0 && generatorPos > languagePos) {
		t.Fatalf("headers are not sorted canonically:\n%s", text)
	}

	firstA := strings.Index(text, `msgctxt "a"`)
	firstB := strings.Index(text, `msgctxt "b"`)
	if !(firstA >= 0 && firstB > firstA) {
		t.Fatalf("entries are not sorted canonically:\n%s", text)
	}
}
