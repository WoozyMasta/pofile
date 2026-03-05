// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"errors"
	"testing"
)

func TestBuildIndex(t *testing.T) {
	t.Parallel()

	document, err := ParseDocument(readFixture(t, "parse/domain_plural.po"))
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}

	index, err := NewIndex(document)
	if err != nil {
		t.Fatalf("NewIndex error: %v", err)
	}

	entry := index.EntryInDomain("ui", "key.ctx", "One file")
	if entry == nil {
		t.Fatal("EntryInDomain returned nil")
	}
	if got := entry.Translations[1]; got != "Много файлов" {
		t.Fatalf("msgstr[1] = %q, want %q", got, "Много файлов")
	}
}

func TestBuildIndexDuplicate(t *testing.T) {
	t.Parallel()

	document := &Document{
		Entries: []*Entry{
			{
				Context: "ctx",
				ID:      "same",
			},
			{
				Context: "ctx",
				ID:      "same",
			},
		},
	}

	_, err := NewIndex(document)
	if err == nil {
		t.Fatal("NewIndex error = nil, want duplicate key error")
	}
	if !errors.Is(err, ErrDuplicateEntryKey) {
		t.Fatalf("NewIndex error = %v, want ErrDuplicateEntryKey", err)
	}
}
