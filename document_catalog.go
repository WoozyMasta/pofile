// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"maps"
	"slices"
	"strings"
)

// ToCatalog converts lossless document to semantic catalog.
func (d *Document) ToCatalog() (*Catalog, error) {
	if d == nil {
		return nil, ErrNilDocument
	}

	catalog := NewCatalog()
	catalog.Messages = make([]*Message, 0, len(d.Entries))
	for _, header := range d.Headers {
		catalog.SetHeader(header.Key, header.Value)
	}
	catalog.Language = catalog.Header("Language")

	for _, entry := range d.Entries {
		if entry == nil || entry.ID == "" {
			continue
		}
		message := catalog.UpsertMessageInDomain(
			entry.Domain,
			entry.Context,
			entry.ID,
			entry.Translations[0],
		)
		message.IDPlural = entry.IDPlural
		message.Obsolete = entry.Obsolete
		message.PreviousContext = entry.PreviousContext
		message.PreviousID = entry.PreviousID
		message.PreviousIDPlural = entry.PreviousIDPlural
		if len(entry.Translations) > 0 {
			for index, value := range entry.Translations {
				message.SetTranslationAt(index, value)
			}
		}
		message.Flags = slices.Clone(entry.Flags)
		message.References = slices.Clone(entry.References)
		message.Comments = make([]string, 0, len(entry.Comments))
		for _, comment := range entry.Comments {
			raw := comment.Raw
			if raw == "" {
				raw = renderComment(comment)
			}
			message.Comments = append(message.Comments, raw)
		}
	}

	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	return catalog, nil
}

// DocumentFromCatalog converts semantic catalog to document.
func DocumentFromCatalog(catalog *Catalog) *Document {
	if catalog == nil {
		return NewDocument()
	}

	document := NewDocument()
	for _, key := range sortedHeaderKeys(catalog.Headers) {
		document.Headers = append(document.Headers, Header{
			Key:   key,
			Value: catalog.Headers[key],
		})
	}
	for _, message := range catalog.Messages {
		if message == nil {
			continue
		}

		entry := &Entry{
			Domain:           message.Domain,
			Context:          message.Context,
			ID:               message.ID,
			IDPlural:         message.IDPlural,
			Obsolete:         message.Obsolete,
			PreviousContext:  message.PreviousContext,
			PreviousID:       message.PreviousID,
			PreviousIDPlural: message.PreviousIDPlural,
			Translations:     make(map[int]string),
		}
		maps.Copy(entry.Translations, message.Translations)
		if len(entry.Translations) == 0 {
			entry.Translations[0] = ""
		}
		if len(message.Comments) > 0 {
			entry.Comments = parseRawComments(message.Comments)
			entry.Comments = ensureEntryMetadataComments(entry.Comments, message)
		} else {
			entry.Comments = synthesizeCommentsFromMetadata(message)
		}
		collectEntryCommentMetadata(entry)

		document.Entries = append(document.Entries, entry)
	}

	return document
}

// sortedHeaderKeys returns stable sorted header keys.
func sortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	return keys
}

// parseRawComments converts raw comment lines into typed comments.
func parseRawComments(rawComments []string) []Comment {
	comments := make([]Comment, 0, len(rawComments))
	for _, raw := range rawComments {
		comment := parseCommentLine(raw, Position{})
		comment.Raw = raw
		comments = append(comments, comment)
	}

	return comments
}

// ensureEntryMetadataComments ensures semantic metadata is represented in comments.
func ensureEntryMetadataComments(comments []Comment, message *Message) []Comment {
	if message == nil {
		return comments
	}

	out := slices.Clone(comments)
	flagSet := make(map[string]struct{})
	referenceSet := make(map[string]struct{})
	hasPreviousContext := false
	hasPreviousID := false
	hasPreviousIDPlural := false

	for _, comment := range out {
		switch comment.Kind {
		case CommentFlags:
			for part := range strings.SplitSeq(comment.Text, ",") {
				flag := strings.ToLower(strings.TrimSpace(part))
				if flag == "" {
					continue
				}

				flagSet[flag] = struct{}{}
			}
		case CommentReference:
			for part := range strings.SplitSeq(comment.Text, " ") {
				reference := strings.TrimSpace(part)
				if reference == "" {
					continue
				}

				referenceSet[reference] = struct{}{}
			}
		case CommentPrevious:
			switch {
			case strings.HasPrefix(comment.Text, "msgctxt "):
				hasPreviousContext = true
			case strings.HasPrefix(comment.Text, "msgid_plural "):
				hasPreviousIDPlural = true
			case strings.HasPrefix(comment.Text, "msgid "):
				hasPreviousID = true
			}
		}
	}

	missingFlags := make([]string, 0)
	for _, flag := range message.Flags {
		trimmed := strings.TrimSpace(flag)
		normalized := strings.ToLower(trimmed)
		if normalized == "" {
			continue
		}
		if _, ok := flagSet[normalized]; ok {
			continue
		}

		missingFlags = append(missingFlags, trimmed)
		flagSet[normalized] = struct{}{}
	}
	if len(missingFlags) > 0 {
		out = append(out, Comment{
			Kind: CommentFlags,
			Text: strings.Join(missingFlags, ", "),
			Raw:  "#, " + strings.Join(missingFlags, ", "),
		})
	}

	missingReferences := make([]string, 0)
	for _, reference := range message.References {
		trimmed := strings.TrimSpace(reference)
		if trimmed == "" {
			continue
		}
		if _, ok := referenceSet[trimmed]; ok {
			continue
		}

		missingReferences = append(missingReferences, trimmed)
		referenceSet[trimmed] = struct{}{}
	}
	if len(missingReferences) > 0 {
		out = append(out, Comment{
			Kind: CommentReference,
			Text: strings.Join(missingReferences, " "),
			Raw:  "#: " + strings.Join(missingReferences, " "),
		})
	}

	if message.PreviousContext != "" && !hasPreviousContext {
		out = append(out, makePreviousComment("msgctxt", message.PreviousContext))
	}
	if message.PreviousID != "" && !hasPreviousID {
		out = append(out, makePreviousComment("msgid", message.PreviousID))
	}
	if message.PreviousIDPlural != "" && !hasPreviousIDPlural {
		out = append(
			out,
			makePreviousComment("msgid_plural", message.PreviousIDPlural),
		)
	}

	return out
}

// makePreviousComment builds one #| previous-value comment.
func makePreviousComment(key, value string) Comment {
	escaped := escapePOString(value)
	text := key + ` "` + escaped + `"`

	return Comment{
		Kind: CommentPrevious,
		Text: text,
		Raw:  "#| " + text,
	}
}

// synthesizeCommentsFromMetadata builds comment lines from semantic fields.
func synthesizeCommentsFromMetadata(message *Message) []Comment {
	if message == nil {
		return nil
	}

	comments := make([]Comment, 0, len(message.Flags)+len(message.References)+3)
	if message.PreviousContext != "" {
		comments = append(
			comments,
			makePreviousComment("msgctxt", message.PreviousContext),
		)
	}
	if message.PreviousID != "" {
		comments = append(comments, makePreviousComment("msgid", message.PreviousID))
	}
	if message.PreviousIDPlural != "" {
		comments = append(
			comments,
			makePreviousComment("msgid_plural", message.PreviousIDPlural),
		)
	}
	if len(message.References) > 0 {
		comments = append(comments, Comment{
			Kind: CommentReference,
			Text: strings.Join(message.References, " "),
			Raw:  "#: " + strings.Join(message.References, " "),
		})
	}
	if len(message.Flags) > 0 {
		comments = append(comments, Comment{
			Kind: CommentFlags,
			Text: strings.Join(message.Flags, ", "),
			Raw:  "#, " + strings.Join(message.Flags, ", "),
		})
	}

	return comments
}
