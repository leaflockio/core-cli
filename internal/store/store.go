// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package store reads and writes structured config files by base path,
// resolving whichever discoverable extension is present on disk. Values
// must use mapstructure tags — see internal/util/codec.
package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/util/codec"
	"github.com/leaflockio/core-cli/internal/util/fsutil"
)

var errBaseHasExtension = errors.New("store: base must not include an extension")

const (
	extYAML  = "yaml"
	extYML   = "yml"
	extJSON  = "json"
	extJSONC = "jsonc"

	defaultExt = extYAML
)

// discoverableExtensions returns the file extensions Load checks, in order.
func discoverableExtensions() []string {
	return []string{extYAML, extYML, extJSON, extJSONC}
}

// formatFor returns the codec.Format for a discoverable extension.
func formatFor(ext string) codec.Format {
	switch ext {
	case extJSON:
		return codec.JSON
	case extJSONC:
		return codec.JSONC
	case extYML:
		return codec.YAML
	default:
		return codec.YAML
	}
}

// validateBase returns an error if base already carries an extension. Every
// exported function in this package takes an extension-less base path and
// appends a discoverable extension itself.
func validateBase(base string) error {
	if ext := filepath.Ext(base); ext != "" {
		return errs.Unexpected(fmt.Errorf("%w: got %q", errBaseHasExtension, base))
	}
	return nil
}

// Load finds base.<ext> for the first matching extension in
// discoverableExtensions and decodes it into v. Returns found=false (not an
// error) if no matching file exists.
//
// If more than one extension variant exists for the same base, the first
// match in discoverableExtensions order wins silently.
func Load(base string, v any) (bool, error) {
	if err := validateBase(base); err != nil {
		return false, err
	}
	for _, ext := range discoverableExtensions() {
		data, err := os.ReadFile(base + "." + ext)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, err
		}
		if err := codec.Unmarshal(data, formatFor(ext), v); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Save encodes v and writes it to base.<ext>, creating base's parent
// directory if it does not exist. The dirPerm and filePerm arguments are the
// caller's scope-appropriate permissions.
//
// If a file already exists for any discoverable extension, Save writes back
// using that same extension, preserving the original format choice.
// Otherwise it writes base.<defaultExt>.
func Save(base string, v any, dirPerm, filePerm fs.FileMode) error {
	if err := validateBase(base); err != nil {
		return err
	}
	ext := ResolvedExt(base)
	data, err := codec.Marshal(v, formatFor(ext))
	if err != nil {
		return err
	}
	return fsutil.WriteFile(base+"."+ext, data, dirPerm, filePerm)
}

// existingExtension returns the extension of the first discoverable file
// matching base, or an empty string if none exists.
func existingExtension(base string) string {
	for _, ext := range discoverableExtensions() {
		if _, err := os.Stat(base + "." + ext); err == nil {
			return ext
		}
	}
	return ""
}

// ResolvedExt returns the discoverable extension currently backing base, or
// the default extension when no matching file exists yet.
//
// If more than one extension variant exists for the same base, the first
// match in discoverableExtensions order wins.
func ResolvedExt(base string) string {
	if ext := existingExtension(base); ext != "" {
		return ext
	}
	return defaultExt
}

// ExistingExtensions returns every discoverable extension for which
// base.<ext> exists, in discoverableExtensions order.
func ExistingExtensions(base string) []string {
	var found []string
	for _, ext := range discoverableExtensions() {
		if _, err := os.Stat(base + "." + ext); err == nil {
			found = append(found, ext)
		}
	}
	return found
}
