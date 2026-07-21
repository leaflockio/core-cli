// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package fsutil provides low-level filesystem helpers used across the CLI.
package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// EnsureDir creates the directory at path and all necessary parents,
// returning path.
//
//	dir, _ := EnsureDir("<root>/license/templates", 0o700)
//	// dir == "<root>/license/templates" ← a directory, ready to use
func EnsureDir(path string, perm fs.FileMode) (string, error) {
	if err := os.MkdirAll(path, perm); err != nil {
		return "", fmt.Errorf("fsutil: create directory %s: %w", path, err)
	}
	return path, nil
}

// EnsureParent creates the parent directory of path, returning path unchanged.
// The file itself is not created.
//
//	f, _ := EnsureParent("<root>/license/license.lock", 0o700)
//	// f == "<root>/license/license.lock" ← a file path
//	// "<root>/license/" now exists
func EnsureParent(path string, perm fs.FileMode) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), perm); err != nil {
		return "", fmt.Errorf("fsutil: create parent for %s: %w", path, err)
	}
	return path, nil
}

// WriteFile ensures the parent directory of path exists with dirPerm, then
// writes data to path with filePerm, creating or truncating the file.
//
//	err := WriteFile("<root>/license/license.lock", data, 0o700, 0o600)
//	// "<root>/license/" is created with 0o700 if missing
//	// "<root>/license/license.lock" is written with 0o600
func WriteFile(path string, data []byte, dirPerm, filePerm fs.FileMode) error {
	if _, err := EnsureParent(path, dirPerm); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("fsutil: write file %s: %w", path, err)
	}
	return nil
}

// CreateFile ensures the parent directory of path exists with dirPerm, then
// opens path for writing with filePerm, creating it if it does not exist.
// The caller is responsible for closing the returned file.
//
//	f, err := CreateFile("<root>/license/license.lock", 0o700, 0o600)
//	// "<root>/license/" is created with 0o700 if missing
//	// f is an open file handle at "<root>/license/license.lock"
func CreateFile(path string, dirPerm, filePerm fs.FileMode) (*os.File, error) {
	if _, err := EnsureParent(path, dirPerm); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return nil, fmt.Errorf("fsutil: create file %s: %w", path, err)
	}
	return f, nil
}
