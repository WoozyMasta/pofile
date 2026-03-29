// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestLintDiagnosticCode(t *testing.T) {
	t.Parallel()

	item := lint.Diagnostic{Code: "PO2009"}
	got := lintDiagnosticCode(item)
	if got != CodeLintPrintfMismatch {
		t.Fatalf("lintDiagnosticCode()=%d, want %d", got, CodeLintPrintfMismatch)
	}

	if got := lintDiagnosticCode(lint.Diagnostic{Code: ""}); got != 0 {
		t.Fatalf("lintDiagnosticCode(empty)=%d, want 0", got)
	}
}
