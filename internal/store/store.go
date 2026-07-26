// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package store reads and writes structured config files by base name,
// resolving whichever discoverable extension is present on disk. Values
// must use mapstructure tags — see internal/util/codec.
package store

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/leaflockio/core-cli/internal/util/codec"
	"github.com/leaflockio/core-cli/internal/util/fsutil"
)

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

// Load finds dir/name.<ext> for the first matching extension in
// discoverableExtensions and decodes it into v. Returns found=false (not an
// error) if no matching file exists.
//
// If more than one extension variant exists for the same name, the first
// match in discoverableExtensions order wins silently.
func Load(dir, name string, v any) (bool, error) {
	for _, ext := range discoverableExtensions() {
		data, err := os.ReadFile(filepath.Join(dir, name+"."+ext))
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

// Save encodes v and writes it to dir/name.<ext>, creating dir if it does
// not exist. The dirPerm and filePerm arguments are the caller's
// scope-appropriate permissions.
//
// If a file already exists for any discoverable extension, Save writes back
// using that same extension, preserving the original format choice.
// Otherwise it writes dir/name.<defaultExt>.
func Save(dir, name string, v any, dirPerm, filePerm fs.FileMode) error {
	ext := existingExtension(dir, name)
	if ext == "" {
		ext = defaultExt
	}
	data, err := codec.Marshal(v, formatFor(ext))
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+"."+ext)
	return fsutil.WriteFile(path, data, dirPerm, filePerm)
}

// existingExtension returns the extension of the first discoverable file
// matching dir/name, or an empty string if none exists.
func existingExtension(dir, name string) string {
	for _, ext := range discoverableExtensions() {
		if _, err := os.Stat(filepath.Join(dir, name+"."+ext)); err == nil {
			return ext
		}
	}
	return ""
}

// ExistingExtensions returns every discoverable extension for which
// dir/name.<ext> exists, in discoverableExtensions order.
func ExistingExtensions(dir, name string) []string {
	var found []string
	for _, ext := range discoverableExtensions() {
		if _, err := os.Stat(filepath.Join(dir, name+"."+ext)); err == nil {
			found = append(found, ext)
		}
	}
	return found
}
