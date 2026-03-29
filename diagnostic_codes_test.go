// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"testing"

	"github.com/woozymasta/lintkit/lint"
)

func TestDiagnosticCatalogLookup(t *testing.T) {
	t.Parallel()

	spec, ok := DiagnosticByCode(CodeLintDuplicateEntry)
	if !ok {
		t.Fatalf("DiagnosticByCode(%q) ok=false, want true", CodeLintDuplicateEntry)
	}
	if spec.Code != CodeLintDuplicateEntry {
		t.Fatalf("spec.Code=%q, want %q", spec.Code, CodeLintDuplicateEntry)
	}
}

func TestLintRuleID(t *testing.T) {
	t.Parallel()

	spec, ok := DiagnosticByCode(CodeLintDuplicateEntry)
	if !ok {
		t.Fatalf("DiagnosticByCode(%q) ok=false, want true", CodeLintDuplicateEntry)
	}
	want := lint.BuildRuleID(
		LintModule,
		spec.Stage,
		spec.Message,
		CodeLintDuplicateEntry,
	)
	if got := LintRuleID(CodeLintDuplicateEntry); got != want {
		t.Fatalf("LintRuleID()=%q, want %q", got, want)
	}
	if got := LintRuleID(0); got != "pofile.unknown" {
		t.Fatalf("LintRuleID(empty)=%q, want %q", got, "pofile.unknown")
	}
}

func TestDiagnosticRuleSpec(t *testing.T) {
	t.Parallel()

	spec, ok := DiagnosticByCode(CodeLintDuplicateEntry)
	if !ok {
		t.Fatalf("DiagnosticByCode(%q) ok=false, want true", CodeLintDuplicateEntry)
	}

	ruleSpec, err := DiagnosticRuleSpec(spec)
	if err != nil {
		t.Fatalf("DiagnosticRuleSpec() error: %v", err)
	}
	if ruleSpec.ID == "" {
		t.Fatal("DiagnosticRuleSpec() returned empty rule ID")
	}
}
