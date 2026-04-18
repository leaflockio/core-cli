// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultLogPath_xdgStateHomeSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(xdgStateHomeEnvVar, dir)

	got := defaultLogPath()
	want := filepath.Join(dir, AppName)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultLogPath_xdgStateHomeUnset(t *testing.T) {
	t.Setenv(xdgStateHomeEnvVar, "")

	home := t.TempDir()
	old := osUserHomeDir
	osUserHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { osUserHomeDir = old })

	got := defaultLogPath()
	want := filepath.Join(home, xdgDefaultStateDir, AppName)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultLogPath_homeDirError(t *testing.T) {
	t.Setenv(xdgStateHomeEnvVar, "")

	old := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errHomeDirFail }
	t.Cleanup(func() { osUserHomeDir = old })

	got := defaultLogPath()
	if got != "." {
		t.Errorf("got %q, want %q", got, ".")
	}
}
