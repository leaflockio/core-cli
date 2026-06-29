// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth

import (
	"errors"

	"github.com/leaflock/core-cli/internal/config"
	"github.com/leaflock/core-cli/internal/errs"
	"github.com/leaflock/core-cli/internal/util/codec"
	"github.com/zalando/go-keyring"
)

const (
	keychainService = config.Entity
	keychainAccount = "credentials"
	backendKeychain = "keychain"
)

// keychainStore stores credentials in the OS-native credential store.
// On macOS this is the Keychain; on Linux it is the Secret Service over D-Bus
// (GNOME Keyring or KWallet); on Windows it is the Credential Manager.
type keychainStore struct{}

func newKeychainStore() Store {
	return &keychainStore{}
}

func (s *keychainStore) Backend() string { return backendKeychain }

func (s *keychainStore) Get() (*Credentials, error) {
	raw, err := keyring.Get(keychainService, keychainAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, errs.Caller(
			errs.AUT001,
			"no credentials stored",
			err,
			errs.Context{
				Cause:      "no credentials entry found in the keychain",
				Resolution: "run 'leaf auth login' to authenticate",
			},
		)
	}
	if err != nil {
		return nil, errs.Caller(
			errs.AUT002,
			"credential read failed",
			err,
			errs.Context{
				Cause:      "keychain access error",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	var creds Credentials
	if err := codec.Unmarshal([]byte(raw), codec.JSON, &creds); err != nil {
		return nil, errs.Caller(
			errs.AUT002,
			"credential read failed",
			err,
			errs.Context{
				Cause:      "stored credentials could not be parsed",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	if creds.IsExpired() {
		return nil, errs.Caller(
			errs.AUT005,
			"credentials have expired",
			nil,
			errs.Context{
				Cause:      "stored credentials have passed their expiry date",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	return &creds, nil
}

func (s *keychainStore) Set(creds *Credentials) error {
	data, err := codec.Marshal(creds, codec.JSON)
	if err != nil {
		return errs.Unexpected(err, errs.Context{Cause: "credentials could not be serialized"})
	}
	if err := keyring.Set(keychainService, keychainAccount, string(data)); err != nil {
		return errs.Caller(
			errs.AUT003,
			"credential write failed",
			err,
			errs.Context{
				Cause:      "keychain access error",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	return nil
}

func (s *keychainStore) Delete() error {
	err := keyring.Delete(keychainService, keychainAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return errs.Caller(
			errs.AUT004,
			"credential delete failed",
			err,
			errs.Context{
				Cause:      "keychain access error",
				Resolution: "run 'leaf auth logout' and try again",
			},
		)
	}
	return nil
}
