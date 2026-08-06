// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"sync"
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
)

func TestBool_constructor(t *testing.T) {
	f := flags.Bool("verbose", "enable verbose output")
	if f.Name != "verbose" {
		t.Errorf("Name = %q, want %q", f.Name, "verbose")
	}
	if f.Usage != "enable verbose output" {
		t.Errorf("Usage = %q, want %q", f.Usage, "enable verbose output")
	}
}

func TestBoolValue_Validate_always_nil(t *testing.T) {
	if err := flags.Bool("verbose", "").Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestBoolValue_Default_zero(t *testing.T) {
	if flags.Bool("verbose", "").Default() != false {
		t.Error("Default() should be false before WithDefault")
	}
}

func TestBoolValue_WithDefault(t *testing.T) {
	f := flags.Bool("verbose", "").WithDefault(true)
	if f.Default() != true {
		t.Errorf("Default() = %v, want true", f.Default())
	}
}

func TestBoolValue_Dest_nil_by_default(t *testing.T) {
	if flags.Bool("verbose", "").Dest() != nil {
		t.Error("Dest() should be nil before WithDest")
	}
}

func TestBoolValue_WithDest(t *testing.T) {
	var dest bool
	f := flags.Bool("verbose", "").WithDest(&dest)
	if f.Dest() != &dest {
		t.Error("Dest() should point to the bound variable")
	}
}

// TestBoolValue_WithDest_independentAcrossCalls guards against a shared
// package-level flag var (e.g. one command flag reused across several
// commands) having one command's WithDest binding silently overwritten by
// another's — WithDest must return an independent copy, not mutate the
// shared receiver in place.
func TestBoolValue_WithDest_independentAcrossCalls(t *testing.T) {
	base := flags.Bool("verbose", "")
	var destA, destB bool
	a := base.WithDest(&destA)
	b := base.WithDest(&destB)

	if a.Dest() != &destA {
		t.Errorf("a.Dest() = %p, want %p", a.Dest(), &destA)
	}
	if b.Dest() != &destB {
		t.Errorf("b.Dest() = %p, want %p", b.Dest(), &destB)
	}
	if a.Dest() == b.Dest() {
		t.Error("a and b should have independent Dest pointers")
	}
	if base.Dest() != nil {
		t.Error("the original base value should be unaffected by either call")
	}
}

// TestBoolValue_WithDest_concurrentCallsAreRaceFree calls WithDest on the
// same shared base from many goroutines at once — the scenario the
// sequential independence test above can't exercise. Run with -race: this
// would fail on the old mutate-in-place implementation and passes on the
// copy-on-write one.
func TestBoolValue_WithDest_concurrentCallsAreRaceFree(t *testing.T) {
	base := flags.Bool("verbose", "")
	const n = 50
	destinations := make([]bool, n)
	results := make([]*flags.BoolValue, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			results[i] = base.WithDest(&destinations[i])
		})
	}
	wg.Wait()

	for i := range n {
		if results[i].Dest() != &destinations[i] {
			t.Errorf("results[%d].Dest() = %p, want %p", i, results[i].Dest(), &destinations[i])
		}
	}
}

func TestBoolValue_WithShorthand(t *testing.T) {
	f := flags.Bool("verbose", "").WithShorthand("v")
	if f.Shorthand != "v" {
		t.Errorf("Shorthand = %q, want %q", f.Shorthand, "v")
	}
}

// boolResolverStub satisfies BoolResolver for interface verification.
type boolResolverStub struct {
	flags.CommandFlag[*flags.BoolValue]
}

func (s boolResolverStub) IsResolver()        {}
func (s boolResolverStub) Resolve(bool) error { return nil }

func TestBoolResolver_satisfied_by_implementation(t *testing.T) {
	var _ flags.BoolResolver = boolResolverStub{}
}
