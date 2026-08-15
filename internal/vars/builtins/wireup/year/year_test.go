// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package year

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/leaflockio/core-cli/internal/vars"
)

func TestCompute_returnsCurrentYear(t *testing.T) {
	got, err := compute(nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	want := strconv.Itoa(time.Now().Year())
	if got != want {
		t.Errorf("compute = %q, want %q", got, want)
	}
}

func TestYear_volatilityIsVolatile(t *testing.T) {
	if Year.Volatility() != vars.Volatile {
		t.Errorf("Volatility = %v, want Volatile", Year.Volatility())
	}
}

func TestYear_patternMatchesYearShapedValues(t *testing.T) {
	re := regexp.MustCompile("^" + Year.Pattern() + "$")
	for _, tt := range []struct {
		in   string
		want bool
	}{
		{"2026", true},
		{"2020-2026", true},
		{"", false},
		{"26", false},
		{"not-a-year", false},
	} {
		if got := re.MatchString(tt.in); got != tt.want {
			t.Errorf("pattern.MatchString(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestYear_registersAndResolves(t *testing.T) {
	if Year == nil {
		t.Fatal("Year is nil — registration failed")
	}

	v, err := vars.New(nil)
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	got, err := v.Resolve("{YEAR}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := strconv.Itoa(time.Now().Year())
	if got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}
