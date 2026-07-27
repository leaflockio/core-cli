// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/util/fsutil"
	"github.com/spf13/cobra"
)

// Depth controls how much of the command path is reflected in the workspace
// directory hierarchy.
type Depth int

const (
	// DepthRoot roots the CommandSpace at the app root with no command segment.
	DepthRoot Depth = iota
	// DepthCommand roots the CommandSpace at the top-level command (e.g. pr/).
	DepthCommand
	// DepthFull roots the CommandSpace at the full command path (e.g. pr/create/).
	DepthFull
)

// Workspace manages filesystem paths for the CLI across user, repo, and
// cache scopes.
type Workspace struct {
	entityRoot string // ~/<EntityFolder>/
	userRoot   string // ~/<EntityFolder>/<AppName>/
	repoRoot   string // <repo-root>/<AppName>/
	cacheRoot  string // <UserCacheDir>/<AppName>/
}

// CommandSpace is a scoped view of the workspace for a specific command.
type CommandSpace struct {
	base string
	perm fs.FileMode
	err  error // set by the constructor when this CommandSpace cannot be used; Dir and File return it directly
}

// osUserHomeDir is the function used to resolve the home directory.
var osUserHomeDir = os.UserHomeDir

// osUserCacheDir is the function used to resolve the platform cache directory.
var osUserCacheDir = os.UserCacheDir

// New constructs a Workspace. If homeDir is empty, os.UserHomeDir is called.
// Returns WSP001 if the home directory cannot be resolved.
func New(homeDir, repoRoot string) (*Workspace, error) {
	if homeDir == "" {
		var err error
		homeDir, err = osUserHomeDir()
		if err != nil {
			return nil, errs.Caller(
				errs.WSP001,
				"home directory could not be resolved",
				err,
				errs.Context{
					Cause:      "HOME environment variable is not set and system user lookup failed",
					Resolution: "set the HOME environment variable and retry",
				},
			)
		}
	}
	ws := &Workspace{
		entityRoot: filepath.Join(homeDir, config.EntityFolder),
		userRoot:   filepath.Join(homeDir, config.EntityFolder, config.AppName),
		repoRoot:   filepath.Join(repoRoot, config.AppName),
	}
	if cacheDir, err := osUserCacheDir(); err == nil {
		ws.cacheRoot = filepath.Join(cacheDir, config.AppName)
	}
	return ws, nil
}

// ForUser returns a CommandSpace rooted at the user-scoped directory for cmd
// at the given depth. Directories are created with 0o700 — private to the
// owner, never readable by other users on the machine.
func (w *Workspace) ForUser(cmd *cobra.Command, depth Depth) *CommandSpace {
	return &CommandSpace{
		base: filepath.Join(w.userRoot, commandSubPath(cmd, depth)),
		perm: 0o700,
	}
}

// ForRepo returns a CommandSpace rooted at the repo-scoped directory for cmd
// at the given depth. Directories are created with 0o755 — readable by CI
// and other processes that access the repository.
func (w *Workspace) ForRepo(cmd *cobra.Command, depth Depth) *CommandSpace {
	return &CommandSpace{
		base: filepath.Join(w.repoRoot, commandSubPath(cmd, depth)),
		perm: 0o755,
	}
}

// ForProjectRoot returns a CommandSpace rooted at the bare repo-scoped leaf/
// directory, with no command segment. Directories are created with
// 0o755, matching ForRepo.
func (w *Workspace) ForProjectRoot() *CommandSpace {
	return &CommandSpace{
		base: w.repoRoot,
		perm: 0o755,
	}
}

// ForGenerated returns a CommandSpace rooted at the repo-scoped generated
// output directory for cmd at the given depth. Generated files are files
// this tool fully owns and writes — lock files, derived artifacts. Directories
// are created with 0o755, matching ForRepo.
func (w *Workspace) ForGenerated(cmd *cobra.Command, depth Depth) *CommandSpace {
	return &CommandSpace{
		base: filepath.Join(w.repoRoot, config.GeneratedDir, commandSubPath(cmd, depth)),
		perm: 0o755,
	}
}

// ForCache returns a CommandSpace rooted at the user's platform cache
// directory for cmd and purpose. Directories are created with 0o755 — cache
// data is disposable and not sensitive. If the platform cache directory
// could not be resolved at construction time, Dir and File on the returned
// CommandSpace return WSP002 instead of a path.
func (w *Workspace) ForCache(cmd *cobra.Command, purpose string) *CommandSpace {
	if w.cacheRoot == "" {
		return &CommandSpace{err: errs.Caller(
			errs.WSP002,
			"cache directory could not be resolved",
			nil,
			errs.Context{
				Cause:      "the platform cache directory API failed",
				Resolution: "cache-scoped operations are unavailable for this run",
			},
		)}
	}
	return &CommandSpace{
		base: filepath.Join(w.cacheRoot, commandSubPath(cmd, DepthCommand), purpose),
		perm: 0o755,
	}
}

// CredentialsPath returns the path to the credentials file.
func (w *Workspace) CredentialsPath() string {
	return filepath.Join(w.entityRoot, config.CredentialsFile)
}

// Dir returns the path to a named subdirectory within this CommandSpace,
// creating it if it does not exist.
func (cs *CommandSpace) Dir(name string) (string, error) {
	if cs.err != nil {
		return "", cs.err
	}
	return fsutil.EnsureDir(filepath.Join(cs.base, name), cs.perm)
}

// File returns the path to a named file within this CommandSpace, creating
// the parent directory if it does not exist.
func (cs *CommandSpace) File(name string) (string, error) {
	if cs.err != nil {
		return "", cs.err
	}
	return fsutil.EnsureParent(filepath.Join(cs.base, name), cs.perm)
}

// Peek returns the path to a named file or subdirectory within this
// CommandSpace, or the CommandSpace's own root path when name is "". It is
// read-only — nothing on disk is created as a result of calling it.
func (cs *CommandSpace) Peek(name string) (string, error) {
	if cs.err != nil {
		return "", cs.err
	}
	return filepath.Join(cs.base, name), nil
}

// commandSubPath extracts the path segment from cmd's command path based on depth.
//
//	DepthRoot    → ""              (e.g. gh → <root>/)
//	DepthCommand → "pr"            (e.g. gh pr create → <root>/pr/)
//	DepthFull    → "pr/create"     (e.g. gh pr create → <root>/pr/create/)
func commandSubPath(cmd *cobra.Command, depth Depth) string {
	if depth == DepthRoot {
		return ""
	}
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) <= 1 {
		return ""
	}
	sub := parts[1:] // drop binary name
	if depth == DepthCommand {
		return sub[0]
	}
	return filepath.Join(sub...)
}
