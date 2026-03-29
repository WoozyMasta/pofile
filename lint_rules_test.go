// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

import (
	"errors"
	"testing"

	"github.com/woozymasta/lintkit/lint"
	"github.com/woozymasta/lintkit/linttest"
)

type lintRuleTestRegistrar struct {
	rules []lint.RuleRunner
}

// Register stores runners in registrar test double.
func (registrar *lintRuleTestRegistrar) Register(
	runners ...lint.RuleRunner,
) error {
	registrar.rules = append(registrar.rules, runners...)

	return nil
}

func TestRegisterLintRulesNilRegistrar(t *testing.T) {
	t.Parallel()

	if err := RegisterLintRules(nil); !errors.Is(err, ErrNilLintRuleRegistrar) {
		t.Fatalf(
			"RegisterLintRules(nil) error=%v, want ErrNilLintRuleRegistrar",
			err,
		)
	}
}

func TestRegisterLintRules(t *testing.T) {
	t.Parallel()

	registrar := &lintRuleTestRegistrar{
		rules: make([]lint.RuleRunner, 0),
	}
	if err := RegisterLintRules(registrar); err != nil {
		t.Fatalf("RegisterLintRules() error: %v", err)
	}
	if len(registrar.rules) != len(DiagnosticCatalog()) {
		t.Fatalf(
			"registered rules=%d, want %d",
			len(registrar.rules),
			len(DiagnosticCatalog()),
		)
	}
}

func TestRegisterLintRulesByScope(t *testing.T) {
	t.Parallel()

	registrar := &lintRuleTestRegistrar{
		rules: make([]lint.RuleRunner, 0),
	}
	if err := RegisterLintRulesByScope(registrar, "parse"); err != nil {
		t.Fatalf("RegisterLintRulesByScope() error: %v", err)
	}
	if len(registrar.rules) == 0 {
		t.Fatal("RegisterLintRulesByScope() registered 0 rules, want >0")
	}
}

func TestRegisterLintRulesByStage(t *testing.T) {
	t.Parallel()

	registrar := &lintRuleTestRegistrar{
		rules: make([]lint.RuleRunner, 0),
	}
	if err := RegisterLintRulesByStage(registrar, StageLint); err != nil {
		t.Fatalf("RegisterLintRulesByStage() error: %v", err)
	}
	if len(registrar.rules) == 0 {
		t.Fatal("RegisterLintRulesByStage() registered 0 rules, want >0")
	}
}

func TestLintRuleSpecsMatchCatalog(t *testing.T) {
	t.Parallel()

	linttest.AssertCatalogContract(
		t,
		LintModule,
		DiagnosticCatalog(),
		LintRuleSpecs(),
		LintRuleID,
	)
}

func TestAttachLintDiagnostics(t *testing.T) {
	t.Parallel()

	run := lint.RunContext{}
	diagnostics := []lint.Diagnostic{
		{
			RuleID:   LintRuleID(CodeLintDuplicateEntry),
			Code:     publicLintCode(CodeLintDuplicateEntry),
			Severity: lint.SeverityError,
			Message:  "duplicate",
			Start: lint.Position{
				Line:   3,
				Column: 2,
				Offset: 11,
			},
		},
	}

	AttachLintDiagnostics(&run, diagnostics)

	grouped, ok := lint.GetIndexedByCode[lint.Diagnostic, lint.Code](
		&run,
		lintRunValueByCodeKey,
	)
	if !ok {
		t.Fatal("GetIndexedByCode() ok=false, want true")
	}
	if len(grouped[CodeLintDuplicateEntry]) != 1 {
		t.Fatalf(
			"grouped[%q] len=%d, want 1",
			CodeLintDuplicateEntry,
			len(grouped[CodeLintDuplicateEntry]),
		)
	}
}
