// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// captureStderr runs fn and returns everything written to os.Stderr during that call.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = old })

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

func TestRoundForDisplay(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{"sub-millisecond exact", 2 * time.Microsecond, 2 * time.Microsecond},
		{"sub-millisecond rounds up", 999 * time.Nanosecond, time.Microsecond},
		{"millisecond boundary goes to ms branch", time.Millisecond, time.Millisecond},
		{"sub-second exact", 500 * time.Millisecond, 500 * time.Millisecond},
		{"sub-second rounds down", 1234 * time.Microsecond, time.Millisecond},
		{"second boundary goes to default branch", time.Second, time.Second},
		{"seconds exact", 1200 * time.Millisecond, 1200 * time.Millisecond},
		{"seconds rounds to nearest 10ms", 2345 * time.Millisecond, 2350 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundForDisplay(tt.in); got != tt.want {
				t.Errorf("roundForDisplay(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestPrintDuration(t *testing.T) {
	out := captureStderr(t, func() { printDuration(1234 * time.Microsecond) })

	want := "\ndone in 1ms (1.234ms)\n"
	if out != want {
		t.Errorf("printDuration output = %q, want %q", out, want)
	}
}
