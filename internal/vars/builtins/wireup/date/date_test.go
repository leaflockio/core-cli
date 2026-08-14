// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package date

import (
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/vars"
)

func TestCompute_returnsCurrentDateISO8601(t *testing.T) {
	got, err := compute(nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	want := time.Now().Format(time.DateOnly)
	if got != want {
		t.Errorf("compute = %q, want %q", got, want)
	}
}

func TestDate_registersAndResolves(t *testing.T) {
	if Date == nil {
		t.Fatal("Date is nil — registration failed")
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	got, err := v.Resolve("{DATE}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := time.Now().Format(time.DateOnly)
	if got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}
