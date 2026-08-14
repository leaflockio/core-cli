// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package yy

import (
	"fmt"
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/vars"
)

func TestCompute_returnsCurrentTwoDigitYear(t *testing.T) {
	got, err := compute(nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	want := fmt.Sprintf("%02d", time.Now().Year()%100)
	if got != want {
		t.Errorf("compute = %q, want %q", got, want)
	}
}

func TestYY_registersAndResolves(t *testing.T) {
	if YY == nil {
		t.Fatal("YY is nil — registration failed")
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	got, err := v.Resolve("{YY}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := fmt.Sprintf("%02d", time.Now().Year()%100)
	if got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}
