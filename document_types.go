// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import "strings"

// Severity defines diagnostic level.
type Severity string

const (
	// SeverityError indicates hard parse/validation failure.
	SeverityError Severity = "error"

	// SeverityWarning indicates non-fatal consistency issue.
	SeverityWarning Severity = "warning"

	// SeverityInfo indicates informational message.
	SeverityInfo Severity = "info"
)

// CommentKind classifies PO comment prefixes.
type CommentKind string

const (
	// CommentTranslator is a regular translator comment "#".
	CommentTranslator CommentKind = "translator"

	// CommentExtracted is an extracted/source comment "#.".
	CommentExtracted CommentKind = "extracted"

	// CommentReference is a source reference comment "#:".
	CommentReference CommentKind = "reference"

	// CommentFlags is a flags comment "#,".
	CommentFlags CommentKind = "flags"

	// CommentPrevious is previous-value comment "#|".
	CommentPrevious CommentKind = "previous"

	// CommentOther covers unknown comment styles.
	CommentOther CommentKind = "other"
)

// WriteMode controls output strategy.
type WriteMode string

const (
	// WriteModePreserve keeps original document order and layout as much as possible.
	WriteModePreserve WriteMode = "preserve"

	// WriteModeCanonical writes deterministic normalized output.
	WriteModeCanonical WriteMode = "canonical"
)

// Position represents 1-based line/column and byte offset.
type Position struct {
	Line   int `json:"line" yaml:"line"`
	Column int `json:"column" yaml:"column"`
	Offset int `json:"offset" yaml:"offset"`
}

// Span represents a byte range in source text.
type Span struct {
	StartOffset int `json:"start_offset" yaml:"start_offset"`
	EndOffset   int `json:"end_offset" yaml:"end_offset"`
}

// Diagnostic describes one parser/linter issue.
type Diagnostic struct {
	Severity Severity `json:"severity" yaml:"severity"`
	Code     string   `json:"code" yaml:"code"`
	Message  string   `json:"message" yaml:"message"`
	Position Position `json:"position" yaml:"position"`
	Span     Span     `json:"span" yaml:"span"`
}

// Comment stores typed PO comment line.
type Comment struct {
	Kind     CommentKind `json:"kind" yaml:"kind"`
	Text     string      `json:"text" yaml:"text"`
	Raw      string      `json:"raw" yaml:"raw"`
	Position Position    `json:"position" yaml:"position"`
}

// Header stores one parsed header key/value with source position.
type Header struct {
	Key      string   `json:"key" yaml:"key"`
	Value    string   `json:"value" yaml:"value"`
	Position Position `json:"position" yaml:"position"`
}

// Entry stores one PO translation unit.
type Entry struct {

	// Translations stores msgstr values by plural index (0 for singular).
	Translations map[int]string `json:"translations,omitempty" yaml:"translations,omitempty"`
	Domain       string         `json:"domain,omitempty" yaml:"domain,omitempty"`
	Context      string         `json:"context,omitempty" yaml:"context,omitempty"`
	ID           string         `json:"id,omitempty" yaml:"id,omitempty"`
	IDPlural     string         `json:"id_plural,omitempty" yaml:"id_plural,omitempty"`

	// Previous values come from "#| msgctxt/msgid/msgid_plural" comments.
	PreviousContext  string `json:"previous_context,omitempty" yaml:"previous_context,omitempty"`
	PreviousID       string `json:"previous_id,omitempty" yaml:"previous_id,omitempty"`
	PreviousIDPlural string `json:"previous_id_plural,omitempty" yaml:"previous_id_plural,omitempty"`

	Comments   []Comment `json:"comments,omitempty" yaml:"comments,omitempty"`
	Flags      []string  `json:"flags,omitempty" yaml:"flags,omitempty"`
	References []string  `json:"references,omitempty" yaml:"references,omitempty"`

	Position Position `json:"position" yaml:"position"`
	Obsolete bool     `json:"obsolete,omitempty" yaml:"obsolete,omitempty"`
}

// Document keeps parsed PO/POT in source order.
type Document struct {
	HeaderComments []Comment `json:"header_comments,omitempty" yaml:"header_comments,omitempty"`
	Headers        []Header  `json:"headers,omitempty" yaml:"headers,omitempty"`
	Entries        []*Entry  `json:"entries,omitempty" yaml:"entries,omitempty"`
}

// ParseOptions controls parser behavior.
type ParseOptions struct {
	// AllowInvalid keeps best-effort parse result with diagnostics.
	AllowInvalid bool `json:"allow_invalid,omitempty" yaml:"allow_invalid,omitempty"`
}

// WriteOptions controls document formatting.
type WriteOptions struct {
	// Mode selects preserve or canonical output behavior.
	Mode WriteMode `json:"mode,omitempty" yaml:"mode,omitempty"`

	// HeaderOrder defines preferred key order before generic keys in canonical mode.
	HeaderOrder []string `json:"header_order,omitempty" yaml:"header_order,omitempty"`

	// SortEntries sorts entries by domain/context/id before write.
	SortEntries bool `json:"sort_entries,omitempty" yaml:"sort_entries,omitempty"`

	// SortHeaders sorts headers by key before write.
	SortHeaders bool `json:"sort_headers,omitempty" yaml:"sort_headers,omitempty"`
}

type previousValues struct {
	Context  string
	ID       string
	IDPlural string
}

// NewDocument creates an empty document.
func NewDocument() *Document {
	return &Document{
		Headers: make([]Header, 0),
		Entries: make([]*Entry, 0),
	}
}

// HeaderValue returns header value by key, or empty string.
func (d *Document) HeaderValue(key string) string {
	if d == nil {
		return ""
	}

	for _, header := range d.Headers {
		if strings.EqualFold(header.Key, key) {
			return header.Value
		}
	}

	return ""
}
