// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leaflock/core-cli/internal/cli/commands/auth"
	"github.com/leaflock/core-cli/internal/errs"
	"github.com/leaflock/core-cli/internal/platform"
)

// newFileStore returns a fileStore-backed Store pointed at a temp path, and the
// path itself so individual tests can manipulate the file directly.
func newFileStore(t *testing.T) (auth.Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	plat := &platform.Platform{CI: platform.CIInfo{Present: true}}
	return auth.New(plat, path), path
}

// validCreds returns a non-expired Credentials value for use in tests.
func validCreds() *auth.Credentials {
	return &auth.Credentials{
		Token:     "tok-abc123",
		Workspace: "acme",
		ExpiresAt: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
}

// TestFileStore_Backend verifies the backend label.
func TestFileStore_Backend(t *testing.T) {
	store, _ := newFileStore(t)
	if got := store.Backend(); got != "file" {
		t.Errorf("Backend() = %q, want %q", got, "file")
	}
}

// TestFileStore_Get_not_found verifies AUT001 when no credentials file exists.
func TestFileStore_Get_not_found(t *testing.T) {
	store, _ := newFileStore(t)
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT001)
}

// TestFileStore_Get_corrupt verifies AUT002 when the credentials file contains invalid JSON.
func TestFileStore_Get_corrupt(t *testing.T) {
	store, path := newFileStore(t)
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT002)
}

// TestFileStore_Get_unreadable verifies AUT002 when the credentials file exists but
// cannot be read due to permissions.
func TestFileStore_Get_unreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permission checks are ineffective when running as root")
	}
	store, path := newFileStore(t)
	if err := store.Set(validCreds()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(path, 0o600) })

	_, err := store.Get()
	assertErrCode(t, err, errs.AUT002)
}

// TestFileStore_Get_expired verifies AUT005 when stored credentials have passed their expiry.
func TestFileStore_Get_expired(t *testing.T) {
	store, _ := newFileStore(t)
	expired := &auth.Credentials{
		Token:     "tok-expired",
		Workspace: "acme",
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}
	if err := store.Set(expired); err != nil {
		t.Fatalf("Set: %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT005)
}

// TestFileStore_Get_valid verifies that valid credentials round-trip correctly.
func TestFileStore_Get_valid(t *testing.T) {
	store, _ := newFileStore(t)
	want := validCreds()
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

// TestFileStore_Set_creates_parent_dir verifies that Set creates the parent directory
// if it does not exist.
func TestFileStore_Set_creates_parent_dir(t *testing.T) {
	dir := t.TempDir()
	// path includes a nested directory that does not exist yet
	path := filepath.Join(dir, "nested", "deep", "credentials.json")
	plat := &platform.Platform{CI: platform.CIInfo{Present: true}}
	store := auth.New(plat, path)

	if err := store.Set(validCreds()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected credentials file to exist after Set: %v", err)
	}
}

// TestFileStore_Set_file_permissions verifies the credentials file is written with 0600.
func TestFileStore_Set_file_permissions(t *testing.T) {
	store, path := newFileStore(t)
	if err := store.Set(validCreds()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("file permissions = %04o, want 0600", got)
	}
}

// TestFileStore_Set_overwrites_existing verifies that a second Set replaces the
// first credentials without error.
func TestFileStore_Set_overwrites_existing(t *testing.T) {
	store, _ := newFileStore(t)
	first := validCreds()
	if err := store.Set(first); err != nil {
		t.Fatalf("Set (first): %v", err)
	}
	second := &auth.Credentials{
		Token:     "tok-new",
		Workspace: "beta",
		ExpiresAt: time.Now().UTC().Add(2 * time.Hour),
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

// TestFileStore_Set_not writable_dir verifies AUT003 when the credentials directory
// cannot be written to.
func TestFileStore_Set_dir_not_writable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permission checks are ineffective when running as root")
	}
	dir := t.TempDir()
	nested := filepath.Join(dir, "locked")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.Chmod(nested, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(nested, 0o700) })

	path := filepath.Join(nested, "credentials.json")
	plat := &platform.Platform{CI: platform.CIInfo{Present: true}}
	store := auth.New(plat, path)

	err := store.Set(validCreds())
	assertErrCode(t, err, errs.AUT003)
}

// TestFileStore_Delete_existing verifies that Delete removes the credentials file
// and a subsequent Get returns AUT001.
func TestFileStore_Delete_existing(t *testing.T) {
	store, _ := newFileStore(t)
	if err := store.Set(validCreds()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.Get()
	assertErrCode(t, err, errs.AUT001)
}

// TestFileStore_Delete_idempotent verifies that Delete returns nil when the
// credentials file does not exist.
func TestFileStore_Delete_idempotent(t *testing.T) {
	store, _ := newFileStore(t)
	if err := store.Delete(); err != nil {
		t.Errorf("Delete on missing file should return nil, got: %v", err)
	}
}

// assertErrCode fails the test if err is not an *errs.Error with the given code.
func assertErrCode(t *testing.T, err error, code errs.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s, got nil", code)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("expected *errs.Error, got %T: %v", err, err)
	}
	if e.Code != code {
		t.Errorf("error code = %q, want %q", e.Code, code)
	}
}
