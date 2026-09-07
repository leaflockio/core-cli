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

// CommandSource describes where a single command's config section lives.
type CommandSource int

const (
	// SourceNone means neither a dedicated file nor a manifest section
	// configures this command.
	SourceNone CommandSource = iota

	// SourceFile means the command has its own dedicated config file.
	SourceFile

	// SourceManifest means the command's section lives inline in the
	// manifest.
	SourceManifest
)

// Sources is the fully resolved result of scanning a directory of
// per-command config files together with the manifest's own content.
type Sources struct {
	// Manifest is the decoded manifest, set whenever a manifest file was
	// found.
	Manifest *Manifest

	// ManifestCommands holds the manifest's own per-command sections,
	// keyed by command name. Set whenever Manifest is set, nil otherwise.
	ManifestCommands map[string]any

	// Commands maps every config-consuming command's name to exactly one
	// of SourceFile, SourceManifest, or SourceNone. A command found in
	// both its own file and a manifest section, or whose own file exists
	// under more than one extension, is not assigned any of these — it's
	// recorded in ManifestConflicts or ExtensionConflicts instead.
	Commands map[string]CommandSource

	// ExtensionConflicts maps a command name to the discoverable extensions
	// found for it, set only when that command's own file exists under more
	// than one extension.
	ExtensionConflicts map[string][]string

	// ManifestConflicts lists, in sorted order, every command name found
	// with both its own dedicated file and a manifest section.
	ManifestConflicts []string

	// Warnings lists every non-blocking issue found: a file that doesn't
	// match the manifest name or any known command name, a file that
	// matches a known command name but that command doesn't consume config,
	// and any file whose extension isn't one of the discoverable ones.
	Warnings []string
}

// ResolveConfig scans dir, reads and decodes the manifest if one exists,
// and resolves every name in configCommands (the commands that actually
// declare a config section) to exactly one CommandSource — SourceFile,
// SourceManifest, or SourceNone — in a single pass, using allCommands
// (every top-level command this tool defines) only to recognize a file
// that matches a command name that doesn't consume config, which is
// worth a warning rather than a resolved source.
//
// A dir that does not exist is reported as an empty Sources, not an
// error.
//
// ResolveConfig returns an error when: the manifest file exists under
// more than one discoverable extension, or exists but can't be read or
// decoded. A manifest coexisting with per-command files is not itself an
// error — only the same command being defined in both is, recorded in
// Sources.ManifestConflicts rather than raised here.
func ResolveConfig(dir string, allCommands, configCommands []string) (*Sources, error) {
	sources := &Sources{
		Commands:           seedSources(configCommands),
		ExtensionConflicts: map[string][]string{},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return sources, nil
		}
		return nil, errDirUnreadable(dir, err)
	}

	extsByBase := groupByBase(entries)
	allSet := toSet(allCommands)
	configSet := toSet(configCommands)

	for _, base := range sortedKeys(extsByBase) {
		discoverable := store.ExistingExtensions(filepath.Join(dir, base))
		stray := notIn(extsByBase[base], discoverable)

		switch {
		case base == config.ManifestFile:
			if err := classifyManifest(dir, base, discoverable, stray, allCommands, sources); err != nil {
				return nil, err
			}

		case allSet[base]:
			classifyCommand(base, discoverable, stray, configSet[base], sources)

		default:
			classifyUnrecognized(base, discoverable, stray, sources)
		}
	}

	sources.resolveCommandSources()
	return sources, nil
}

// resolveCommandSources finalizes s.Commands using s.ManifestCommands. A
// command found in both its own file and a manifest section is recorded
// in ManifestConflicts instead of being assigned either source, mirroring
// ExtensionConflicts.
func (s *Sources) resolveCommandSources() {
	if s.ManifestCommands == nil {
		return
	}
	var conflicts []string
	for name, source := range s.Commands {
		if _, inManifest := s.ManifestCommands[name]; !inManifest {
			continue
		}
		switch source {
		case SourceFile:
			conflicts = append(conflicts, name)
			s.Commands[name] = SourceNone
		case SourceNone:
			s.Commands[name] = SourceManifest
		case SourceManifest:
		}
	}
	sort.Strings(conflicts)
	s.ManifestConflicts = conflicts
}

// seedSources returns a Commands map with every name in allCommands set
// to SourceNone, so a lookup is always defined even for a command whose
// file was never found.
func seedSources(allCommands []string) map[string]CommandSource {
	sources := make(map[string]CommandSource, len(allCommands))
	for _, name := range allCommands {
		sources[name] = SourceNone
	}
	return sources
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

// classifyManifest updates sources for the manifest base name and reports
// whether it's present — exactly one discoverable extension — and, if
// so, reads and decodes it into sources.Manifest and
// sources.ManifestCommands. Returns an error when it exists under more
// than one discoverable extension, or exists but can't be read or
// decoded.
func classifyManifest(dir, base string, discoverable, stray, allCommands []string, sources *Sources) error {
	if len(discoverable) > 1 {
		return errManifestAmbiguous(base, discoverable)
	}
	sources.Warnings = append(sources.Warnings, strayWarnings(base, stray, "the manifest")...)
	if len(discoverable) != 1 {
		return nil
	}

	var raw map[string]any
	if _, err := store.Load(filepath.Join(dir, base), &raw); err != nil {
		return errManifestUnreadable(err)
	}

	manifest, commands, err := LoadManifest(raw, allCommands)
	if err != nil {
		return err
	}
	sources.Manifest = manifest
	sources.ManifestCommands = commands
	return nil
}

// classifyCommand updates sources for a base name that matches a known
// command. Extension conflicts and a SourceFile assignment only apply
// when the command actually declares a config section (hasConfig) — an
// ambiguous or stray file for a command that never reads config has
// nothing to resolve, so it's only ever worth a warning.
func classifyCommand(base string, discoverable, stray []string, hasConfig bool, sources *Sources) {
	switch {
	case !hasConfig:
		sources.Warnings = append(sources.Warnings,
			fmt.Sprintf("%s: recognized command but does not read a config section", base))
	case len(discoverable) > 1:
		sources.ExtensionConflicts[base] = discoverable
	case len(discoverable) == 1:
		sources.Commands[base] = SourceFile
	}
	sources.Warnings = append(sources.Warnings, strayWarnings(base, stray, base)...)
}

// classifyUnrecognized appends warnings for a base name that matches
// neither the manifest nor any known command.
func classifyUnrecognized(base string, discoverable, stray []string, sources *Sources) {
	if len(discoverable) > 0 {
		sources.Warnings = append(sources.Warnings, unrecognizedWarning(base, discoverable))
	}
	sources.Warnings = append(sources.Warnings, strayWarnings(base, stray, "")...)
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

// errManifestUnreadable reports that the manifest file exists but could
// not be read or decoded.
func errManifestUnreadable(err error) error {
	return errs.Caller(
		errs.CCF005,
		"config file could not be read",
		err,
		errs.Context{
			Cause:      "the manifest exists but could not be read or decoded",
			Resolution: "check the file's syntax and permissions",
		},
	)
}

// ErrManifestConflict reports that name is defined both in the manifest
// and in its own dedicated config file — ambiguous which one applies.
func ErrManifestConflict(name string) error {
	return errs.Caller(
		errs.CCF002,
		fmt.Sprintf("%s is defined in both the manifest and its own config file", name),
		nil,
		errs.Context{
			Cause: fmt.Sprintf(
				"%s has a section in the manifest and its own dedicated config file — ambiguous which one applies",
				name,
			),
			Resolution: fmt.Sprintf(
				"remove %s's manifest section, or delete its dedicated config file — not both",
				name,
			),
		},
	)
}

// ErrExtensionConflict reports that name's own config file exists under
// more than one discoverable extension — ambiguous which one applies.
func ErrExtensionConflict(name string, extensions []string) error {
	return errs.Caller(
		errs.CCF004,
		"config file exists under more than one extension",
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("%s has files with extensions: %s", name, strings.Join(extensions, ", ")),
			Resolution: "keep only one and remove the others",
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
