// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/woozymasta/lintkit/lint"
)

// Parse decodes PO/POT bytes into semantic catalog.
func Parse(data []byte) (*Catalog, error) {
	document, err := ParseDocument(data)
	if err != nil {
		return nil, err
	}

	return document.ToCatalog()
}

// ParseReader decodes PO/POT stream into semantic catalog.
func ParseReader(reader io.Reader) (*Catalog, error) {
	document, err := ParseDocumentReader(reader)
	if err != nil {
		return nil, err
	}

	return document.ToCatalog()
}

// ParseCatalogWithDiagnostics decodes PO/POT bytes and returns diagnostics.
func ParseCatalogWithDiagnostics(
	data []byte,
	options ParseOptions,
) (*Catalog, []lint.Diagnostic, error) {
	document, diagnostics, err := ParseDocumentWithOptions(data, options)
	if err != nil && document == nil {
		return nil, diagnostics, err
	}

	catalog, convertErr := document.ToCatalog()
	if convertErr != nil {
		return nil, diagnostics, convertErr
	}

	return catalog, diagnostics, err
}

// ParseFile opens and parses a PO/POT file into semantic catalog.
func ParseFile(path string) (*Catalog, error) {
	document, err := ParseDocumentFile(path)
	if err != nil {
		return nil, err
	}

	return document.ToCatalog()
}

// ParseDir loads all *.po files from directory indexed by file base name.
func ParseDir(dir string) (map[string]*Catalog, error) {
	files, err := listPOFiles(dir)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*Catalog, len(files))
	for _, name := range files {
		path := filepath.Join(dir, name)
		catalog, err := ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", path, err)
		}

		key := strings.TrimSuffix(name, filepath.Ext(name))
		out[key] = catalog
	}

	return out, nil
}

// ParseDocumentFile opens and parses one PO/POT file into document.
func ParseDocumentFile(path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open po file: %w", err)
	}
	defer func() { _ = file.Close() }()

	document, err := ParseDocumentReader(file)
	if err != nil {
		return nil, fmt.Errorf("parse po file: %w", err)
	}

	return document, nil
}

// ParseDocumentDir loads all *.po files from directory indexed by file base name.
func ParseDocumentDir(dir string) (map[string]*Document, error) {
	files, err := listPOFiles(dir)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*Document, len(files))
	for _, name := range files {
		path := filepath.Join(dir, name)
		document, err := ParseDocumentFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", path, err)
		}

		key := strings.TrimSuffix(name, filepath.Ext(name))
		out[key] = document
	}

	return out, nil
}

// UnmarshalText decodes semantic catalog from PO/POT text bytes.
func (c *Catalog) UnmarshalText(data []byte) error {
	parsed, err := Parse(data)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrNilCatalog
	}

	*c = *parsed
	return nil
}

// listPOFiles returns sorted .po filenames from directory.
func listPOFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read po directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".po") {
			continue
		}

		files = append(files, entry.Name())
	}
	sort.Strings(files)

	return files, nil
}
