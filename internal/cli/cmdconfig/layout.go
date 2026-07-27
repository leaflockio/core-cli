// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/store"
)

// LayoutMode describes how a directory of per-command config files is
// organized.
type LayoutMode int

const (
	// LayoutNone means the directory has neither a manifest nor any
	// per-command file.
	LayoutNone LayoutMode = iota

	// LayoutFlat means the directory has a single manifest file holding
	// every command's section.
	LayoutFlat

	// LayoutModular means the directory has one file per command, each
	// holding that command's own section.
	LayoutModular
)

// Layout is the classified result of scanning a directory of per-command
// config files.
type Layout struct {
	// Mode is the active layout, or LayoutNone if the directory has no
	// manifest and no per-command file.
	Mode LayoutMode

	// ExtensionConflicts maps a command name to the discoverable extensions
	// found for it, set only when that command's own file exists under more
	// than one extension.
	ExtensionConflicts map[string][]string

	// Warnings lists every non-blocking issue found: a file that doesn't
	// match the manifest name or any known command name, a file that
	// matches a known command name but that command doesn't consume config,
	// and any file whose extension isn't one of the discoverable ones.
	// Warnings never block — callers should point the user at a diagnostic
	// command for the detail rather than surfacing these inline.
	Warnings []string
}

// DetectLayout scans dir and classifies its contents against allCommands
// (every top-level command this tool defines) and configCommands (the
// subset of allCommands that actually declare a config section).
//
// A dir that does not exist is reported as LayoutNone, not an error.
// DetectLayout does not create dir or any file inside it.
//
// DetectLayout returns an error for two conditions only: the manifest file
// found under more than one discoverable extension, and the manifest
// coexisting with any per-command file.
func DetectLayout(dir string, allCommands, configCommands []string) (*Layout, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Layout{ExtensionConflicts: map[string][]string{}}, nil
		}
		return nil, errDirUnreadable(dir, err)
	}

	extsByBase := groupByBase(entries)
	allSet := toSet(allCommands)
	configSet := toSet(configCommands)

	layout := &Layout{ExtensionConflicts: map[string][]string{}}
	manifestPresent := false
	recognizedCount := 0

	for _, base := range sortedKeys(extsByBase) {
		discoverable := store.ExistingExtensions(filepath.Join(dir, base))
		stray := notIn(extsByBase[base], discoverable)

		switch {
		case base == config.ManifestFile:
			present, err := classifyManifest(base, discoverable, stray, layout)
			if err != nil {
				return nil, err
			}
			manifestPresent = present

		case allSet[base]:
			if classifyCommand(base, discoverable, stray, configSet[base], layout) {
				recognizedCount++
			}

		default:
			classifyUnrecognized(base, discoverable, stray, layout)
		}
	}

	if manifestPresent && recognizedCount > 0 {
		return nil, errManifestConflict()
	}

	layout.Mode = resolveMode(manifestPresent, recognizedCount)
	return layout, nil
}

// groupByBase groups every regular file in entries by base name — the
// filename without its extension. Directories are skipped.
func groupByBase(entries []os.DirEntry) map[string][]string {
	extsByBase := map[string][]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.TrimPrefix(filepath.Ext(entry.Name()), ".")
		base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		extsByBase[base] = append(extsByBase[base], ext)
	}
	return extsByBase
}

// sortedKeys returns m's keys in sorted order, for deterministic iteration.
func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// classifyManifest updates layout for the manifest base name and reports
// whether the manifest is present — exactly one discoverable extension.
// Returns an error when it exists under more than one.
func classifyManifest(base string, discoverable, stray []string, layout *Layout) (bool, error) {
	if len(discoverable) > 1 {
		return false, errManifestAmbiguous(base, discoverable)
	}
	layout.Warnings = append(layout.Warnings, strayWarnings(base, stray, "the manifest")...)
	return len(discoverable) == 1, nil
}

// classifyCommand updates layout for a base name that matches a known
// command and reports whether that command is recognized without
// ambiguity — exactly one discoverable extension.
func classifyCommand(base string, discoverable, stray []string, hasConfig bool, layout *Layout) bool {
	switch {
	case len(discoverable) > 1:
		layout.ExtensionConflicts[base] = discoverable
	case len(discoverable) == 1 && !hasConfig:
		layout.Warnings = append(layout.Warnings,
			fmt.Sprintf("%s: recognized command but does not read a config section", base))
	}
	layout.Warnings = append(layout.Warnings, strayWarnings(base, stray, base)...)
	return len(discoverable) == 1
}

// classifyUnrecognized appends warnings for a base name that matches
// neither the manifest nor any known command.
func classifyUnrecognized(base string, discoverable, stray []string, layout *Layout) {
	if len(discoverable) > 0 {
		layout.Warnings = append(layout.Warnings, unrecognizedWarning(base, discoverable))
	}
	layout.Warnings = append(layout.Warnings, strayWarnings(base, stray, "")...)
}

// resolveMode determines the active LayoutMode from what DetectLayout found.
func resolveMode(manifestPresent bool, recognizedCount int) LayoutMode {
	switch {
	case manifestPresent:
		return LayoutFlat
	case recognizedCount > 0:
		return LayoutModular
	default:
		return LayoutNone
	}
}

// errDirUnreadable reports that dir exists but could not be read.
func errDirUnreadable(dir string, err error) error {
	return errs.Caller(
		errs.CCF003,
		"config directory could not be read",
		err,
		errs.Context{
			Cause:      fmt.Sprintf("%q could not be read", dir),
			Resolution: "check the directory's permissions",
		},
	)
}

// errManifestAmbiguous reports that the manifest file exists under more
// than one discoverable extension.
func errManifestAmbiguous(base string, discoverable []string) error {
	return errs.Caller(
		errs.CCF001,
		"manifest file exists with more than one extension",
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("found %s", joinFiles(base, discoverable)),
			Resolution: "keep only one manifest file and remove the others",
		},
	)
}

// errManifestConflict reports that the manifest file coexists with one or
// more per-command config files.
func errManifestConflict() error {
	return errs.Caller(
		errs.CCF002,
		"manifest file coexists with one or more per-command config files",
		nil,
		errs.Context{
			Cause:      "a config directory can be a flat manifest or one file per command, never both at once",
			Resolution: "delete the manifest file, or delete the per-command files — not both at once",
		},
	)
}

// unrecognizedWarning reports a base name that matches neither the manifest
// nor any known command, for each discoverable extension found.
func unrecognizedWarning(base string, discoverable []string) string {
	if len(discoverable) > 1 {
		return fmt.Sprintf("%s: unrecognized file (found %s)", base, joinFiles(base, discoverable))
	}
	return fmt.Sprintf("%s.%s: unrecognized file", base, discoverable[0])
}

// joinFiles renders base+ext for each ext in exts, comma-separated (e.g.
// "manifest.yaml, manifest.json"), for use in messages listing every
// variant found for the same base name.
func joinFiles(base string, exts []string) string {
	files := make([]string, len(exts))
	for i, ext := range exts {
		files[i] = base + "." + ext
	}
	return strings.Join(files, ", ")
}

// strayWarnings reports every extension found for base that store doesn't
// manage. Owner names what base actually is ("the manifest", a command name,
// or "" when base isn't recognized at all) so the message reads naturally
// either way.
func strayWarnings(base string, stray []string, owner string) []string {
	warnings := make([]string, 0, len(stray))
	for _, ext := range stray {
		name := base
		if ext != "" {
			name = base + "." + ext
		}
		if owner == "" {
			warnings = append(warnings, fmt.Sprintf("%s: unrecognized file", name))
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s: not a supported config extension for %s", name, owner))
	}
	return warnings
}

// notIn returns every element of all that isn't also in exclude.
func notIn(all, exclude []string) []string {
	var out []string
	for _, v := range all {
		if !slices.Contains(exclude, v) {
			out = append(out, v)
		}
	}
	return out
}

// toSet converts ss to a set for O(1) membership checks.
func toSet(ss []string) map[string]bool {
	set := make(map[string]bool, len(ss))
	for _, s := range ss {
		set[s] = true
	}
	return set
}
