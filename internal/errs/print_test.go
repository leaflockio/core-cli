// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package errs_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
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

// resetDisplay restores the default DisplayConfig after each test so tests
// don't bleed state into one another.
func resetDisplay(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		errs.Configure(errs.DisplayConfig{
			ShowCode:       true,
			ShowResolution: true,
			ShowUnderlying: false,
		})
	})
}

func TestPrint_nil(t *testing.T) {
	out := captureStderr(t, func() { errs.Print(nil) })
	if out != "" {
		t.Errorf("Print(nil) wrote %q, want empty", out)
	}
}

func TestPrint_plainError(t *testing.T) {
	out := captureStderr(t, func() {
		errs.Print(errUnderlying)
	})
	if !strings.Contains(out, "underlying error") {
		t.Errorf("expected plain error message in output, got %q", out)
	}
}

func TestPrint_showCode(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowCode: true})

	e := errs.Caller(testCode, "something failed", nil)
	out := captureStderr(t, func() { errs.Print(e) })

	if !strings.Contains(out, string(testCode)) {
		t.Errorf("expected code %q in output, got %q", testCode, out)
	}
}

func TestPrint_hideCode(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowCode: false})

	e := errs.Caller(testCode, "something failed", nil)
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, string(testCode)) {
		t.Errorf("expected code %q to be hidden, got %q", testCode, out)
	}
	if !strings.Contains(out, "something failed") {
		t.Errorf("expected message in output, got %q", out)
	}
}

func TestPrint_showResolution(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowResolution: true})

	e := errs.Caller(testCode, "msg", nil, errs.Context{
		Cause:      "the cause",
		Resolution: "the fix",
	})
	out := captureStderr(t, func() { errs.Print(e) })

	if !strings.Contains(out, "the fix") {
		t.Errorf("expected resolution in output, got %q", out)
	}
}

func TestPrint_hideResolution(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowResolution: false})

	e := errs.Caller(testCode, "msg", nil, errs.Context{
		Cause:      "the cause",
		Resolution: "the fix",
	})
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, "the fix") {
		t.Errorf("expected resolution to be hidden, got %q", out)
	}
	if !strings.Contains(out, "the cause") {
		t.Errorf("expected cause in output, got %q", out)
	}
}

func TestPrint_showUnderlying(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowUnderlying: true})

	e := errs.Caller(testCode, "msg", errUnderlying)
	out := captureStderr(t, func() { errs.Print(e) })

	if !strings.Contains(out, errUnderlying.Error()) {
		t.Errorf("expected underlying error in output, got %q", out)
	}
}

func TestPrint_hideUnderlying(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowUnderlying: false})

	e := errs.Caller(testCode, "msg", errUnderlying)
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, errUnderlying.Error()) {
		t.Errorf("expected underlying error to be hidden, got %q", out)
	}
}

func TestPrint_nilUnderlying_showUnderlyingEnabled(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowUnderlying: true})

	// ShowUnderlying=true but Err is nil — no underlying line should appear.
	e := errs.Caller(testCode, "msg", nil)
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, "underlying:") {
		t.Errorf("expected no underlying line when Err is nil, got %q", out)
	}
}

func TestPrint_emptyCause_skipped(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowResolution: true})

	e := errs.Caller(testCode, "msg", nil, errs.Context{
		Cause:      "",
		Resolution: "the fix",
	})
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, "cause:") {
		t.Errorf("expected empty cause to be skipped, got %q", out)
	}
}

func TestPrint_multipleContexts(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowResolution: true})

	e := errs.Caller(testCode, "msg", nil,
		errs.Context{Cause: "cause one", Resolution: "fix one"},
		errs.Context{Cause: "cause two", Resolution: "fix two"},
	)
	out := captureStderr(t, func() { errs.Print(e) })

	if !strings.Contains(out, "cause one") {
		t.Errorf("expected first cause in output, got %q", out)
	}
	if !strings.Contains(out, "cause two") {
		t.Errorf("expected second cause in output, got %q", out)
	}
}

func TestConfigure_updatesActiveSettings(t *testing.T) {
	resetDisplay(t)
	errs.Configure(errs.DisplayConfig{ShowCode: false, ShowResolution: false})

	e := errs.Caller(testCode, "msg", nil, errs.Context{Cause: "the cause", Resolution: "the resolution hint"})
	out := captureStderr(t, func() { errs.Print(e) })

	if strings.Contains(out, string(testCode)) {
		t.Errorf("ShowCode=false: expected code hidden, got %q", out)
	}
	if strings.Contains(out, "the resolution hint") {
		t.Errorf("ShowResolution=false: expected resolution hidden, got %q", out)
	}
}
