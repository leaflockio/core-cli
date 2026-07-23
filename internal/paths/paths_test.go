// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package paths_test

import (
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/paths"
)

func TestScope_String(t *testing.T) {
	tests := []struct {
		scope paths.Scope
		want  string
	}{
		{paths.ScopeUser, "user"},
		{paths.ScopeProject, "project"},
		{paths.ScopeCache, "cache"},
		{paths.Scope(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.scope.String(); got != tt.want {
			t.Errorf("Scope(%d).String() = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

func TestKnownPath_Print_brief(t *testing.T) {
	p := paths.KnownPath{
		Name:  "manifest",
		Path:  "/repo/leaf/manifest.yaml",
		Desc:  "flat project config",
		Scope: paths.ScopeProject,
	}
	got := p.Print(true)
	if got != p.Path {
		t.Errorf("Print(true) = %q, want %q", got, p.Path)
	}
}

func TestKnownPath_Print_full(t *testing.T) {
	p := paths.KnownPath{
		Name:  "manifest",
		Path:  "/repo/leaf/manifest.yaml",
		Desc:  "flat project config",
		Scope: paths.ScopeProject,
	}
	got := p.Print(false)
	for _, want := range []string{p.Name, p.Path, p.Desc, "project"} {
		if !strings.Contains(got, want) {
			t.Errorf("Print(false) = %q, want it to contain %q", got, want)
		}
	}
}
