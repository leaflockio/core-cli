// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package paths_test

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/paths"
)

func TestNewRegistry_empty(t *testing.T) {
	r := paths.NewRegistry()
	if got := r.All(); len(got) != 0 {
		t.Errorf("All() = %v, want empty", got)
	}
}

func TestRegistry_Add_single(t *testing.T) {
	r := paths.NewRegistry()
	p := paths.KnownPath{Name: "manifest", Path: "/repo/leaf/manifest.yaml"}

	if err := r.Add(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := r.All()
	if len(got) != 1 || got[0] != p {
		t.Errorf("All() = %v, want [%v]", got, p)
	}
}

func TestRegistry_Add_variadic(t *testing.T) {
	r := paths.NewRegistry()
	p1 := paths.KnownPath{Name: "manifest"}
	p2 := paths.KnownPath{Name: "user-config"}
	p3 := paths.KnownPath{Name: "credentials"}

	if err := r.Add(p1, p2, p3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := r.All()
	if len(got) != 3 {
		t.Fatalf("All() length = %d, want 3", len(got))
	}
	if got[0] != p1 || got[1] != p2 || got[2] != p3 {
		t.Errorf("All() = %v, want [%v %v %v] in order", got, p1, p2, p3)
	}
}

func TestRegistry_Add_acrossCalls(t *testing.T) {
	r := paths.NewRegistry()
	if err := r.Add(paths.KnownPath{Name: "manifest"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Add(paths.KnownPath{Name: "user-config"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(r.All()); got != 2 {
		t.Errorf("All() length = %d, want 2", got)
	}
}

func TestRegistry_Add_duplicateName(t *testing.T) {
	r := paths.NewRegistry()
	if err := r.Add(paths.KnownPath{Name: "manifest", Path: "/first"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := r.Add(paths.KnownPath{Name: "manifest", Path: "/second"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.INT000 {
		t.Errorf("expected INT000, got %v", err)
	}

	// The original entry must still be the only one registered.
	got := r.All()
	if len(got) != 1 || got[0].Path != "/first" {
		t.Errorf("All() = %v, want only the first entry to survive", got)
	}
}

func TestRegistry_Add_duplicateWithinSameCall(t *testing.T) {
	r := paths.NewRegistry()

	err := r.Add(
		paths.KnownPath{Name: "manifest"},
		paths.KnownPath{Name: "manifest"},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// The first of the two duplicate entries was registered before the
	// second was rejected.
	if got := len(r.All()); got != 1 {
		t.Errorf("All() length = %d, want 1", got)
	}
}
