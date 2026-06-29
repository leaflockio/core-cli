// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

//go:build integration

package auth_test

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/leaflock/core-cli/internal/cli/commands/auth"
	"github.com/leaflock/core-cli/internal/errs"
	"github.com/leaflock/core-cli/internal/platform"
)

// keychainTestStore returns a keychain-backed Store for the current OS, or calls
// t.Skip if the keychain is not available.
//
// On Linux, New() probes the keyring daemon and falls back to fileStore if the
// daemon is absent. Checking Backend() after construction catches that case and
// skips rather than running against a file backend.
func keychainTestStore(t *testing.T) auth.Store {
	t.Helper()

	var os platform.OS
	switch runtime.GOOS {
	case "darwin":
		os = platform.MacOS
	case "linux":
		os = platform.Linux
	case "windows":
		os = platform.Windows
	default:
		t.Skipf("keychain not supported on %s", runtime.GOOS)
	}

	store := auth.New(&platform.Platform{OS: os}, "")
	if store.Backend() != "keychain" {
		t.Skip("keychain daemon not available on this system")
	}
	return store
}

// saveAndRestore reads whatever is currently in the keychain and registers a
// cleanup that puts it back (or deletes it if there was nothing). This keeps
// integration tests non-destructive for developers who have real credentials stored.
func saveAndRestore(t *testing.T, store auth.Store) {
	t.Helper()
	existing, err := store.Get()
	var e *errs.Error
	if err != nil && !(errors.As(err, &e) && e.Code == errs.AUT001) {
		t.Fatalf("saveAndRestore: unexpected error reading existing credentials: %v", err)
	}
	t.Cleanup(func() {
		if existing != nil {
			if err := store.Set(existing); err != nil {
				t.Errorf("saveAndRestore: failed to restore credentials: %v", err)
			}
		} else {
			if err := store.Delete(); err != nil {
				t.Errorf("saveAndRestore: failed to clean up test credentials: %v", err)
			}
		}
	})
}

// TestKeychainStore_Get_not_found verifies AUT001 when no entry exists in the keychain.
func TestKeychainStore_Get_not_found(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete (setup): %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT001)
}

// TestKeychainStore_Set_Get_roundtrip verifies that credentials survive a Set → Get cycle.
func TestKeychainStore_Set_Get_roundtrip(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	want := &auth.Credentials{
		Token:     "test-token-abc",
		Workspace: "test-workspace",
		ExpiresAt: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	if err := store.Set(want); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := store.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Token != want.Token {
		t.Errorf("Token = %q, want %q", got.Token, want.Token)
	}
	if got.Workspace != want.Workspace {
		t.Errorf("Workspace = %q, want %q", got.Workspace, want.Workspace)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want.ExpiresAt)
	}
}

// TestKeychainStore_Get_expired verifies AUT005 when the stored credentials have expired.
func TestKeychainStore_Get_expired(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	expired := &auth.Credentials{
		Token:     "test-token-expired",
		Workspace: "test-workspace",
		ExpiresAt: time.Now().UTC().Add(-time.Hour).Truncate(time.Second),
	}
	if err := store.Set(expired); err != nil {
		t.Fatalf("Set: %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT005)
}

// TestKeychainStore_Set_overwrites verifies that a second Set replaces the first entry.
func TestKeychainStore_Set_overwrites(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	first := &auth.Credentials{Token: "test-token-first", Workspace: "test-workspace"}
	if err := store.Set(first); err != nil {
		t.Fatalf("Set (first): %v", err)
	}
	second := &auth.Credentials{
		Token:     "test-token-second",
		Workspace: "test-workspace-beta",
		ExpiresAt: time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second),
	}
	if err := store.Set(second); err != nil {
		t.Fatalf("Set (second): %v", err)
	}
	got, err := store.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Token != second.Token {
		t.Errorf("Token = %q, want %q", got.Token, second.Token)
	}
}

// TestKeychainStore_Delete_removes verifies that Delete removes the entry and a
// subsequent Get returns AUT001.
func TestKeychainStore_Delete_removes(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	if err := store.Set(&auth.Credentials{Token: "test-token-delete", Workspace: "test-workspace"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT001)
}

// TestKeychainStore_Delete_idempotent verifies that Delete returns nil when no
// entry exists.
func TestKeychainStore_Delete_idempotent(t *testing.T) {
	store := keychainTestStore(t)
	saveAndRestore(t, store)

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete (setup): %v", err)
	}
	if err := store.Delete(); err != nil {
		t.Errorf("Delete on missing entry should return nil, got: %v", err)
	}
}
