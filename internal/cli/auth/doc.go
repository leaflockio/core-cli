// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package auth manages credential storage and retrieval for the CLI.
//
// # Overview
//
// The auth package provides a platform-aware credential store. On macOS and
// Windows it uses the OS-native secure credential store without writing
// anything to disk. On Linux it probes for a running keyring daemon (GNOME
// Keyring or KWallet) and uses it when available; otherwise it falls back to
// a permission-restricted JSON file. In CI environments the file backend is
// always selected regardless of OS. The selection is automatic and transparent
// to callers — all interaction goes through the [Store] interface.
//
// # Credential model
//
// A [Credentials] value holds everything the CLI needs to authenticate:
//
//	type Credentials struct {
//	    Token     string    // Personal Access Token (PAT) — long-lived, user-generated
//	    ExpiresAt time.Time // zero value means non-expiring
//	    Workspace string    // workspace/org slug
//	}
//
// The current model uses a Personal Access Token (PAT): the user generates a
// token in the dashboard and enters it once via the login command. The token
// is stored as-is and sent with every API request. This is distinct from the
// OAuth access-token + refresh-token model:
//
//   - PAT: single long-lived token, stored directly, never auto-refreshed.
//     Revoked manually in the dashboard. Preferred for team/enterprise use
//     (auditable, fine-grained scope control, works in CI without a browser).
//
//   - OAuth access token: short-lived (~1 hour), obtained via browser/device
//     flow, auto-refreshed using a separate long-lived refresh token. The
//     refresh token is the sensitive secret; the access token can be re-derived.
//     OAuth is a future enhancement and would require expanding [Credentials]
//     with AccessToken, RefreshToken, and ExpiresAt fields.
//
// Credentials are written atomically and read back as a single unit. The store
// never exposes individual fields in isolation — always the full struct — so
// callers cannot partially overwrite a credential set.
//
// # Backend selection
//
// [New] accepts a [*platform.Platform] and returns the most secure [Store]
// available on the current platform. The selection runs once at startup and
// does not change for the lifetime of the process:
//
//	Platform                    Probe?   Backend chosen
//	──────────────────────────────────────────────────────────────────────────
//	macOS (interactive)         no       KeychainStore  → macOS Keychain
//	Windows                     no       KeychainStore  → Windows Credential Manager
//	Linux, CI detected          no       FileStore      → credentials file
//	Linux, no CI                yes      KeychainStore  if daemon responds
//	                                     FileStore      if daemon absent/errors
//
// macOS and Windows skip the probe because their credential stores are always
// available in interactive sessions. Linux requires a probe because keyring
// availability depends on whether a GNOME Keyring or KWallet daemon is running
// in the session — this is not guaranteed, especially on servers and in
// containers.
//
// The probe is a lightweight no-op read against the keyring. If it returns any
// error other than "item not found", the daemon is considered unavailable and
// [FileStore] is selected. The probe result is not cached beyond the lifetime
// of the process.
//
// CI detection uses [platform.CIInfo] already resolved at startup
// (GITHUB_ACTIONS, GITLAB_CI, TF_BUILD, CIRCLECI, etc.). No extra detection
// is needed.
//
// # Store interface
//
//	type Store interface {
//	    // Get reads the stored credentials. Returns ErrNotFound (AUT001) if
//	    // no credentials have been saved yet.
//	    Get() (*Credentials, error)
//
//	    // Set persists creds, replacing any previously stored value.
//	    Set(creds *Credentials) error
//
//	    // Delete removes the stored credentials. Returns nil if they did not
//	    // exist — callers do not need to call Get first.
//	    Delete() error
//
//	    // Backend returns a human-readable label for the active backend
//	    // ("keychain" or "file"), used in auth status output.
//	    Backend() string
//	}
//
// # Implementations
//
// [KeychainStore] — darwin, linux (with daemon), windows
//
// Stores the credentials JSON blob as a single keychain item under the service
// name config.Entity and the account name "credentials". Uses
// github.com/zalando/go-keyring (Apache 2.0) which wraps:
//   - macOS: Security framework (SecItemAdd / SecItemCopyMatching)
//   - Linux: Secret Service over D-Bus (GNOME Keyring / KWallet)
//   - Windows: Windows Credential Manager (CredWrite / CredRead)
//
// The keychain item is protected by the OS — other processes cannot read it
// without the user granting explicit permission. On macOS, the first access
// per session may prompt a system dialog; subsequent accesses in the same
// session are silent. No file is written to disk when this backend is active.
//
// [FileStore] — linux (headless/CI), fallback
//
// Reads and writes credentials at the path returned by
// [workspace.Workspace.CredentialsPath]. The file is written with 0600
// permissions (owner read/write only). The parent directory is created with
// 0700 if it does not exist.
//
// # Error codes
//
// All errors are [errs.Error] values in the AUT domain:
//
//	AUT001  Credentials not found    — no credentials stored yet
//	AUT002  Credential read failed   — backend I/O error or keychain permission denied
//	AUT003  Credential write failed  — backend I/O error or keychain permission denied
//	AUT004  Credential delete failed — same causes as AUT003
//	AUT005  Credentials expired      — token ExpiresAt is in the past
//
// # Login command
//
// The login command is the sole entry point that populates the store. It is
// the first command a new user runs and the recovery path when credentials
// expire. Its flow:
//
//  1. Prompt the user for their PAT (input is masked, never echoed).
//  2. Validate the token against the API (single authenticated GET — no
//     browser redirect).
//  3. Call Store.Set with the validated credentials.
//  4. Print which backend was used: "Credentials saved to keychain." or
//     "Credentials saved to file."
//
// Every other command that requires auth calls Store.Get inside its
// PersistentPreRunE. If Get returns AUT001 or AUT005 the command exits with a
// consistent error — it never re-prompts inline.
//
// # Command surface
//
//	auth           Parent command — prints status summary
//	auth login     Prompt → validate → store credentials
//	auth logout    Delete stored credentials and confirm
//	auth status    Print active backend, workspace, token expiry
//
// # File layout
//
//	internal/auth/
//	├── doc.go           This file — package documentation
//	├── auth.go          Store interface, New() factory, backend selection
//	├── credentials.go   Credentials struct, JSON marshal/unmarshal, expiry check
//	├── keychain.go      KeychainStore (build constraint: darwin || linux || windows)
//	├── filestore.go     FileStore, permission enforcement
//	└── errors.go        AUT error code definitions
//
//	cmd/auth/
//	├── auth.go          auth parent command
//	├── login.go         auth login
//	├── logout.go        auth logout
//	└── status.go        auth status
//
// # Usage example
//
//	creds, err := a.Auth.Get()
//	if err != nil {
//	    return err
//	}
//	client := api.NewClient(creds.Token)
package auth
