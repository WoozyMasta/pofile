// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"errors"
	"strings"
	"testing"
)

func TestNewCatalog(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	if catalog == nil {
		t.Fatal("NewCatalog() returned nil")
	}
	if catalog.Headers == nil {
		t.Fatal("Headers map is nil")
	}
	if catalog.Messages == nil {
		t.Fatal("Messages slice is nil")
	}
	if len(catalog.Messages) != 0 {
		t.Fatalf("messages len = %d, want 0", len(catalog.Messages))
	}
}

func TestSetMessageUpdate(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	catalog.UpsertMessage("ctx", "msg", "first")
	catalog.UpsertMessage("ctx", "msg", "second")

	if len(catalog.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(catalog.Messages))
	}

	message := catalog.FindMessage("ctx", "msg")
	if message == nil {
		t.Fatal("FindMessage returned nil")
	}
	if message.TranslationAt(0) != "second" {
		t.Fatalf("translation = %q, want %q", message.TranslationAt(0), "second")
	}
}

func TestDeleteMessage(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	catalog.UpsertMessage("ctx", "a", "A")
	catalog.UpsertMessage("ctx", "b", "B")

	if deleted := catalog.DeleteMessage("ctx", "a"); !deleted {
		t.Fatal("DeleteMessage returned false, want true")
	}
	if got := catalog.FindMessage("ctx", "a"); got != nil {
		t.Fatal("deleted message is still present")
	}
	if deleted := catalog.DeleteMessage("ctx", "missing"); deleted {
		t.Fatal("DeleteMessage returned true for missing key")
	}
}

func TestTranslationLookup(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	catalog.UpsertMessage("ctx", "hello", "privet")

	if got := catalog.Translation("ctx", "hello"); got != "privet" {
		t.Fatalf("Translation = %q, want %q", got, "privet")
	}
	if got := catalog.Translation("ctx", "missing"); got != "" {
		t.Fatalf("Translation(missing) = %q, want empty", got)
	}
	if !catalog.IsTranslated("ctx", "hello") {
		t.Fatal("IsTranslated should be true")
	}
	if catalog.IsTranslated("ctx", "missing") {
		t.Fatal("IsTranslated should be false")
	}
}

func TestHasFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []string
		comments []string
		flag     string
		want     bool
	}{
		{
			name:  "parsed flags field",
			flags: []string{"fuzzy", "notranslate"},
			flag:  "notranslate",
			want:  true,
		},
		{
			name:     "flag comment style",
			comments: []string{"#, fuzzy, notranslate"},
			flag:     "notranslate",
			want:     true,
		},
		{
			name:     "translator comment is ignored",
			comments: []string{"# notranslate"},
			flag:     "notranslate",
			want:     false,
		},
		{
			name:     "missing flag",
			comments: []string{"# normal comment"},
			flag:     "notranslate",
			want:     false,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			message := &Message{
				Flags:    testCase.flags,
				Comments: testCase.comments,
			}
			if got := message.HasFlag(testCase.flag); got != testCase.want {
				t.Fatalf("HasFlag = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestParseReaderBasic(t *testing.T) {
	t.Parallel()

	input := `msgid ""
msgstr ""
"Project-Id-Version: test\n"
"Language: ru\n"

msgctxt "KEY1"
msgid "Original"
msgstr "Translated"
`

	catalog, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseReader error: %v", err)
	}

	if got := catalog.Header("Project-Id-Version"); got != "test" {
		t.Fatalf("Project-Id-Version = %q, want %q", got, "test")
	}
	if got := catalog.Language; got != "ru" {
		t.Fatalf("Language = %q, want %q", got, "ru")
	}
	if len(catalog.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(catalog.Messages))
	}

	message := catalog.Messages[0]
	if message.Context != "KEY1" ||
		message.ID != "Original" ||
		message.TranslationAt(0) != "Translated" {
		t.Fatalf(
			"message mismatch: got (%q, %q, %q)",
			message.Context,
			message.ID,
			message.TranslationAt(0),
		)
	}
}

func TestParseReaderComments(t *testing.T) {
	t.Parallel()

	input := `msgid ""
msgstr ""

# some comment
#, notranslate
msgctxt "KEY1"
msgid "Text"
msgstr ""
`

	catalog, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseReader error: %v", err)
	}

	message := catalog.FindMessage("KEY1", "Text")
	if message == nil {
		t.Fatal("FindMessage returned nil")
	}
	if len(message.Comments) != 2 {
		t.Fatalf("comments len = %d, want 2", len(message.Comments))
	}
	if !message.HasFlag("notranslate") {
		t.Fatal(`HasFlag("notranslate") returned false`)
	}
}

func TestFormatRoundTrip(t *testing.T) {
	t.Parallel()

	original := NewCatalog()
	original.SetHeader("Language", "ru")
	original.Language = "ru"
	entry := original.UpsertMessage("KEY1", "Text 1", "Текст 1")
	entry.Comments = []string{"# comment 1"}
	original.UpsertMessage("KEY2", "Text 2", "")

	data, err := Format(original)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if got := parsed.Header("Language"); got != "ru" {
		t.Fatalf("Language header = %q, want %q", got, "ru")
	}
	if len(parsed.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(parsed.Messages))
	}
	if got := parsed.Translation("KEY1", "Text 1"); got != "Текст 1" {
		t.Fatalf("translation = %q, want %q", got, "Текст 1")
	}
}

func TestFormatRoundTripMetadataWithRawComments(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	message := catalog.UpsertMessage("KEY1", "Text 1", "")
	message.Comments = []string{"# translator note"}
	message.Flags = []string{"notranslate"}
	message.References = []string{"file.cpp:10"}
	message.PreviousContext = "OldCtx"
	message.PreviousID = `Old "text"`
	message.PreviousIDPlural = "Old plural"

	data, err := Format(catalog)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := parsed.FindMessage("KEY1", "Text 1")
	if got == nil {
		t.Fatal("FindMessage returned nil")
	}
	if !got.HasFlag("notranslate") {
		t.Fatal(`HasFlag("notranslate") returned false`)
	}
	if len(got.References) != 1 || got.References[0] != "file.cpp:10" {
		t.Fatalf("references = %#v, want [file.cpp:10]", got.References)
	}
	if got.PreviousContext != "OldCtx" {
		t.Fatalf("PreviousContext = %q, want %q", got.PreviousContext, "OldCtx")
	}
	if got.PreviousID != `Old "text"` {
		t.Fatalf("PreviousID = %q, want %q", got.PreviousID, `Old "text"`)
	}
	if got.PreviousIDPlural != "Old plural" {
		t.Fatalf("PreviousIDPlural = %q, want %q", got.PreviousIDPlural, "Old plural")
	}
}

func TestMergeTemplate(t *testing.T) {
	t.Parallel()

	template := NewCatalog()
	template.SetHeader("MIME-Version", "1.0")
	template.UpsertMessage("A", "Hello", "")
	template.UpsertMessage("B", "World", "")

	existing := NewCatalog()
	existing.Language = "ru"
	existing.SetHeader("Language", "ru")
	existing.SetHeader("Last-Translator", "Test User")
	entry := existing.UpsertMessage("A", "Hello", "Привет")
	entry.Comments = []string{"# keep me"}
	existing.UpsertMessage("C", "Legacy", "Легаси")

	merged, err := MergeTemplate(template, existing)
	if err != nil {
		t.Fatalf("MergeTemplate error: %v", err)
	}

	if merged.Language != "ru" {
		t.Fatalf("language = %q, want %q", merged.Language, "ru")
	}
	if got := merged.Header("Last-Translator"); got != "Test User" {
		t.Fatalf("Last-Translator = %q, want %q", got, "Test User")
	}
	if got := merged.Translation("A", "Hello"); got != "Привет" {
		t.Fatalf("translation = %q, want %q", got, "Привет")
	}
	if merged.FindMessage("C", "Legacy") != nil {
		t.Fatal("legacy message should not be included")
	}
}

func TestValidateDuplicateMessage(t *testing.T) {
	t.Parallel()

	catalog := NewCatalog()
	catalog.Messages = append(catalog.Messages,
		&Message{Context: "A", ID: "B"},
		&Message{Context: "A", ID: "B"},
	)

	err := catalog.Validate()
	if err == nil {
		t.Fatal("Validate error = nil, want duplicate error")
	}
	if !errors.Is(err, ErrDuplicateMessage) {
		t.Fatalf("Validate error = %v, want ErrDuplicateMessage", err)
	}
}
