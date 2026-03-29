// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

const (
	parseBufferSize = 512
)

// ParseDocument parses PO/POT bytes into lossless document.
func ParseDocument(data []byte) (*Document, error) {
	document, diagnostics, err := ParseDocumentWithOptions(
		data,
		ParseOptions{},
	)
	if err != nil {
		return nil, err
	}
	if hasErrorDiagnostics(diagnostics) {
		return nil, formatParseError(diagnostics)
	}

	return document, nil
}

// ParseDocumentWithOptions parses PO/POT bytes and returns diagnostics.
func ParseDocumentWithOptions(
	data []byte,
	options ParseOptions,
) (*Document, []lint.Diagnostic, error) {
	return ParseDocumentReaderWithOptions(bytes.NewReader(data), options)
}

// ParseDocumentReader parses PO/POT from reader into lossless document.
func ParseDocumentReader(reader io.Reader) (*Document, error) {
	document, diagnostics, err := ParseDocumentReaderWithOptions(
		reader,
		ParseOptions{},
	)
	if err != nil {
		return nil, err
	}
	if hasErrorDiagnostics(diagnostics) {
		return nil, formatParseError(diagnostics)
	}

	return document, nil
}

// ParseDocumentReaderWithOptions parses reader and returns diagnostics.
func ParseDocumentReaderWithOptions(
	reader io.Reader,
	options ParseOptions,
) (*Document, []lint.Diagnostic, error) {
	state := parseState{
		document:      NewDocument(),
		options:       options,
		currentDomain: "",
		pending:       make([]Comment, 0),
		diagnostics:   make([]lint.Diagnostic, 0),
	}

	br := bufio.NewReaderSize(reader, parseBufferSize)
	for {
		rawLine, readErr := br.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return nil, nil, fmt.Errorf("read po: %w", readErr)
		}
		if readErr == io.EOF && rawLine == "" {
			break
		}

		state.line++
		state.offset += len(rawLine)
		line := strings.TrimRight(rawLine, "\r\n")
		trimmed := strings.TrimSpace(line)
		position := Position{
			Line:   state.line,
			Column: lineColumn(line),
			Offset: state.offset - len(rawLine),
		}

		if trimmed == "" {
			state.flushCurrent()
			state.section = ""
			if readErr == io.EOF {
				break
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			state.parseComment(trimmed, position)
			if readErr == io.EOF {
				break
			}
			continue
		}

		state.parseDirective(trimmed, position, false)
		if readErr == io.EOF {
			break
		}
	}

	state.flushCurrent()
	state.addMissingMsgIDDiagnostics()

	if hasErrorDiagnostics(state.diagnostics) && !options.AllowInvalid {
		return state.document, state.diagnostics, formatParseError(state.diagnostics)
	}

	return state.document, state.diagnostics, nil
}

type parseState struct {
	document      *Document
	current       *Entry
	pendingPrev   previousValues
	section       string
	currentDomain string
	diagnostics   []lint.Diagnostic
	pending       []Comment
	sectionIndex  int
	line          int
	offset        int
	options       ParseOptions
	headerSeen    bool
}

// parseComment parses a comment or an obsolete directive.
func (s *parseState) parseComment(trimmed string, position Position) {
	if obsoletePayload, ok := strings.CutPrefix(trimmed, "#~"); ok {
		obsoletePayload = strings.TrimSpace(obsoletePayload)
		if obsoletePayload == "" {
			comment := Comment{
				Kind:     CommentOther,
				Text:     "",
				Raw:      trimmed,
				Position: position,
			}
			s.pending = append(s.pending, comment)
			return
		}

		if strings.HasPrefix(obsoletePayload, "#") {
			comment := parseCommentLine(obsoletePayload, position)
			comment.Raw = trimmed
			s.pending = append(s.pending, comment)
			s.applyCommentMetadata(comment)
			return
		}

		s.parseDirective(obsoletePayload, position, true)
		return
	}

	comment := parseCommentLine(trimmed, position)
	s.pending = append(s.pending, comment)
	s.applyCommentMetadata(comment)
}

// applyCommentMetadata collects flags/reference/previous metadata.
func (s *parseState) applyCommentMetadata(comment Comment) {
	if comment.Kind != CommentPrevious {
		return
	}

	switch {
	case strings.HasPrefix(comment.Text, "msgctxt "):
		value, ok := extractQuotedValueChecked(comment.Text)
		if ok {
			s.pendingPrev.Context = value
		}
	case strings.HasPrefix(comment.Text, "msgid_plural "):
		value, ok := extractQuotedValueChecked(comment.Text)
		if ok {
			s.pendingPrev.IDPlural = value
		}
	case strings.HasPrefix(comment.Text, "msgid "):
		value, ok := extractQuotedValueChecked(comment.Text)
		if ok {
			s.pendingPrev.ID = value
		}
	}
}

// parseDirective parses one non-comment PO line.
func (s *parseState) parseDirective(
	trimmed string,
	position Position,
	obsolete bool,
) {
	switch {
	case strings.HasPrefix(trimmed, "domain "):
		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				"domain must contain quoted value",
				position,
			)
			return
		}
		s.currentDomain = value

	case strings.HasPrefix(trimmed, "msgctxt "):
		entry := s.ensureEntry(true, position, obsolete)
		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				`msgctxt must contain quoted value`,
				position,
			)
			return
		}
		entry.Context = value
		s.section = "msgctxt"

	case strings.HasPrefix(trimmed, "msgid_plural "):
		entry := s.ensureEntry(false, position, obsolete)
		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				`msgid_plural must contain quoted value`,
				position,
			)
			return
		}
		entry.IDPlural = value
		s.section = "msgid_plural"

	case strings.HasPrefix(trimmed, "msgid "):
		entry := s.ensureEntry(true, position, obsolete)
		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				`msgid must contain quoted value`,
				position,
			)
			return
		}
		entry.ID = value
		s.section = "msgid"

	case strings.HasPrefix(trimmed, "msgstr["):
		entry := s.ensureEntry(false, position, obsolete)
		index, value, ok := parseIndexedMsgStr(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseBadMsgStrIndex,
				`invalid msgstr[n] form`,
				position,
			)
			return
		}
		entry.Translations[index] = value
		s.section = "msgstr"
		s.sectionIndex = index

	case strings.HasPrefix(trimmed, "msgstr "):
		entry := s.ensureEntry(false, position, obsolete)
		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				`msgstr must contain quoted value`,
				position,
			)
			return
		}
		entry.Translations[0] = value
		s.section = "msgstr"
		s.sectionIndex = 0

	case strings.HasPrefix(trimmed, `"`):
		if s.current == nil || s.section == "" {
			s.addDiagnostic(
				CodeParseBadContinuation,
				"continuation string has no active section",
				position,
			)
			return
		}

		value, ok := extractQuotedValueChecked(trimmed)
		if !ok {
			s.addDiagnostic(
				CodeParseMissingQuote,
				`continuation line must be quoted`,
				position,
			)
			return
		}
		s.appendContinuation(value)

	default:
		s.addDiagnostic(
			CodeParseUnknownLine,
			"unknown PO directive",
			position,
		)
	}
}

// ensureEntry returns current entry or creates a new one.
func (s *parseState) ensureEntry(
	startNewOnExisting bool,
	position Position,
	obsolete bool,
) *Entry {
	if s.current != nil && startNewOnExisting && s.currentHasData() {
		s.flushCurrent()
	}
	if s.current == nil {
		s.current = &Entry{
			Domain:       s.currentDomain,
			Obsolete:     obsolete,
			Position:     position,
			Comments:     append([]Comment(nil), s.pending...),
			Translations: make(map[int]string),
		}
		s.current.PreviousContext = s.pendingPrev.Context
		s.current.PreviousID = s.pendingPrev.ID
		s.current.PreviousIDPlural = s.pendingPrev.IDPlural
		s.pending = s.pending[:0]
		s.pendingPrev = previousValues{}
		collectEntryCommentMetadata(s.current)
	} else if obsolete {
		s.current.Obsolete = true
	}

	return s.current
}

// currentHasData reports whether current entry has started.
func (s *parseState) currentHasData() bool {
	if s.current == nil {
		return false
	}

	return s.current.ID != "" ||
		s.current.IDPlural != "" ||
		len(s.current.Translations) > 0
}

// appendContinuation appends continuation text to active field.
func (s *parseState) appendContinuation(value string) {
	if s.current == nil {
		return
	}

	switch s.section {
	case "msgctxt":
		s.current.Context += value
	case "msgid":
		s.current.ID += value
	case "msgid_plural":
		s.current.IDPlural += value
	case "msgstr":
		current := s.current.Translations[s.sectionIndex]
		s.current.Translations[s.sectionIndex] = current + value
	}
}

// flushCurrent closes current entry and appends it to document.
func (s *parseState) flushCurrent() {
	if s.current == nil {
		return
	}

	entry := s.current
	s.current = nil
	s.section = ""
	s.sectionIndex = 0

	if !s.headerSeen && entry.ID == "" && entry.IDPlural == "" {
		s.document.HeaderComments = append([]Comment(nil), entry.Comments...)
		parsedHeaders := parseHeadersFromEntry(entry)
		if len(parsedHeaders) > 0 {
			s.document.Headers = parsedHeaders
		}
		s.headerSeen = true
		return
	}
	if entry.ID == "" &&
		entry.Context == "" &&
		entry.IDPlural == "" &&
		len(entry.Translations) == 0 {
		return
	}

	s.document.Entries = append(s.document.Entries, entry)
}

// addMissingMsgIDDiagnostics emits diagnostics for malformed entries.
func (s *parseState) addMissingMsgIDDiagnostics() {
	for _, entry := range s.document.Entries {
		if entry == nil {
			continue
		}
		if entry.ID != "" {
			continue
		}

		s.addDiagnostic(
			CodeParseMissingMsgID,
			"entry has no msgid",
			entry.Position,
		)
	}
}

// addDiagnostic appends one diagnostic.
func (s *parseState) addDiagnostic(
	code lint.Code,
	message string,
	position Position,
) {
	s.diagnostics = append(
		s.diagnostics,
		newLintDiagnostic(lint.SeverityError, code, message, position),
	)
}

// parseCommentLine parses typed comment prefix and body.
func parseCommentLine(trimmed string, position Position) Comment {
	comment := Comment{
		Kind:     CommentOther,
		Text:     strings.TrimSpace(strings.TrimPrefix(trimmed, "#")),
		Raw:      trimmed,
		Position: position,
	}

	switch {
	case strings.HasPrefix(trimmed, "#."):
		comment.Kind = CommentExtracted
		comment.Text = strings.TrimSpace(strings.TrimPrefix(trimmed, "#."))
	case strings.HasPrefix(trimmed, "#:"):
		comment.Kind = CommentReference
		comment.Text = strings.TrimSpace(strings.TrimPrefix(trimmed, "#:"))
	case strings.HasPrefix(trimmed, "#,"):
		comment.Kind = CommentFlags
		comment.Text = strings.TrimSpace(strings.TrimPrefix(trimmed, "#,"))
	case strings.HasPrefix(trimmed, "#|"):
		comment.Kind = CommentPrevious
		comment.Text = strings.TrimSpace(strings.TrimPrefix(trimmed, "#|"))
	default:
		comment.Kind = CommentTranslator
		comment.Text = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
	}

	return comment
}

// collectEntryCommentMetadata extracts flags and references from comments.
func collectEntryCommentMetadata(entry *Entry) {
	flags := make([]string, 0)
	references := make([]string, 0)

	for _, comment := range entry.Comments {
		switch comment.Kind {
		case CommentFlags:
			for part := range strings.SplitSeq(comment.Text, ",") {
				flag := strings.TrimSpace(part)
				if flag == "" {
					continue
				}
				flags = append(flags, flag)
			}
		case CommentReference:
			for part := range strings.SplitSeq(comment.Text, " ") {
				ref := strings.TrimSpace(part)
				if ref == "" {
					continue
				}
				references = append(references, ref)
			}
		}
	}

	entry.Flags = flags
	entry.References = references
}

// parseHeadersFromEntry parses header lines from msgstr[0].
func parseHeadersFromEntry(entry *Entry) []Header {
	value := entry.Translations[0]
	if value == "" {
		return nil
	}

	estimate := strings.Count(value, "\n") + 1
	headers := make([]Header, 0, estimate)
	start := 0
	for start <= len(value) {
		end := strings.IndexByte(value[start:], '\n')
		var line string
		if end < 0 {
			line = value[start:]
			start = len(value) + 1
		} else {
			line = value[start : start+end]
			start += end + 1
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		index := strings.IndexByte(trimmed, ':')
		if index <= 0 {
			continue
		}

		headers = append(headers, Header{
			Key:      strings.TrimSpace(trimmed[:index]),
			Value:    strings.TrimSpace(trimmed[index+1:]),
			Position: entry.Position,
		})
	}

	if len(headers) == 0 {
		return nil
	}

	return headers
}

// parseIndexedMsgStr parses "msgstr[n] "value"".
func parseIndexedMsgStr(line string) (int, string, bool) {
	open := strings.Index(line, "[")
	closeIndex := strings.Index(line, "]")
	if open == -1 || closeIndex == -1 || closeIndex <= open+1 {
		return 0, "", false
	}

	indexValue := line[open+1 : closeIndex]
	index, err := strconv.Atoi(indexValue)
	if err != nil || index < 0 {
		return 0, "", false
	}

	value, ok := extractQuotedValueChecked(line)
	if !ok {
		return 0, "", false
	}

	return index, value, true
}

// extractQuotedValueChecked extracts quoted value and reports malformed lines.
func extractQuotedValueChecked(line string) (string, bool) {
	start := strings.IndexByte(line, '"')
	if start == -1 {
		return "", false
	}

	closing := -1
	hasEscape := false
	for i := start + 1; i < len(line); i++ {
		ch := line[i]
		if ch == '\\' {
			hasEscape = true
			i++
			continue
		}
		if ch == '"' {
			closing = i
			break
		}
	}

	if closing == -1 {
		return "", false
	}
	if !hasEscape {
		return line[start+1 : closing], true
	}

	var (
		out     strings.Builder
		escaped bool
	)
	out.Grow(closing - start - 1)
	for i := start + 1; i < closing; i++ {
		ch := line[i]
		if escaped {
			switch ch {
			case 'n':
				out.WriteByte('\n')
			case 't':
				out.WriteByte('\t')
			case 'r':
				out.WriteByte('\r')
			case '\\':
				out.WriteByte('\\')
			case '"':
				out.WriteByte('"')
			default:
				out.WriteByte('\\')
				out.WriteByte(ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}

		out.WriteByte(ch)
	}

	if escaped {
		out.WriteByte('\\')
	}

	return out.String(), true
}

// hasErrorDiagnostics reports whether diagnostics list contains errors.
func hasErrorDiagnostics(diagnostics []lint.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == lint.SeverityError {
			return true
		}
	}

	return false
}

// formatParseError formats top-level parse failure.
func formatParseError(diagnostics []lint.Diagnostic) error {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity != lint.SeverityError {
			continue
		}

		return fmt.Errorf(
			"%s at line %d:%d: %s",
			diagnostic.Code,
			diagnostic.Start.Line,
			diagnostic.Start.Column,
			diagnostic.Message,
		)
	}

	return errors.New("parse failed")
}

// lineColumn returns first non-space column (1-based).
func lineColumn(line string) int {
	for i := 0; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			return i + 1
		}
	}

	return 1
}
