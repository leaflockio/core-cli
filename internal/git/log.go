// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import (
	"time"

	"github.com/leaflockio/core-cli/internal/errs"
)

// FileCreatedDate resolves the author date of path's earliest commit in
// the repository at dir, following renames.
func FileCreatedDate(dir, path string) (time.Time, error) {
	return fileLogDate(dir, path, "--reverse")
}

// FileModifiedDate resolves the author date of path's most recent commit
// in the repository at dir, following renames.
func FileModifiedDate(dir, path string) (time.Time, error) {
	return fileLogDate(dir, path, "-1")
}

// fileLogDate runs git log on path with the given ordering/limit flag and
// parses the first resulting commit's author date.
func fileLogDate(dir, path, flag string) (time.Time, error) {
	out, err := runOutput(dir, "log", "--follow", "--format=%aI", flag, "--", path)
	if err != nil {
		return time.Time{}, err
	}
	lines := splitLines(string(out))
	if len(lines) == 0 {
		return time.Time{}, errs.Caller(errs.GIT006,
			"no commit history for "+path,
			nil,
			errs.Context{
				Cause:      "the file has no commits yet — normal for a new, untracked, or uncommitted file",
				Resolution: "commit the file to give it a git history to derive a date from",
			},
		)
	}
	return time.Parse(time.RFC3339, lines[0])
}
