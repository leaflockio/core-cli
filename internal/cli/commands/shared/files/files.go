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
// special-cased to match all in the current working directory.
func (f *Files) Resolve(root string, args []string) ([]string, error) {
	var universe []string
	var err error

	switch {
	case f.PR:
		universe, err = git.ResolvePRFiles(root, f.Base, f.DiffFilter)
	case f.Staged:
		universe, err = git.ResolveStaged(root, f.DiffFilter)
	case f.NoGitignore:
		universe, err = fstree.Walk(root)
	default:
		universe, err = git.ListFiles(root)
	}
	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		argPatterns := make([]string, len(args))
		for i, a := range args {
			if a == "." {
				a = fstree.MatchAll
			}
			argPatterns[i] = a
		}
		universe = fstree.Filter(universe, argPatterns, nil)
	}

	include := f.include
	if len(include) == 0 {
		include = []string{fstree.MatchAll}
	}
	return fstree.Filter(universe, include, f.exclude), nil
}
