// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package files

import (
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/fstree"
	"github.com/leaflockio/core-cli/internal/git"
)

type Files struct {
	Staged      bool
	PR          bool
	Base        string
	DiffFilter  string
	NoGitignore bool

	include []string
	exclude []string
}

// Validate reports an error unless exactly one file source is given, and,
// when --staged/--pr is used, isGit is true.
func (f *Files) Validate(args []string, isGit bool) error {
	hasArgs := len(args) > 0
	if hasArgs && (f.Staged || f.PR) {
		return errs.Caller(errs.SRC001,
			"positional file arguments and --staged/--pr are mutually exclusive",
			nil,
			errs.Context{
				Cause:      "explicit file arguments were given alongside --staged or --pr",
				Resolution: "use either explicit file arguments or --staged/--pr, not both",
			},
		)
	} else if !hasArgs && !f.Staged && !f.PR {
		return errs.Caller(errs.SRC002,
			"no file source specified",
			nil,
			errs.Context{
				Cause:      "none of --staged, --pr, or explicit file arguments were given",
				Resolution: `use --staged, --pr, or pass file paths directly (use "." for everything)`,
			},
		)
	}
	if (f.Staged || f.PR) && !isGit {
		return errs.Caller(errs.GIT001,
			"could not resolve files",
			nil,
			errs.Context{
				Cause:      "--staged/--pr require a git repository, but the working directory isn't one",
				Resolution: "run from inside a git repository, or pass file paths directly",
			},
		)
	}
	return nil
}

// WithInclude appends each given pattern source to the include list.
func (f *Files) WithInclude(patterns ...[]string) *Files {
	for _, p := range patterns {
		f.include = append(f.include, p...)
	}
	return f
}

// WithExclude appends each given pattern source to the exclude list.
func (f *Files) WithExclude(patterns ...[]string) *Files {
	for _, p := range patterns {
		f.exclude = append(f.exclude, p...)
	}
	return f
}

// Resolve returns the files this invocation targets, using root as the
// repository root and args as positional file/glob arguments — "." is
// special-cased to match all in the current working directory. The
// accompanying Stats reports how many files survived each narrowing stage.
func (f *Files) Resolve(root string, args []string) ([]string, Stats, error) {
	discovered, err := f.discover(root)
	if err != nil {
		return nil, Stats{}, err
	}
	stats := Stats{Discovered: len(discovered)}

	query := discovered
	if len(args) > 0 {
		query = fstree.Include(discovered, argPatterns(args))
	}
	stats.Query = len(query)

	included := fstree.Include(query, f.includePatterns())
	stats.Included = len(included)

	final := fstree.Exclude(included, f.exclude)
	stats.Final = len(final)
	stats.Excluded = stats.Included - stats.Final

	return final, stats, nil
}

// discover selects the source mechanism — a PR diff, staged changes, a
// gitignore-blind filesystem walk, or a plain git listing — and returns its
// raw file list, before any narrowing.
func (f *Files) discover(root string) ([]string, error) {
	switch {
	case f.PR:
		return git.ResolvePRFiles(root, f.Base, f.DiffFilter)
	case f.Staged:
		return git.ResolveStaged(root, f.DiffFilter)
	case f.NoGitignore:
		return fstree.Walk(root)
	default:
		return git.ListFiles(root)
	}
}

// argPatterns translates positional file/glob arguments into fstree glob
// patterns, special-casing "." to match everything.
func argPatterns(args []string) []string {
	patterns := make([]string, len(args))
	for i, a := range args {
		if a == "." {
			a = fstree.MatchAll
		}
		patterns[i] = a
	}
	return patterns
}

// includePatterns returns f's include patterns, defaulting to everything
// when none were given.
func (f *Files) includePatterns() []string {
	if len(f.include) == 0 {
		return []string{fstree.MatchAll}
	}
	return f.include
}
