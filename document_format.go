// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
)

// FormatDocument encodes document into PO/POT text.
func FormatDocument(document *Document, options *WriteOptions) ([]byte, error) {
	if document == nil {
		return nil, ErrNilDocument
	}

	settings := normalizeWriteOptions(options)
	var builder strings.Builder

	writeDocumentHeader(&builder, document, settings)
	writeDocumentEntries(&builder, document.Entries, settings)

	return []byte(builder.String()), nil
}

// WriteDocumentFile formats document and writes file to disk.
func WriteDocumentFile(path string, document *Document, options *WriteOptions) error {
	data, err := FormatDocument(document, options)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write po file: %w", err)
	}

	return nil
}

// normalizeWriteOptions fills defaults for write behavior.
func normalizeWriteOptions(options *WriteOptions) WriteOptions {
	if options == nil {
		return WriteOptions{
			Mode:        WriteModePreserve,
			HeaderOrder: append([]string(nil), headerOrder...),
		}
	}

	out := *options
	if out.Mode == "" {
		out.Mode = WriteModePreserve
	}
	if len(out.HeaderOrder) == 0 {
		out.HeaderOrder = append([]string(nil), headerOrder...)
	}

	return out
}

// writeDocumentHeader writes header entry and comments.
func writeDocumentHeader(builder *strings.Builder, document *Document, options WriteOptions) {
	for _, comment := range document.HeaderComments {
		writeCommentLine(builder, comment, false)
	}
	if len(document.HeaderComments) > 0 {
		builder.WriteString("\n")
	}

	builder.WriteString("msgid \"\"\n")
	builder.WriteString("msgstr \"\"\n")

	headers := cloneHeaders(document.Headers)
	if options.Mode == WriteModeCanonical && options.SortHeaders {
		sortHeaders(headers, options.HeaderOrder)
	}
	for _, header := range headers {
		writeHeaderLine(builder, header.Key, header.Value)
	}

	builder.WriteString("\n")
}

// writeDocumentEntries writes entries in selected order.
func writeDocumentEntries(builder *strings.Builder, entries []*Entry, options WriteOptions) {
	ordered := cloneEntries(entries)
	if options.SortEntries || options.Mode == WriteModeCanonical {
		sort.SliceStable(ordered, func(i, j int) bool {
			left := ordered[i]
			right := ordered[j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.Domain != right.Domain {
				return left.Domain < right.Domain
			}
			if left.Context != right.Context {
				return left.Context < right.Context
			}
			return left.ID < right.ID
		})
	}

	lastDomain := ""
	for _, entry := range ordered {
		if entry == nil {
			continue
		}

		if entry.Domain != "" && entry.Domain != lastDomain {
			writeDomainLine(builder, entry.Domain, entry.Obsolete)
			builder.WriteString("\n")
			lastDomain = entry.Domain
		}

		for _, comment := range entry.Comments {
			writeCommentLine(builder, comment, entry.Obsolete)
		}

		writeEntryField(builder, "msgctxt", entry.Context, entry.Obsolete)
		writeEntryField(builder, "msgid", entry.ID, entry.Obsolete)
		writeEntryField(builder, "msgid_plural", entry.IDPlural, entry.Obsolete)

		if hasPluralTranslations(entry) {
			indexes := make([]int, 0, len(entry.Translations))
			for index := range entry.Translations {
				indexes = append(indexes, index)
			}
			slices.Sort(indexes)
			for _, index := range indexes {
				key := fmt.Sprintf("msgstr[%d]", index)
				writeEntryField(builder, key, entry.Translations[index], entry.Obsolete)
			}
		} else {
			writeEntryField(builder, "msgstr", entry.Translations[0], entry.Obsolete)
		}

		builder.WriteString("\n")
	}
}

// hasPluralTranslations reports whether entry should use msgstr[n] output.
func hasPluralTranslations(entry *Entry) bool {
	if entry == nil {
		return false
	}
	if entry.IDPlural != "" {
		return true
	}

	for index := range entry.Translations {
		if index != 0 {
			return true
		}
	}

	return false
}

// cloneHeaders clones headers slice.
func cloneHeaders(headers []Header) []Header {
	if headers == nil {
		return nil
	}

	out := make([]Header, len(headers))
	copy(out, headers)
	return out
}

// cloneEntries shallow-clones entries slice.
func cloneEntries(entries []*Entry) []*Entry {
	if entries == nil {
		return nil
	}

	out := make([]*Entry, len(entries))
	copy(out, entries)
	return out
}

// sortHeaders sorts headers by preferred order then key name.
func sortHeaders(headers []Header, preferred []string) {
	indexes := make(map[string]int, len(preferred))
	for index, key := range preferred {
		indexes[strings.ToLower(strings.TrimSpace(key))] = index
	}
	sort.SliceStable(headers, func(i, j int) bool {
		left := strings.ToLower(headers[i].Key)
		right := strings.ToLower(headers[j].Key)
		leftRank, leftKnown := indexes[left]
		rightRank, rightKnown := indexes[right]
		switch {
		case leftKnown && rightKnown:
			if leftRank != rightRank {
				return leftRank < rightRank
			}
		case leftKnown:
			return true
		case rightKnown:
			return false
		}
		return headers[i].Key < headers[j].Key
	})
}

// writeCommentLine writes one comment line.
func writeCommentLine(builder *strings.Builder, comment Comment, obsolete bool) {
	line := strings.TrimSpace(comment.Raw)
	if line == "" {
		line = renderComment(comment)
	}

	if obsolete && !strings.HasPrefix(line, "#~") {
		builder.WriteString("#~ ")
		builder.WriteString(line)
		builder.WriteString("\n")
		return
	}

	builder.WriteString(line)
	builder.WriteString("\n")
}

// renderComment reconstructs comment line from typed data.
func renderComment(comment Comment) string {
	text := comment.Text
	switch comment.Kind {
	case CommentExtracted:
		return "#. " + text
	case CommentReference:
		return "#: " + text
	case CommentFlags:
		return "#, " + text
	case CommentPrevious:
		return "#| " + text
	default:
		return "# " + text
	}
}

// writeDomainLine writes domain directive.
func writeDomainLine(builder *strings.Builder, domain string, obsolete bool) {
	line := "domain "
	if obsolete {
		line = "#~ " + line
	}
	builder.WriteString(line)
	writeQuotedString(builder, domain)
	builder.WriteString("\n")
}

// writeEntryField writes one msg* field when value is relevant.
func writeEntryField(builder *strings.Builder, key string, value string, obsolete bool) {
	if key == "msgctxt" && value == "" {
		return
	}
	if key == "msgid_plural" && value == "" {
		return
	}

	prefix := ""
	if obsolete {
		prefix = "#~ "
	}
	builder.WriteString(prefix)
	builder.WriteString(key)
	builder.WriteString(" ")
	writeQuotedString(builder, value)
	builder.WriteString("\n")
}
