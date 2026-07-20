// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth

import (
	"errors"
	"os"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/util/codec"
	"github.com/leaflockio/core-cli/internal/util/fsutil"
)

const (
	backendFile = "file"
	dirPerm     = 0o700
	filePerm    = 0o600
)

// fileStore stores credentials in a permission-restricted JSON file.
type fileStore struct {
	path string
}

func newFileStore(path string) Store {
	return &fileStore{path: path}
}

func (s *fileStore) Backend() string { return backendFile }

func (s *fileStore) Get() (*Credentials, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errs.Caller(
			errs.AUT001,
			"no credentials stored",
			err,
			errs.Context{
				Cause:      "credentials file does not exist",
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
				Cause:      "credentials file could not be read",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	var creds Credentials
	if err := codec.Unmarshal(data, codec.JSON, &creds); err != nil {
		return nil, errs.Caller(
			errs.AUT002,
			"credential read failed",
			err,
			errs.Context{
				Cause:      "credentials file is corrupt or in an unrecognized format",
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

func (s *fileStore) Set(creds *Credentials) error {
	data, err := codec.Marshal(creds, codec.JSON)
	if err != nil {
		return errs.Unexpected(err, errs.Context{Cause: "credentials could not be serialized"})
	}
	if err := fsutil.WriteFile(s.path, data, dirPerm, filePerm); err != nil {
		return errs.Caller(
			errs.AUT003,
			"credential write failed",
			err,
			errs.Context{
				Cause:      "credentials file could not be written",
				Resolution: "run 'leaf auth login' to re-authenticate",
			},
		)
	}
	return nil
}

func (s *fileStore) Delete() error {
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errs.Caller(
			errs.AUT004,
			"credential delete failed",
			err,
			errs.Context{
				Cause:      "credentials file could not be removed",
				Resolution: "run 'leaf auth logout' and try again",
			},
		)
	}
	return nil
}
