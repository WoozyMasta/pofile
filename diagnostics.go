// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/pofile

package pofile

// newDiagnostic builds one diagnostic with span derived from position.
func newDiagnostic(
	severity Severity,
	code string,
	message string,
	position Position,
) Diagnostic {
	return Diagnostic{
		Severity: severity,
		Code:     code,
		Message:  message,
		Position: position,
		Span:     spanAtPosition(position),
	}
}

// spanAtPosition builds a minimal non-empty span for a source position.
func spanAtPosition(position Position) Span {
	start := max(position.Offset, 0)

	return Span{
		StartOffset: start,
		EndOffset:   start + 1,
	}
}
