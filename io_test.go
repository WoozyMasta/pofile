// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"path/filepath"
	"testing"
)

func TestWriteFileReadFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ru.po")

	catalog := NewCatalog()
	catalog.Language = "ru"
	catalog.SetHeader("Language", "ru")
	catalog.UpsertMessage("KEY", "Text", "Текст")

	if err := WriteFile(path, catalog); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	loaded, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	loadedByParse, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	if got := loaded.Translation("KEY", "Text"); got != "Текст" {
		t.Fatalf("translation = %q, want %q", got, "Текст")
	}
	if got := loadedByParse.Translation("KEY", "Text"); got != "Текст" {
		t.Fatalf("translation(ParseFile) = %q, want %q", got, "Текст")
	}

	document, err := ParseDocumentFile(path)
	if err != nil {
		t.Fatalf("ParseDocumentFile error: %v", err)
	}
	if len(document.Entries) != 1 {
		t.Fatalf("document entries = %d, want 1", len(document.Entries))
	}
}

func TestParseDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	ru := NewCatalog()
	ru.Language = "ru"
	ru.SetHeader("Language", "ru")
	ru.UpsertMessage("KEY", "Text", "Текст")
	if err := WriteFile(filepath.Join(dir, "ru.po"), ru); err != nil {
		t.Fatalf("write ru.po: %v", err)
	}

	en := NewCatalog()
	en.Language = "en"
	en.SetHeader("Language", "en")
	en.UpsertMessage("KEY", "Text", "Text")
	if err := WriteFile(filepath.Join(dir, "en.po"), en); err != nil {
		t.Fatalf("write en.po: %v", err)
	}

	loaded, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir error: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("map len = %d, want 2", len(loaded))
	}
	if loaded["ru"] == nil || loaded["en"] == nil {
		t.Fatal("expected ru and en keys")
	}

	documents, err := ParseDocumentDir(dir)
	if err != nil {
		t.Fatalf("ParseDocumentDir error: %v", err)
	}
	if len(documents) != 2 {
		t.Fatalf("document map len = %d, want 2", len(documents))
	}
}
