// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"strings"
	"testing"
)

func TestFormatPrefix(t *testing.T) {
	got := formatPrefix("internal/config")
	want := "[internal/config]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResetDebug_off(t *testing.T) {
	oldLevel := debugLevel
	oldPrefix := debugPkgPrefix
	t.Cleanup(func() {
		debugLevel = oldLevel
		debugPkgPrefix = oldPrefix
	})

	resetDebug("0")

	if debugLevel != debugLevelOff {
		t.Errorf("debugLevel = %d, want %d", debugLevel, debugLevelOff)
	}
	if debugPkgPrefix != "" {
		t.Errorf("debugPkgPrefix = %q, want empty", debugPkgPrefix)
	}
}

func TestResetDebug_pkg(t *testing.T) {
	oldLevel := debugLevel
	oldPrefix := debugPkgPrefix
	t.Cleanup(func() {
		debugLevel = oldLevel
		debugPkgPrefix = oldPrefix
	})

	resetDebug("1")

	if debugLevel != debugLevelPkg {
		t.Errorf("debugLevel = %d, want %d", debugLevel, debugLevelPkg)
	}
	if !strings.Contains(debugPkgPrefix, "internal/config") {
		t.Errorf("debugPkgPrefix = %q, want to contain internal/config", debugPkgPrefix)
	}
}

func TestResetDebug_invalid(t *testing.T) {
	oldLevel := debugLevel
	oldPrefix := debugPkgPrefix
	t.Cleanup(func() {
		debugLevel = oldLevel
		debugPkgPrefix = oldPrefix
	})

	const invalidLevel = "not-a-number"
	out := captureStderr(t, func() {
		resetDebug(invalidLevel)
	})

	if debugLevel != debugLevelOff {
		t.Errorf("debugLevel = %d, want %d (invalid input should disable debug)", debugLevel, debugLevelOff)
	}
	if !strings.Contains(out, envVarConfigDebug) {
		t.Errorf("expected env var name in warning output, got %q", out)
	}
	if !strings.Contains(out, invalidLevel) {
		t.Errorf("expected invalid value in warning output, got %q", out)
	}
}

func TestDebugf_off(t *testing.T) {
	oldLevel := debugLevel
	debugLevel = debugLevelOff
	t.Cleanup(func() { debugLevel = oldLevel })

	out := captureStderr(t, func() {
		debugf(levelDebug, "should not appear")
	})
	if out != "" {
		t.Errorf("expected no output at level off, got %q", out)
	}
}

func TestDebugf_pkgLevel_debug(t *testing.T) {
	oldLevel := debugLevel
	oldPrefix := debugPkgPrefix
	debugLevel = debugLevelPkg
	debugPkgPrefix = "[internal/config]"
	t.Cleanup(func() {
		debugLevel = oldLevel
		debugPkgPrefix = oldPrefix
	})

	out := captureStderr(t, func() {
		debugf(levelDebug, "search paths: %v", []string{"/tmp"})
	})

	if !strings.Contains(out, "[internal/config]") {
		t.Errorf("expected prefix in output, got %q", out)
	}
	if !strings.Contains(out, "DEBUG") {
		t.Errorf("expected DEBUG level in output, got %q", out)
	}
	if !strings.Contains(out, "search paths:") {
		t.Errorf("expected message in output, got %q", out)
	}
}

func TestDebugf_pkgLevel_warn(t *testing.T) {
	oldLevel := debugLevel
	oldPrefix := debugPkgPrefix
	debugLevel = debugLevelPkg
	debugPkgPrefix = "[internal/config]"
	t.Cleanup(func() {
		debugLevel = oldLevel
		debugPkgPrefix = oldPrefix
	})

	out := captureStderr(t, func() {
		debugf(levelWarn, "no config found")
	})

	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN level in output, got %q", out)
	}
	if !strings.Contains(out, "no config found") {
		t.Errorf("expected message in output, got %q", out)
	}
}

func TestDebugf_fileLevel(t *testing.T) {
	oldLevel := debugLevel
	debugLevel = debugLevelFile
	t.Cleanup(func() { debugLevel = oldLevel })

	out := captureStderr(t, func() {
		debugf(levelDebug, "file level message")
	})

	if !strings.Contains(out, "debug_test.go") {
		t.Errorf("expected calling file in prefix, got %q", out)
	}
	if !strings.Contains(out, "DEBUG") {
		t.Errorf("expected DEBUG in output, got %q", out)
	}
}
