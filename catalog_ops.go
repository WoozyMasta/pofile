// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

// Validate checks catalog structural consistency.
func (c *Catalog) Validate() error {
	if c == nil {
		return ErrNilCatalog
	}

	seen := make(map[string]struct{}, len(c.Messages))
	for index, message := range c.Messages {
		if message == nil {
			return fmt.Errorf("messages[%d]: %w", index, ErrNilMessage)
		}
		if message.ID == "" {
			return fmt.Errorf("messages[%d]: %w", index, ErrMessageIDRequired)
		}

		key := message.Domain + "\x00" + message.Context + "\x00" + message.ID
		if _, ok := seen[key]; ok {
			return fmt.Errorf(
				"messages[%d] (%q, %q, %q): %w",
				index,
				message.Domain,
				message.Context,
				message.ID,
				ErrDuplicateMessage,
			)
		}
		seen[key] = struct{}{}
	}

	return nil
}

// MergeTemplate applies template messages and keeps existing translations.
func MergeTemplate(template, existing *Catalog) (*Catalog, error) {
	if template == nil {
		return nil, ErrTemplateRequired
	}
	if err := template.Validate(); err != nil {
		return nil, err
	}
	if existing != nil {
		if err := existing.Validate(); err != nil {
			return nil, err
		}
	}

	out := template.Clone()
	if out.Headers == nil {
		out.Headers = make(map[string]string)
	}

	if existing != nil {
		applyExistingHeaders(out, existing)
		applyExistingMessages(out, existing)
	}

	if out.Language != "" {
		out.SetHeader("Language", out.Language)
	} else if lang := out.Header("Language"); lang != "" {
		out.Language = lang
	}

	return out, nil
}

// ContentHash returns deterministic hash for effective catalog content.
func (c *Catalog) ContentHash() string {
	if c == nil {
		return ""
	}

	hash := sha256.New()
	writeHashContent(hash, c)

	return hex.EncodeToString(hash.Sum(nil))
}

// applyExistingHeaders merges existing headers into output.
func applyExistingHeaders(out, existing *Catalog) {
	maps.Copy(out.Headers, existing.Headers)
	if existing.Language != "" {
		out.Language = existing.Language
	}
}

// applyExistingMessages copies translations and comments from existing catalog.
func applyExistingMessages(out, existing *Catalog) {
	for _, message := range out.Messages {
		if message == nil {
			continue
		}
		current := existing.FindMessageInDomain(
			message.Domain,
			message.Context,
			message.ID,
		)
		if current == nil && message.Domain != "" {
			// Fallback for catalogs produced without domain support.
			current = existing.FindMessageInDomain(
				"",
				message.Context,
				message.ID,
			)
		}
		if current == nil {
			continue
		}

		message.Translations = maps.Clone(current.Translations)
		if message.Translations == nil {
			message.Translations = make(map[int]string)
		}
		if len(current.Comments) > 0 {
			message.Comments = slices.Clone(current.Comments)
		}
		if len(current.Flags) > 0 {
			message.Flags = slices.Clone(current.Flags)
		}
		if len(current.References) > 0 {
			message.References = slices.Clone(current.References)
		}
		message.IDPlural = current.IDPlural
		message.Obsolete = current.Obsolete
		message.PreviousContext = current.PreviousContext
		message.PreviousID = current.PreviousID
		message.PreviousIDPlural = current.PreviousIDPlural
	}
}

// writeHashContent writes hash input excluding dynamic date headers.
func writeHashContent(writer io.Writer, catalog *Catalog) {
	excluded := map[string]struct{}{
		"PO-Revision-Date":  {},
		"POT-Creation-Date": {},
		"X-Content-Hash":    {},
	}

	_, _ = io.WriteString(writer, catalog.Language)
	_, _ = io.WriteString(writer, "\n")

	keys := make([]string, 0, len(catalog.Headers))
	for key := range catalog.Headers {
		if _, skip := excluded[key]; skip {
			continue
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		_, _ = io.WriteString(writer, key)
		_, _ = io.WriteString(writer, ":")
		_, _ = io.WriteString(writer, catalog.Headers[key])
		_, _ = io.WriteString(writer, "\n")
	}

	for _, message := range catalog.Messages {
		if message == nil {
			continue
		}
		_, _ = io.WriteString(writer, message.Domain)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.Context)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.ID)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.IDPlural)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, boolToken(message.Obsolete))
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.PreviousContext)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.PreviousID)
		_, _ = io.WriteString(writer, "\n")
		_, _ = io.WriteString(writer, message.PreviousIDPlural)
		_, _ = io.WriteString(writer, "\n")

		indexes := make([]int, 0, len(message.Translations))
		for index := range message.Translations {
			indexes = append(indexes, index)
		}
		slices.Sort(indexes)
		for _, index := range indexes {
			_, _ = io.WriteString(writer, strconv.Itoa(index))
			_, _ = io.WriteString(writer, "=")
			_, _ = io.WriteString(writer, message.Translations[index])
			_, _ = io.WriteString(writer, "\n")
		}

		for _, flag := range message.Flags {
			_, _ = io.WriteString(writer, "flag:")
			_, _ = io.WriteString(writer, flag)
			_, _ = io.WriteString(writer, "\n")
		}
		for _, reference := range message.References {
			_, _ = io.WriteString(writer, "ref:")
			_, _ = io.WriteString(writer, reference)
			_, _ = io.WriteString(writer, "\n")
		}
		for _, comment := range message.Comments {
			_, _ = io.WriteString(writer, comment)
			_, _ = io.WriteString(writer, "\n")
		}
		_, _ = io.WriteString(writer, "\n")
	}
}

// boolToken formats bool for hash stream.
func boolToken(value bool) string {
	if value {
		return "1"
	}

	return "0"
}
