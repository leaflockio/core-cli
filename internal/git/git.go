// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package git resolves the set of files for a command to operate on using
// git-based scoping. It handles --staged (git index) and --pr (diff against
// a base ref) resolution.
package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/leaflock/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

const (
	defaultRemote     = "origin"
	defaultBranch     = "main"
	fallbackBranch    = "master"
	shallowFetchDepth = "1"
)

// Mockable runner for git commands that capture output.
var cmdOutput = func(args ...string) ([]byte, error) {
	return exec.Command("git", args...).Output()
}

// Mockable runner for git commands that discard output.
var cmdRun = func(args ...string) error {
	return exec.Command("git", args...).Run()
}

// Flags holds the git-scope flag values for commands that support --staged
// and --pr resolution. AddTo registers the flags onto a command.
type Flags struct {
	Staged     bool   // --staged    : git staged files only
	PR         bool   // --pr        : files changed in the current PR
	Base       string // --base      : base branch/ref for --pr (overrides auto-detection)
	DiffFilter string // --diff-filter: git diff-filter value (default: ACM)
}

// AddTo registers the git-scope flags onto cmd as local flags.
func (f *Flags) AddTo(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.Staged, "staged", false,
		"staged files only (git diff --cached, for pre-commit hooks)")
	cmd.Flags().BoolVar(&f.PR, "pr", false,
		"files changed in the current PR (CI-aware, for pull-request checks)")
	cmd.Flags().StringVar(&f.Base, "base", "",
		"base branch or ref for --pr (overrides auto-detection from CI environment)")
	cmd.Flags().StringVar(&f.DiffFilter, "diff-filter", "ACM",
		"git diff-filter characters controlling which changed files are included (default: ACM)")
}

// Base should be the already-resolved ref (from platform.Provider.ResolveBase).
// If the diff fails due to a shallow clone, autoFetch triggers a fetch retry.
func ResolvePRFiles(base, diffFilter string, autoFetch bool) ([]string, error) {
	files, err := gitDiffFiles(base, diffFilter)
	if err == nil {
		return files, nil
	}
	if !autoFetch {
		return nil, prFetchError(base)
	}
	return resolveWithFetch(base, diffFilter)
}

// Fetches the base ref then retries the diff.
func resolveWithFetch(base, diffFilter string) ([]string, error) {
	fetchRef := strings.TrimPrefix(base, defaultRemote+"/")
	if err := cmdRun("fetch", "--depth="+shallowFetchDepth, defaultRemote, fetchRef); err != nil {
		return resolveFallback(base, diffFilter)
	}
	files, err := gitDiffFiles(base, diffFilter)
	if err != nil {
		return nil, prFetchError(base)
	}
	return files, nil
}

// Tries the fallback branch when the primary base ref fetch fails.
func resolveFallback(base, diffFilter string) ([]string, error) {
	if base != defaultRemote+"/"+defaultBranch {
		return nil, prFetchError(base)
	}
	fallback := defaultRemote + "/" + fallbackBranch
	if files, err := gitDiffFiles(fallback, diffFilter); err == nil {
		return files, nil
	}
	if err := cmdRun("fetch", "--depth="+shallowFetchDepth, defaultRemote, fallbackBranch); err == nil {
		if files, err := gitDiffFiles(fallback, diffFilter); err == nil {
			return files, nil
		}
	}
	return nil, prFetchError(base)
}

// ResolveStaged returns the list of currently staged files.
func ResolveStaged(diffFilter string) ([]string, error) {
	out, err := cmdOutput("diff", "--cached", "--name-only", "--diff-filter="+diffFilter)
	if err != nil {
		return nil, errs.Caller(errs.LIC005,
			"could not list staged files",
			err,
			errs.Context{
				Cause:      "the working directory may not be a git repository",
				Resolution: "run from inside a git repository or use --all instead",
			},
		)
	}
	return splitLines(string(out)), nil
}

// gitDiffFiles runs git diff between base and HEAD and returns the changed files.
func gitDiffFiles(base, diffFilter string) ([]string, error) {
	out, err := cmdOutput("diff", base+"...HEAD", "--name-only", "--diff-filter="+diffFilter)
	if err != nil {
		return nil, err
	}
	return splitLines(string(out)), nil
}

// prFetchError returns a structured error explaining how to fix the --pr failure.
func prFetchError(base string) error {
	return errs.Caller(errs.LIC006,
		fmt.Sprintf("base branch %q is not available in the local git history", base),
		nil,
		errs.Context{
			Cause:      "the repository was cloned with a shallow depth that excludes the base commit",
			Resolution: fmt.Sprintf("run: git fetch %s %s", defaultRemote, strings.TrimPrefix(base, defaultRemote+"/")),
		},
		errs.Context{
			Cause:      "the base ref may not exist on the remote",
			Resolution: "use --base to specify a different base branch",
		},
	)
}

// splitLines splits newline-delimited output, trimming blanks.
func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(s), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}
