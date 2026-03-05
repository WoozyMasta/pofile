// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"maps"
	"slices"
	"strings"
)

// Catalog is an in-memory PO/POT semantic model.
type Catalog struct {
	// Headers stores header values from the initial msgid/msgstr block.
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Language is copied from "Language" header when present.
	Language string `json:"language,omitempty" yaml:"language,omitempty"`

	// Messages stores translation units.
	Messages []*Message `json:"messages,omitempty" yaml:"messages,omitempty"`
}

// Message is one translation entry.
type Message struct {
	// Translations stores msgstr values by plural index.
	// Index 0 is the singular translation.
	Translations map[int]string `json:"translations,omitempty" yaml:"translations,omitempty"`

	// Domain stores optional gettext domain.
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty"`

	// Context corresponds to msgctxt.
	Context string `json:"context,omitempty" yaml:"context,omitempty"`

	// ID corresponds to msgid.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// IDPlural corresponds to msgid_plural.
	IDPlural string `json:"id_plural,omitempty" yaml:"id_plural,omitempty"`

	// Previous values come from "#|" comments.
	PreviousContext string `json:"previous_context,omitempty" yaml:"previous_context,omitempty"`

	// PreviousID is previous msgid from "#| msgid".
	PreviousID string `json:"previous_id,omitempty" yaml:"previous_id,omitempty"`

	// PreviousIDPlural is previous msgid_plural from "#| msgid_plural".
	PreviousIDPlural string `json:"previous_id_plural,omitempty" yaml:"previous_id_plural,omitempty"`

	// Comments stores entry comments as raw lines.
	Comments []string `json:"comments,omitempty" yaml:"comments,omitempty"`

	// Flags stores parsed values from "#," comments.
	Flags []string `json:"flags,omitempty" yaml:"flags,omitempty"`

	// References stores parsed values from "#:" comments.
	References []string `json:"references,omitempty" yaml:"references,omitempty"`

	// Obsolete marks an obsolete "#~" entry.
	Obsolete bool `json:"obsolete,omitempty" yaml:"obsolete,omitempty"`
}

// NewCatalog creates an empty catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		Headers:  make(map[string]string),
		Messages: make([]*Message, 0),
	}
}

// Clone makes a deep copy of catalog.
func (c *Catalog) Clone() *Catalog {
	if c == nil {
		return nil
	}

	clone := NewCatalog()
	clone.Language = c.Language
	maps.Copy(clone.Headers, c.Headers)
	for _, message := range c.Messages {
		clone.Messages = append(clone.Messages, cloneMessage(message))
	}

	return clone
}

// SetHeader sets or replaces one header key.
func (c *Catalog) SetHeader(key, value string) {
	if c == nil {
		return
	}
	if c.Headers == nil {
		c.Headers = make(map[string]string)
	}

	c.Headers[key] = value
}

// Header returns one header value, or empty string when missing.
func (c *Catalog) Header(key string) string {
	if c == nil {
		return ""
	}
	if c.Headers == nil {
		return ""
	}

	return c.Headers[key]
}

// UpsertMessage upserts one singular message in default domain.
func (c *Catalog) UpsertMessage(context, id, translation string) *Message {
	return c.UpsertMessageInDomain("", context, id, translation)
}

// UpsertMessageInDomain upserts one singular message by domain+context+id.
func (c *Catalog) UpsertMessageInDomain(
	domain, context, id, translation string,
) *Message {
	if c == nil {
		return nil
	}

	found := c.FindMessageInDomain(domain, context, id)
	if found != nil {
		found.SetTranslationAt(0, translation)
		return found
	}

	message := &Message{
		Domain:       domain,
		Context:      context,
		ID:           id,
		Translations: map[int]string{0: translation},
	}
	c.Messages = append(c.Messages, message)

	return message
}

// FindMessage returns message by context+id in default domain.
func (c *Catalog) FindMessage(context, id string) *Message {
	return c.FindMessageInDomain("", context, id)
}

// DeleteMessage removes message by context+id in default domain.
func (c *Catalog) DeleteMessage(context, id string) bool {
	return c.DeleteMessageInDomain("", context, id)
}

// FindMessageInDomain returns message by domain+context+id, or nil when missing.
func (c *Catalog) FindMessageInDomain(domain, context, id string) *Message {
	if c == nil {
		return nil
	}

	for _, message := range c.Messages {
		if message == nil {
			continue
		}
		if message.Domain == domain &&
			message.Context == context &&
			message.ID == id {
			return message
		}
	}

	return nil
}

// DeleteMessageInDomain removes message by domain+context+id.
func (c *Catalog) DeleteMessageInDomain(domain, context, id string) bool {
	if c == nil {
		return false
	}

	for index, message := range c.Messages {
		if message == nil {
			continue
		}
		if message.Domain == domain &&
			message.Context == context &&
			message.ID == id {
			c.Messages = append(c.Messages[:index], c.Messages[index+1:]...)
			return true
		}
	}

	return false
}

// Translation returns singular translation by context+id in default domain.
func (c *Catalog) Translation(context, id string) string {
	return c.TranslationInDomain("", context, id)
}

// TranslationInDomain returns singular translation by domain+context+id.
func (c *Catalog) TranslationInDomain(domain, context, id string) string {
	message := c.FindMessageInDomain(domain, context, id)
	if message == nil {
		return ""
	}

	return message.TranslationAt(0)
}

// TranslationN returns plural translation by index in default domain.
func (c *Catalog) TranslationN(context, id string, index int) string {
	return c.TranslationNInDomain("", context, id, index)
}

// TranslationNInDomain returns plural translation by index.
func (c *Catalog) TranslationNInDomain(
	domain, context, id string,
	index int,
) string {
	message := c.FindMessageInDomain(domain, context, id)
	if message == nil {
		return ""
	}

	return message.TranslationAt(index)
}

// IsTranslated reports whether singular translation is non-empty in default domain.
func (c *Catalog) IsTranslated(context, id string) bool {
	return c.IsTranslatedInDomain("", context, id)
}

// IsTranslatedInDomain reports whether singular translation is non-empty.
func (c *Catalog) IsTranslatedInDomain(domain, context, id string) bool {
	return c.TranslationInDomain(domain, context, id) != ""
}

// HasFlag reports whether message has a given flag in "#," comment.
func (m *Message) HasFlag(flag string) bool {
	if m == nil {
		return false
	}

	expected := strings.ToLower(strings.TrimSpace(flag))
	if expected == "" {
		return false
	}

	if len(m.Flags) > 0 {
		for _, item := range m.Flags {
			if strings.EqualFold(strings.TrimSpace(item), expected) {
				return true
			}
		}
	}

	for _, raw := range m.Comments {
		trimmed := strings.TrimSpace(raw)
		if !strings.HasPrefix(trimmed, "#,") {
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "#,"))
		for part := range strings.SplitSeq(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), expected) {
				return true
			}
		}
	}

	return false
}

// TranslationAt returns translation value by plural index.
func (m *Message) TranslationAt(index int) string {
	if m == nil {
		return ""
	}
	if m.Translations == nil {
		return ""
	}

	return m.Translations[index]
}

// SetTranslationAt sets translation value by plural index.
func (m *Message) SetTranslationAt(index int, value string) {
	if m == nil {
		return
	}
	if m.Translations == nil {
		m.Translations = make(map[int]string)
	}

	m.Translations[index] = value
}

// IsPlural reports whether message has plural source or plural translations.
func (m *Message) IsPlural() bool {
	if m == nil {
		return false
	}
	if m.IDPlural != "" {
		return true
	}
	for index := range m.Translations {
		if index > 0 {
			return true
		}
	}

	return false
}

// cloneMessage deep-copies one message.
func cloneMessage(message *Message) *Message {
	if message == nil {
		return nil
	}

	clone := &Message{
		Domain:           message.Domain,
		Context:          message.Context,
		ID:               message.ID,
		IDPlural:         message.IDPlural,
		Obsolete:         message.Obsolete,
		PreviousContext:  message.PreviousContext,
		PreviousID:       message.PreviousID,
		PreviousIDPlural: message.PreviousIDPlural,
	}
	if len(message.Translations) > 0 {
		clone.Translations = maps.Clone(message.Translations)
	}
	if len(message.Comments) > 0 {
		clone.Comments = slices.Clone(message.Comments)
	}
	if len(message.Flags) > 0 {
		clone.Flags = slices.Clone(message.Flags)
	}
	if len(message.References) > 0 {
		clone.References = slices.Clone(message.References)
	}

	return clone
}
