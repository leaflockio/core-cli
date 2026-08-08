// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package fstree

import (
	"os"
	"path/filepath"
	"strings"
)

// statFile reports whether a path exists, and whether it's a directory.
var statFile = os.Stat

// absPath makes a path absolute.
var absPath = filepath.Abs

// Resolve turns positional path arguments into absolute file paths. Using
// the package doc's example tree:
//
//	Resolve(["main.go"])                 → ["/repo/main.go"]
//	Resolve(["config"])                  → ["/repo/config/settings.json", "/repo/config/settings.yaml"]
//	Resolve(["."])                       → every file in the tree
//	Resolve(["*.go"])                    → ["/repo/main.go"]
//	Resolve(["**/*.{js,jsx}"])           → ["/repo/src/index.js", "/repo/src/App.jsx"]
//	Resolve(["config/*.{json,yaml}"])    → ["/repo/config/settings.json", "/repo/config/settings.yaml"]
//	Resolve(["missing.go"])              → error (doesn't exist, and isn't a pattern)
//	Resolve(["*.xyz"])                   → [] (a pattern matching nothing is not an error)
func Resolve(paths []string) ([]string, error) {
	seen := make(map[string]bool, len(paths))
	var out []string
	for _, p := range paths {
		files, err := resolveOne(p)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if seen[f] {
				continue
			}
			seen[f] = true
			out = append(out, f)
		}
	}
	return out, nil
}

// canonicalize(root) → absolute path.
//
//	canonicalize("main.go") → "/repo/main.go"
func canonicalize(root string) (string, error) {
	return absPath(root)
}

// rel makes an absolute path relative to root, in slash form.
//
//	rel("/repo", "/repo/config/settings.json") → "config/settings.json"
func rel(root, path string) (string, error) {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(r), nil
}

// RebasePatterns rewrites every absolute pattern in patterns to be relative
// to root, leaving a relative pattern or glob untouched. A leading "!" is
// preserved around the rebase, so an exclude negation entry keeps working:
//
//	RebasePatterns("/repo", ["/repo/main.go", "*.md", "!/repo/vendor/**"])
//	    → ["main.go", "*.md", "!vendor/**"]
func RebasePatterns(root string, patterns []string) []string {
	out := make([]string, len(patterns))
	for i, p := range patterns {
		negated, isNegated := strings.CutPrefix(p, "!")
		target := p
		if isNegated {
			target = negated
		}

		rebased := target
		if filepath.IsAbs(target) {
			if r, err := rel(root, target); err == nil {
				rebased = r
			}
		}

		if isNegated {
			rebased = "!" + rebased
		}
		out[i] = rebased
	}
	return out
}

// rebase(root, files) joins files — already relative to root, per fsWalk's
// contract — onto root's canonical form:
//
//	rebase("config", ["settings.json", "settings.yaml"])
//	    → ["/repo/config/settings.json", "/repo/config/settings.yaml"]
//
// root is only canonicalized once here, however many files there are.
func rebase(root string, files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	canonicalRoot, err := canonicalize(root)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = filepath.Join(canonicalRoot, f)
	}
	return out, nil
}

// resolveOne expands a single path argument per Resolve's rules.
func resolveOne(path string) ([]string, error) {
	info, err := statFile(path)
	switch {
	case err == nil && info.IsDir():
		files, werr := fsWalk(path)
		if werr != nil {
			return nil, werr
		}
		return rebase(path, files)
	case err == nil:
		c, cerr := canonicalize(path)
		if cerr != nil {
			return nil, cerr
		}
		return []string{c}, nil
	case !hasGlobMeta(path):
		return nil, err
	}

	root := globPrefix(path)
	candidates, err := fsWalk(root)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, f := range candidates {
		full := filepath.ToSlash(filepath.Join(root, f))
		if globMatch(path, full) {
			matches = append(matches, f)
		}
	}
	return rebase(root, matches)
}

// hasGlobMeta reports whether s contains glob meta characters, meaning it
// should be treated as a pattern rather than a literal path when it doesn't
// exist on disk.
func hasGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[{")
}

// globPrefix returns the directory to walk when resolving pattern as a
// glob: everything before the first path segment containing a wildcard.
// Returns "." when pattern has no static directory component.
func globPrefix(pattern string) string {
	segments := strings.Split(filepath.ToSlash(pattern), "/")
	end := len(segments)
	for i, seg := range segments {
		if hasGlobMeta(seg) {
			end = i
			break
		}
	}
	prefix := strings.Join(segments[:end], "/")
	if prefix == "" {
		return "."
	}
	return filepath.FromSlash(prefix)
}
