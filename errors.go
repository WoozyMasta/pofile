// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import "errors"

var (
	// ErrNilCatalog indicates that a catalog argument is nil.
	ErrNilCatalog = errors.New("catalog is nil")

	// ErrNilDocument indicates that a document argument is nil.
	ErrNilDocument = errors.New("document is nil")

	// ErrTemplateRequired indicates that template catalog is missing.
	ErrTemplateRequired = errors.New("template catalog is required")

	// ErrNilMessage indicates that message list contains a nil item.
	ErrNilMessage = errors.New("message is nil")

	// ErrMessageIDRequired indicates message with empty msgid.
	ErrMessageIDRequired = errors.New("message id is required")

	// ErrDuplicateMessage indicates duplicate context+id pair.
	ErrDuplicateMessage = errors.New("duplicate message")

	// ErrDuplicateEntryKey indicates duplicate domain+context+id in document index.
	ErrDuplicateEntryKey = errors.New("duplicate entry key")
)
