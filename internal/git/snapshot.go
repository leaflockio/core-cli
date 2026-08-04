// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import "slices"

// conventionalRemote is the remote name checked for among several,
// only used as a tiebreaker when it's confirmed to actually be one
// of the configured remotes.
const conventionalRemote = "origin"

// Snapshot holds git-level facts about a directory, resolved once.
type Snapshot struct {
	// Installed reports whether the git binary is available on PATH.
	Installed bool
	// IsRepo reports whether the directory is inside a git repository.
	// The remaining fields are only ever populated when this is true.
	IsRepo bool
	// RootDir is the repository root.
	RootDir string
	// DefaultRemote is the repository's actual default remote: the only
	// remote configured, or "origin" when there are several and it's one
	// of them. Empty when neither holds (no remotes, or several with no
	// "origin" among them).
	DefaultRemote string
	// RemoteURL is the URL of DefaultRemote, or "" if it's empty too.
	RemoteURL string
	// Branch is the current branch, or "" when HEAD is detached.
	Branch string
	// CommitSHA is the commit HEAD points to.
	CommitSHA string
	// Shallow reports whether the repository is a shallow clone.
	Shallow bool
	// DefaultBranch is the default remote's actual default branch, resolved
	// from its recorded HEAD. Empty when that isn't recorded locally.
	DefaultBranch string
	// Dirty reports whether the working tree has uncommitted changes.
	Dirty bool
	// Version is the local git installation's version string.
	Version string
}

// Detect resolves a Snapshot for dir. Always returns a usable Snapshot —
// an unavailable git binary or a non-repository directory just leave
// Installed/IsRepo false, not an error. The returned errors are the raw,
// unwrapped reasons any individual fact couldn't be resolved. Nil when
// every fact resolved successfully.
func Detect(dir string) (*Snapshot, []error) {
	s := &Snapshot{}
	if !IsInstalled() {
		return s, nil
	}
	s.Installed = true

	root, err := RepoRoot(dir)
	if err != nil {
		return s, nil
	}
	s.IsRepo = true
	s.RootDir = root

	var errs []error
	collect := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	remotes, err := ListRemotes(root)
	collect(err)
	s.DefaultRemote = resolveDefaultRemote(remotes)

	if s.DefaultRemote != "" {
		s.RemoteURL, err = RemoteURL(root, s.DefaultRemote)
		collect(err)
	}
	s.Branch, err = CurrentBranch(root)
	collect(err)
	s.CommitSHA, err = CurrentCommit(root)
	collect(err)
	s.Shallow, err = IsShallow(root)
	collect(err)
	if s.DefaultRemote != "" {
		s.DefaultBranch, err = RemoteDefaultBranch(root, s.DefaultRemote)
		collect(err)
	}
	s.Dirty, err = IsDirty(root)
	collect(err)
	s.Version, err = Version()
	collect(err)

	return s, errs
}

// BaseRef returns the ref for this Snapshot's resolved default remote and
// branch. Errors when either wasn't determined — see DefaultRemote/DefaultBranch.
func (s *Snapshot) BaseRef() (string, error) {
	return BaseRefFrom(s.DefaultRemote, s.DefaultBranch)
}

// resolveDefaultRemote picks the one remote to treat as the default. With
// exactly one remote configured, it's unambiguously that one, regardless
// of its name. With more than one, conventionalRemote is used only when it's
// actually among them. Otherwise there's no reliable signal to pick from,
// and this returns "".
func resolveDefaultRemote(remotes []string) string {
	if len(remotes) == 1 {
		return remotes[0]
	}
	if slices.Contains(remotes, conventionalRemote) {
		return conventionalRemote
	}
	return ""
}
