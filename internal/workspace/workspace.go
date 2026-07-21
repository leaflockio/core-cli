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

// Workspace manages filesystem paths for the CLI across user and repo scopes.
type Workspace struct {
	entityRoot string // ~/<EntityFolder>/
	userRoot   string // ~/<EntityFolder>/<AppName>/
	repoRoot   string // <repo-root>/<AppName>/
}

// CommandSpace is a scoped view of the workspace for a specific command.
type CommandSpace struct {
	base string
	perm fs.FileMode
}

// osUserHomeDir is the function used to resolve the home directory. Tests
// override this to simulate failures without touching the real home directory.
var osUserHomeDir = os.UserHomeDir

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
	return &Workspace{
		entityRoot: filepath.Join(homeDir, config.EntityFolder),
		userRoot:   filepath.Join(homeDir, config.EntityFolder, config.AppName),
		repoRoot:   filepath.Join(repoRoot, config.AppName),
	}, nil
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

// CredentialsPath returns the path to the credentials file.
func (w *Workspace) CredentialsPath() string {
	return filepath.Join(w.entityRoot, config.CredentialsFile)
}

// Dir returns the path to a named subdirectory within this CommandSpace,
// creating it if it does not exist.
func (cs *CommandSpace) Dir(name string) (string, error) {
	return fsutil.EnsureDir(filepath.Join(cs.base, name), cs.perm)
}

// File returns the path to a named file within this CommandSpace, creating
// the parent directory if it does not exist.
func (cs *CommandSpace) File(name string) (string, error) {
	return fsutil.EnsureParent(filepath.Join(cs.base, name), cs.perm)
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
