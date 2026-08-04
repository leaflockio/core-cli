// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

// ResolvePRFiles resolves the set of files changed relative to base within
// the repository at dir.
func ResolvePRFiles(dir, base, diffFilter string) ([]string, error) {
	if base == "" {
		return nil, noBaseError()
	}

	files, err := gitDiffFiles(dir, base, diffFilter)
	if err != nil {
		return nil, prFetchError(base)
	}
	return files, nil
}

// refOn builds a "remote/branch" ref string.
func refOn(remote, branch string) string {
	return remote + "/" + branch
}

// BaseRefFrom builds a base ref from remote and branch.
func BaseRefFrom(remote, branch string) (string, error) {
	if remote == "" || branch == "" {
		return "", noBaseError()
	}
	return refOn(remote, branch), nil
}

// noBaseError explains that there's no ref to diff against: nothing was
// given to build one from.
func noBaseError() error {
	return errs.Caller(errs.GIT003,
		"no base to diff against",
		nil,
		errs.Context{
			Cause:      "no base ref, and no remote/branch to build one from, was given",
			Resolution: "provide an explicit base ref, or a remote and branch to build one from",
		},
	)
}

// remoteOf returns the remote name prefix of ref, e.g. "upstream" from
// "upstream/release", or "" when ref has no such prefix.
func remoteOf(ref string) string {
	if remote, _, ok := strings.Cut(ref, "/"); ok {
		return remote
	}
	return ""
}

// ResolveStaged returns the list of currently staged files in the
// repository at dir.
func ResolveStaged(dir, diffFilter string) ([]string, error) {
	out, err := runOutput(dir, "diff", "--cached", "--name-only", "--diff-filter="+diffFilter)
	if err != nil {
		return nil, errs.Caller(errs.GIT001,
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

// gitDiffFiles runs git diff between base and HEAD, in dir, and returns the
// changed files.
func gitDiffFiles(dir, base, diffFilter string) ([]string, error) {
	out, err := runOutput(dir, "diff", base+"...HEAD", "--name-only", "--diff-filter="+diffFilter)
	if err != nil {
		return nil, err
	}
	return splitLines(string(out)), nil
}

// prFetchError returns a structured error explaining why base couldn't be
// diffed against.
func prFetchError(base string) error {
	var contexts []errs.Context
	if remote := remoteOf(base); remote != "" {
		contexts = append(contexts, errs.Context{
			Cause:      "the repository was cloned with a shallow depth that excludes the base commit",
			Resolution: fmt.Sprintf("run: git fetch %s %s", remote, strings.TrimPrefix(base, remote+"/")),
		})
	}
	contexts = append(contexts, errs.Context{
		Cause:      "the base ref may not exist, or the repository doesn't have enough history to reach it",
		Resolution: "provide a different base ref",
	})
	return errs.Caller(errs.GIT002,
		fmt.Sprintf("base branch %q is not available in the local git history", base),
		nil,
		contexts...,
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
