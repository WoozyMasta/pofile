// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import "fmt"

// EntryKey identifies one PO entry by domain, context, and id.
type EntryKey struct {
	Domain  string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
	ID      string `json:"id" yaml:"id"`
}

// Index provides fast key-based lookup for a Document.
type Index struct {
	entries map[EntryKey]*Entry
}

// NewIndex builds a lookup index for document entries.
func NewIndex(document *Document) (*Index, error) {
	if document == nil {
		return nil, ErrNilDocument
	}

	entries := make(map[EntryKey]*Entry, len(document.Entries))
	for index, entry := range document.Entries {
		if entry == nil || entry.ID == "" {
			continue
		}

		key := EntryKey{
			Domain:  entry.Domain,
			Context: entry.Context,
			ID:      entry.ID,
		}
		if _, ok := entries[key]; ok {
			return nil, fmt.Errorf(
				"entries[%d] (%q, %q, %q): %w",
				index,
				entry.Domain,
				entry.Context,
				entry.ID,
				ErrDuplicateEntryKey,
			)
		}

		entries[key] = entry
	}

	return &Index{entries: entries}, nil
}

// Entry returns entry by full key, or nil when missing.
func (i *Index) Entry(key EntryKey) *Entry {
	if i == nil {
		return nil
	}

	return i.entries[key]
}

// EntryInDomain returns entry by domain, context, and id.
func (i *Index) EntryInDomain(domain, context, id string) *Entry {
	return i.Entry(EntryKey{
		Domain:  domain,
		Context: context,
		ID:      id,
	})
}

// EntryDefaultDomain returns entry in default domain.
func (i *Index) EntryDefaultDomain(context, id string) *Entry {
	return i.EntryInDomain("", context, id)
}
