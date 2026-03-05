// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"fmt"
	"os"
	"strings"
)

var headerOrder = []string{
	"Project-Id-Version",
	"POT-Creation-Date",
	"PO-Revision-Date",
	"Last-Translator",
	"Language-Team",
	"Language",
	"MIME-Version",
	"Content-Type",
	"Content-Transfer-Encoding",
	"X-Generator",
}

// Format encodes semantic catalog into PO/POT text bytes.
func Format(catalog *Catalog) ([]byte, error) {
	return FormatWithOptions(catalog, nil)
}

// FormatWithOptions encodes semantic catalog using document write options.
func FormatWithOptions(catalog *Catalog, options *WriteOptions) ([]byte, error) {
	if catalog == nil {
		return nil, ErrNilCatalog
	}
	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	document := DocumentFromCatalog(catalog)
	return FormatDocument(document, options)
}

// WriteFile encodes semantic catalog and writes to file path.
func WriteFile(path string, catalog *Catalog) error {
	return WriteFileWithOptions(path, catalog, nil)
}

// WriteFileWithOptions encodes semantic catalog and writes with options.
func WriteFileWithOptions(path string, catalog *Catalog, options *WriteOptions) error {
	data, err := FormatWithOptions(catalog, options)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write po file: %w", err)
	}

	return nil
}

// MarshalText encodes catalog into PO/POT bytes.
func (c *Catalog) MarshalText() ([]byte, error) {
	return Format(c)
}

// writeHeaderLine writes one escaped header line.
func writeHeaderLine(builder *strings.Builder, key, value string) {
	line := key + ": " + value + "\n"
	builder.WriteString(`"`)
	builder.WriteString(escapePOString(line))
	builder.WriteString(`"` + "\n")
}

// writeQuotedString writes one escaped string with multiline support.
func writeQuotedString(builder *strings.Builder, value string) {
	if value == "" {
		builder.WriteString(`""`)
		return
	}

	if !strings.Contains(value, "\n") {
		builder.WriteString(`"`)
		builder.WriteString(escapePOString(value))
		builder.WriteString(`"`)
		return
	}

	lines := strings.Split(value, "\n")
	for index, line := range lines {
		if index > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(`"`)
		builder.WriteString(escapePOString(line))
		if index < len(lines)-1 {
			builder.WriteString(`\n`)
		}
		builder.WriteString(`"`)
	}
}

// escapePOString escapes PO special characters.
func escapePOString(value string) string {
	var out strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '"':
			out.WriteString(`\"`)
		case '\n':
			out.WriteString(`\n`)
		case '\t':
			out.WriteString(`\t`)
		case '\r':
			out.WriteString(`\r`)
		default:
			out.WriteRune(r)
		}
	}

	return out.String()
}
