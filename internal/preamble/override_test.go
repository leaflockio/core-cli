// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

// snapshotOverrides registers a t.Cleanup that restores patterns and
// preserveLines to their state when called — so a test invoking
// ApplyOverrides never leaks state into any other test in this package.
func snapshotOverrides(t *testing.T) {
	t.Helper()
	n := len(patterns)
	prevPreserveLines := preserveLines
	t.Cleanup(func() {
		patterns = patterns[:n]
		preserveLines = prevPreserveLines
	})
}

func TestApplyOverrides_patternIsAppliedByDetect(t *testing.T) {
	snapshotOverrides(t)

	lines := []string{"# custom-header-marker", "code here"}
	if Detect(lines, "f.py") != 0 {
		t.Fatal("precondition failed: content already matches something")
	}

	warnings, err := ApplyOverrides([]PatternOverride{{Pattern: `^# custom-header-marker`}}, nil)
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}
	if warnings != nil {
		t.Errorf("warnings = %v, want nil", warnings)
	}

	if got := Detect(lines, "f.py"); got != 1 {
		t.Errorf("Detect after ApplyOverrides = %d, want 1", got)
	}
}

func TestApplyOverrides_patternRespectsExtScope(t *testing.T) {
	snapshotOverrides(t)

	_, err := ApplyOverrides([]PatternOverride{
		{Pattern: `^# scoped-marker`, Exts: []string{"rb"}},
	}, nil)
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}

	lines := []string{"# scoped-marker"}
	if got := Detect(lines, "f.rb"); got != 1 {
		t.Errorf("Detect(in scope) = %d, want 1", got)
	}
	if got := Detect(lines, "f.py"); got != 0 {
		t.Errorf("Detect(out of scope) = %d, want 0", got)
	}
}

func TestApplyOverrides_patternsAreAdditiveToBuiltins(t *testing.T) {
	snapshotOverrides(t)

	_, err := ApplyOverrides([]PatternOverride{{Pattern: `^# custom`}}, nil)
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}

	// A built-in marker (shebang) still matches after a custom pattern
	// is added.
	if got := Detect([]string{"#!/bin/sh"}, "f.sh"); got != 1 {
		t.Errorf("Detect(shebang) after ApplyOverrides = %d, want 1", got)
	}
}

func TestApplyOverrides_invalidPatternIsCallerError(t *testing.T) {
	snapshotOverrides(t)
	before := len(patterns)

	_, err := ApplyOverrides([]PatternOverride{{Pattern: `[unclosed`}}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.PRE001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.PRE001)
	}
	if len(patterns) != before {
		t.Errorf("len(patterns) = %d, want unchanged %d", len(patterns), before)
	}
}

func TestApplyOverrides_nonPositivePreserveLinesIsCallerError(t *testing.T) {
	snapshotOverrides(t)

	tests := []int{0, -1, -5}
	for _, n := range tests {
		_, err := ApplyOverrides(nil, map[string]int{"*.go": n})
		if err == nil {
			t.Fatalf("PreserveLines=%d: expected error, got nil", n)
		}
		var e *errs.Error
		if !errors.As(err, &e) {
			t.Fatalf("PreserveLines=%d: error is not an *errs.Error: %v", n, err)
		}
		if e.Code != errs.PRE002 {
			t.Errorf("PreserveLines=%d: Code = %q, want %q", n, e.Code, errs.PRE002)
		}
	}
}

func TestApplyOverrides_allOrNothingAcrossPatternsAndPerFile(t *testing.T) {
	snapshotOverrides(t)
	beforePatterns := len(patterns)

	_, err := ApplyOverrides(
		[]PatternOverride{{Pattern: `^# valid-pattern`}},
		map[string]int{"*.go": 0},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(patterns) != beforePatterns {
		t.Errorf("len(patterns) = %d, want unchanged %d (pattern from failed call must not apply)",
			len(patterns), beforePatterns)
	}
	if got := Detect([]string{"# valid-pattern"}, "f.go"); got != 0 {
		t.Errorf("Detect = %d, want 0 (pattern from failed ApplyOverrides call must not be active)", got)
	}
}

func TestApplyOverrides_warningsForLargeValues(t *testing.T) {
	snapshotOverrides(t)

	warnings, err := ApplyOverrides(nil, map[string]int{
		"small.go": 5,
		"large.go": largePreserveLinesThreshold + 1,
		"huge.go":  largePreserveLinesThreshold + 100,
	})
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}
	if len(warnings) != 2 {
		t.Fatalf("len(warnings) = %d, want 2: %v", len(warnings), warnings)
	}
	// Sorted by pattern: "huge.go" < "large.go".
	if warnings[0] != "per_file entry huge.go has PreserveLines 110, unusually large for a preamble" {
		t.Errorf("warnings[0] = %q", warnings[0])
	}
	if warnings[1] != "per_file entry large.go has PreserveLines 11, unusually large for a preamble" {
		t.Errorf("warnings[1] = %q", warnings[1])
	}
}

func TestApplyOverrides_thresholdValueItselfDoesNotWarn(t *testing.T) {
	snapshotOverrides(t)

	warnings, err := ApplyOverrides(nil, map[string]int{"f.go": largePreserveLinesThreshold})
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}
	if warnings != nil {
		t.Errorf("warnings = %v, want nil at exactly the threshold", warnings)
	}
}

func TestDetect_preserveLinesBypassesMarkersAndPatterns(t *testing.T) {
	snapshotOverrides(t)

	_, err := ApplyOverrides(nil, map[string]int{"*.go": 2})
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}

	// No marker or pattern would normally match these lines, but the
	// PreserveLines override should still claim the first 2 lines.
	lines := []string{"ordinary code", "more ordinary code", "package main"}
	if got := Detect(lines, "main.go"); got != 2 {
		t.Errorf("Detect = %d, want 2", got)
	}
}

func TestDetect_preserveLinesBoundedByLineCount(t *testing.T) {
	snapshotOverrides(t)

	_, err := ApplyOverrides(nil, map[string]int{"*.go": 100})
	if err != nil {
		t.Fatalf("ApplyOverrides: %v", err)
	}

	lines := []string{"one", "two"}
	if got := Detect(lines, "main.go"); got != len(lines) {
		t.Errorf("Detect = %d, want %d (bounded by len(lines))", got, len(lines))
	}
}

func TestResolvePreserveLines_doublestarMatch(t *testing.T) {
	snapshotOverrides(t)
	preserveLines = map[string]int{"vendor/**": 3, "**/*.pb.go": 5}

	if got := resolvePreserveLines("vendor/pkg/file.go"); got != 3 {
		t.Errorf("resolvePreserveLines(vendor/pkg/file.go) = %d, want 3", got)
	}
	if got := resolvePreserveLines("api/service.pb.go"); got != 5 {
		t.Errorf("resolvePreserveLines(api/service.pb.go) = %d, want 5", got)
	}
	if got := resolvePreserveLines("main.go"); got != 0 {
		t.Errorf("resolvePreserveLines(main.go) = %d, want 0 (no match)", got)
	}
}

func TestResolvePreserveLines_deterministicOnMultipleMatches(t *testing.T) {
	snapshotOverrides(t)
	// Both patterns match "main.go"; "*.go" sorts before "main.go"
	// alphabetically, so it should win.
	preserveLines = map[string]int{"main.go": 9, "*.go": 4}

	if got := resolvePreserveLines("main.go"); got != 4 {
		t.Errorf("resolvePreserveLines(main.go) = %d, want 4 (alphabetically first match)", got)
	}
}

func TestResolvePreserveLines_emptyMap(t *testing.T) {
	snapshotOverrides(t)
	preserveLines = nil

	if got := resolvePreserveLines("main.go"); got != 0 {
		t.Errorf("resolvePreserveLines(main.go) = %d, want 0", got)
	}
}
