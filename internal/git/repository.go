// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"path/filepath"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

// git ls-files arguments, for gitignore-aware file enumeration.
const (
	argLsCached          = "--cached"
	argLsOthers          = "--others"
	argLsExcludeStandard = "--exclude-standard"
)

const StrTrue = "true"

// RepoRoot resolves the absolute path to the top level of the git repository
// containing dir. Returns an error if dir is not inside a git repository.
func RepoRoot(dir string) (string, error) {
	out, err := runOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteURL resolves the URL configured for remote inside the repository
// at dir.
func RemoteURL(dir, remote string) (string, error) {
	out, err := runOutput(dir, "remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ListRemotes returns the names of every remote configured in the
// repository at dir.
func ListRemotes(dir string) ([]string, error) {
	out, err := runOutput(dir, "remote")
	if err != nil {
		return nil, err
	}
	return splitLines(string(out)), nil
}

// ListFiles enumerates every tracked and untracked-but-not-ignored file in
// the repository at dir, via git ls-files. Returns paths relative to dir.
func ListFiles(dir string) ([]string, error) {
	out, err := runOutput(dir, "ls-files", argLsCached, argLsOthers, argLsExcludeStandard)
	if err != nil {
		return nil, errs.Caller(errs.GIT001,
			"could not list files",
			err,
			errs.Context{
				Cause:      "the working directory may not be a git repository",
				Resolution: "run from inside a git repository, or use --no-gitignore",
			},
		)
	}
	return parseLsFiles(out), nil
}

// parseLsFiles converts raw git ls-files output into OS-separator paths.
func parseLsFiles(out []byte) []string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, filepath.FromSlash(line))
	}
	return files
}

// CurrentBranch returns the current branch name in the repository at dir.
// Returns an error when HEAD is detached (there is no branch to name).
func CurrentBranch(dir string) (string, error) {
	out, err := runOutput(dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentCommit returns the commit SHA that HEAD points to in the
// repository at dir.
func CurrentCommit(dir string) (string, error) {
	out, err := runOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// IsShallow reports whether the repository at dir is a shallow clone.
func IsShallow(dir string) (bool, error) {
	out, err := runOutput(dir, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == StrTrue, nil
}

// RemoteDefaultBranch resolves remote's actual default branch from
// its locally-recorded HEAD ref. Returns an error when that ref
// isn't recorded locally.
func RemoteDefaultBranch(dir, remote string) (string, error) {
	prefix := "refs/remotes/" + remote + "/"
	out, err := runOutput(dir, "symbolic-ref", prefix+"HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), prefix), nil
}

// IsDirty reports whether the working tree at dir has uncommitted changes.
func IsDirty(dir string) (bool, error) {
	out, err := runOutput(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// Version returns the local git installation's version string.
func Version() (string, error) {
	out, err := runOutput("", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
