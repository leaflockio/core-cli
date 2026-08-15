// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package filename

import (
	"regexp"
	"testing"

	"github.com/leaflockio/core-cli/internal/vars"
)

func TestFileName_volatilityIsStable(t *testing.T) {
	if FileName.Volatility() != vars.Stable {
		t.Errorf("Volatility = %v, want Stable", FileName.Volatility())
	}
}

func TestFileName_patternMatchesRealisticFileNames(t *testing.T) {
	re := regexp.MustCompile("^" + FileName.Pattern() + "$")
	for _, tt := range []struct {
		in   string
		want bool
	}{
		{"main.go", true},
		{"file_name.go", true},
		{"README.md", true},
		{"", false},
		{"has space.go", false},
		{"has/slash.go", false},
	} {
		if got := re.MatchString(tt.in); got != tt.want {
			t.Errorf("pattern.MatchString(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestFileName_registersAndResolves(t *testing.T) {
	if FileName == nil {
		t.Fatal("FileName is nil — registration failed")
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	if err := Set(v, "/some/dir/main.go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := v.Resolve("{FILE_NAME}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "main.go" {
		t.Errorf("Resolve = %q, want %q", got, "main.go")
	}
}

func TestFileName_recomputesForEachSetCall(t *testing.T) {
	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}

	if err := Set(v, "/a/first.go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	first, err := v.Resolve("{FILE_NAME}")
	if err != nil {
		t.Fatalf("Resolve (1st): %v", err)
	}
	if first != "first.go" {
		t.Errorf("Resolve (1st) = %q, want %q", first, "first.go")
	}

	if err := Set(v, "/b/second.go"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	second, err := v.Resolve("{FILE_NAME}")
	if err != nil {
		t.Fatalf("Resolve (2nd): %v", err)
	}
	if second != "second.go" {
		t.Errorf("Resolve (2nd) = %q, want %q", second, "second.go")
	}
}

func TestFileName_errorsWithoutSetCall(t *testing.T) {
	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}

	if _, err := v.Resolve("{FILE_NAME}"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
