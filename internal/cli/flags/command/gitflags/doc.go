// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package gitflags holds reusable command flags for selecting files by git
// state.
//
// # Staged
//
// Staged (--staged) limits the file selection to what's currently staged in
// git — the same set git diff --cached would show.
//
// # PR
//
// PR (--pr) limits the file selection to what changed in the current pull
// request — the files that differ between the current HEAD and Base. See
// Base to change what HEAD is compared against.
//
// # Base
//
// Base (--base) is the branch, tag, or commit that PR (--pr) diffs
// against. Any git revision expression is accepted — a branch name, a
// remote-qualified branch (origin/main), a tag, a full or abbreviated
// commit SHA, or a suffixed expression (HEAD~2). It's meaningful only
// alongside PR: used on its own, there is nothing to diff against. An
// empty value means none was given.
//
// # DiffFilter's value format
//
// DiffFilter (--diff-filter) controls which kinds of file changes are
// included, using the same character codes as git's own --diff-filter:
//
//	A  Added
//	C  Copied
//	D  Deleted
//	M  Modified
//	R  Renamed
//	T  Type changed
//	U  Unmerged
//	X  Unknown
//	B  Pairing broken
//
// Any combination of these letters may be used, e.g. "ACM" (the default)
// includes added, copied, and modified files. A letter may be lowercased to
// exclude that status instead of including it — e.g. "acm" excludes added,
// copied, and modified files.
//
// A single trailing "*" may follow the letters, e.g. "ACM*" — meant to be
// combined with letters, not used alone. It's a modifier, not an ordinary
// filter character: "ACM*" means "if any file is added, copied, or
// modified, widen the result to every file in this diff — otherwise select
// nothing." That's a different, wider result than plain "ACM" (which only
// selects the added/copied/modified files themselves). A bare "*" with no
// letters is a degenerate case with nothing to widen from — it selects
// nothing at all, not "everything," so it's not a useful value on its own.
//
// # NoGitignore
//
// NoGitignore (--no-gitignore) disables gitignore-aware file discovery, so
// files a .gitignore would normally hide are included too.
package gitflags
