// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package fstree resolves which files a command should operate on: a plain
// recursive walk, glob-based include/exclude filtering, or turning
// positional CLI arguments into a concrete file list.
//
// Example tree used throughout this documentation, with cwd /repo:
//
//	/repo
//	├── main.go
//	├── main_test.go
//	├── build                    (a file)
//	├── config
//	│   ├── settings.json
//	│   └── settings.yaml
//	├── src
//	│   ├── index.js
//	│   ├── App.jsx
//	│   └── build                (a directory — same name as the file above)
//	│       └── bundle.js
//	└── vendor
//	    ├── deps.go
//	    └── utils
//	        └── patched.go
//
// # Two ways to get a file list
//
// [Walk] plus [Filter] is the "gather everything, then narrow it down" path
// — Walk lists every file under a root, Filter narrows that list with
// include/exclude glob patterns. This is what backs a --all flag paired
// with --include/--exclude.
//
// [Resolve] is the "turn what the user typed into files" path — a single
// call that handles a literal file, a literal directory, or a glob pattern,
// whatever was given as a positional argument.
//
// # The same pattern means different things in each
//
//	Filter(Walk(".")-result, ["*.go"], nil) → main.go, main_test.go, vendor/deps.go, vendor/utils/patched.go
//	Resolve(["*.go"])                       → main.go, main_test.go   (not vendor/deps.go)
//
// Filter's "*.go" matches by file name alone, at any depth — a pattern with
// no "/" isn't anchored to one directory level. Resolve's "*.go" only
// matches within a single directory — "*" never crosses a "/". Each
// behaves the way its own input calls for: Filter narrows an already
// broad list of files down by name, so matching anywhere is the useful
// default; Resolve expands one argument someone typed, where crossing
// into unrelated directories the same argument didn't ask for would be
// surprising.
//
// # Walk: plain recursive listing
//
//	Walk(".")       → every file in the tree except anything under .git
//	Walk("config")  → config/settings.json, config/settings.yaml
//
// # Filter: include/exclude
//
// A file is kept only if it matches an include pattern and isn't excluded.
// A bare pattern like "*.json" matches by name at any depth; adding a "/"
// anchors it to that exact path; "**" crosses directories explicitly;
// "{a,b}" matches either alternative; a trailing "/" matches a directory
// and never a same-named file; and a "!" exclude entry cancels a match from
// an earlier exclude entry, so the last one that matches a given file wins.
//
//	Filter(all, ["**/*.{js,jsx}"], nil)
//	    → src/App.jsx, src/build/bundle.js, src/index.js
//	Filter(all, [MatchAll], ["*.json"])
//	    → everything except config/settings.json
//	Filter(all, ["config/*.json"], nil)
//	    → config/settings.json only (anchored — has a "/")
//	Filter(all, [MatchAll], ["vendor/**"])
//	    → everything except vendor
//	Filter(all, [MatchAll], ["vendor/**", "!vendor/utils/patched.go"])
//	    → same, but vendor/utils/patched.go survives
//	Filter(all, [MatchAll], ["build/"])
//	    → excludes src/build/bundle.js, keeps the file named build
//
// # Resolve: positional arguments
//
//	Resolve(["main.go"])     → the file, as an absolute path
//	Resolve(["vendor"])      → vendor/deps.go, vendor/utils/patched.go
//	Resolve(["*.go"])        → main.go, main_test.go
//	Resolve(["missing.go"])  → error — not a pattern, and doesn't exist
//	Resolve(["*.xyz"])       → [] — a pattern matching nothing isn't an error
//
// Every path Resolve returns is absolute and symlink-resolved, even when
// the input was relative, so the result stays valid no matter what the
// working directory does afterward, and the same file reached through two
// different spellings resolves to one identical string.
package fstree
