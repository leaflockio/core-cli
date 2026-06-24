// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth_test

import (
	"path/filepath"
	"testing"

	"github.com/leaflock/core-cli/internal/cli/auth"
	"github.com/leaflock/core-cli/internal/platform"
)

// TestNew_CI_uses_file verifies that a CI environment always selects the file backend
// regardless of OS.
func TestNew_CI_uses_file(t *testing.T) {
	plat := &platform.Platform{
		OS: platform.MacOS,
		CI: platform.CIInfo{Present: true},
	}
	store := auth.New(plat, filepath.Join(t.TempDir(), "credentials.json"))
	if got := store.Backend(); got != "file" {
		t.Errorf("Backend() = %q, want %q", got, "file")
	}
}

// TestNew_unknown_OS_uses_file verifies that an unrecognized OS falls back to the
// file backend.
func TestNew_unknown_OS_uses_file(t *testing.T) {
	plat := &platform.Platform{OS: platform.UnknownOS}
	store := auth.New(plat, filepath.Join(t.TempDir(), "credentials.json"))
	if got := store.Backend(); got != "file" {
		t.Errorf("Backend() = %q, want %q", got, "file")
	}
}

// TestNew_macos_no_CI_uses_keychain verifies that macOS without CI selects the
// keychain backend.
func TestNew_macos_no_CI_uses_keychain(t *testing.T) {
	plat := &platform.Platform{OS: platform.MacOS}
	store := auth.New(plat, filepath.Join(t.TempDir(), "credentials.json"))
	if got := store.Backend(); got != "keychain" {
		t.Errorf("Backend() = %q, want %q", got, "keychain")
	}
}

// TestNew_windows_no_CI_uses_keychain verifies that Windows without CI selects the
// keychain backend.
func TestNew_windows_no_CI_uses_keychain(t *testing.T) {
	plat := &platform.Platform{OS: platform.Windows}
	store := auth.New(plat, filepath.Join(t.TempDir(), "credentials.json"))
	if got := store.Backend(); got != "keychain" {
		t.Errorf("Backend() = %q, want %q", got, "keychain")
	}
}
