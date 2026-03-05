// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"os"
	"path/filepath"
	"testing"
)

// readFixture reads one testdata fixture file.
func readFixture(t *testing.T, relativePath string) []byte {
	t.Helper()

	path := filepath.Join("testdata", filepath.FromSlash(relativePath))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %q: %v", relativePath, err)
	}

	return data
}
