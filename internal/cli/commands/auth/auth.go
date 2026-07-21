// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth

import (
	"errors"

	"github.com/leaflockio/core-cli/internal/platform"
	"github.com/zalando/go-keyring"
)

// Store is the interface for reading and writing credentials.
type Store interface {
	// Get reads the stored credentials. Returns AUT001 if no credentials have
	// been saved yet, or AUT005 if they have expired.
	Get() (*Credentials, error)

	// Set persists creds, replacing any previously stored value.
	Set(creds *Credentials) error

	// Delete removes the stored credentials. Returns nil if they did not exist.
	Delete() error

	// Backend returns a human-readable label for the active backend
	// ("keychain" or "file").
	Backend() string
}

// New returns the most secure Store available on the current platform.
// The selection runs once and does not change for the lifetime of the process.
func New(plat *platform.Platform, credentialsPath string) Store {
	if plat.CI.Present {
		return newFileStore(credentialsPath)
	}

	switch plat.OS {
	case platform.MacOS, platform.Windows:
		return newKeychainStore()
	case platform.Linux:
		return probeKeychain(credentialsPath)
	case platform.UnknownOS:
		return newFileStore(credentialsPath)
	}
	return newFileStore(credentialsPath)
}

// probeKeychain attempts a lightweight read against the keyring daemon.
// If the daemon responds (even with ErrNotFound), KeychainStore is returned.
// Any other error means the daemon is absent and FileStore is used instead.
func probeKeychain(credentialsPath string) Store {
	_, err := keyring.Get(keychainService, keychainAccount)
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return newKeychainStore()
	}
	return newFileStore(credentialsPath)
}
