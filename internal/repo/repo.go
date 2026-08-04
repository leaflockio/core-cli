// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package repo detects information about the current repository at startup.
// It inspects the working directory for git metadata, remote URL, language
// composition, and license file presence.
package repo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/licenseclassifier/v2/assets"
	"github.com/leaflockio/core-cli/internal/fstree"
	"github.com/leaflockio/core-cli/internal/git"
	"github.com/leaflockio/core-cli/internal/repo/lang"
)

// detectGit resolves the git details for dir.
var detectGit = git.Detect

// walkFiles walks the repository file tree.
var walkFiles = fstree.Walk

// osGetwd resolves the current working directory.
var osGetwd = os.Getwd

// LicenseInfo holds the result of license file detection at the repo root.
type LicenseInfo struct {
	// Found reports whether a license file was found at the repo root.
	Found bool
	// File is the name of the detected license file (e.g. "LICENSE", "LICENSE.md").
	File string
	// SPDXID is the SPDX identifier detected from the license file content
	// (e.g. "MIT", "Apache-2.0"). Empty when the license could not be classified.
	SPDXID string
}

// Info holds repository information detected at startup.
// All fields are best-effort; undetectable fields are left at their zero value.
type Info struct {
	// IsGit reports whether the working directory is inside a git repository.
	IsGit bool
	// RootDir is the absolute path to the repository or working directory root.
	RootDir string
	// RemoteURL is the URL of the repository's default remote — the only
	// remote configured, or "origin" when there are several and it's one of
	// them. Empty if not a git repo or no default remote could be resolved.
	RemoteURL string
	// Host is the git hosting provider parsed from RemoteURL (e.g. "github.com").
	// Empty when RemoteURL is empty.
	Host string
	// Owner is the organization or username portion of RemoteURL (e.g. "leaflockio").
	// Empty when RemoteURL is empty.
	Owner string
	// RepoName is the repository name portion of RemoteURL (e.g. "core-cli").
	// Empty when RemoteURL is empty.
	RepoName string
	// Languages is the language composition of the repository, ordered by
	// file count descending.
	Languages lang.Composition
	// License holds the result of license file detection at the repo root.
	License LicenseInfo
}

// Detect builds an Info by inspecting the current working directory.
// It never returns an error; partial information is returned when detection fails.
func Detect() *Info {
	info := &Info{}

	wd, err := osGetwd()
	if err != nil {
		wd = ""
	}

	snap, _ := detectGit(wd)
	if snap.IsRepo {
		info.IsGit = true
		info.RootDir = snap.RootDir
		info.RemoteURL = snap.RemoteURL
		info.Host, info.Owner, info.RepoName = parseRemoteURL(info.RemoteURL)
	} else {
		info.RootDir = wd
	}

	info.License = detectLicense(info.RootDir)
	files, err := walkFiles(info.RootDir, true)
	if err != nil {
		files = nil
	}
	info.Languages = lang.Build(lang.Count(files))

	return info
}

// parseRemoteURL parses a git remote URL and extracts the host, owner, and
// repository name. Supports both HTTPS and SSH formats:
//
//	https://github.com/leaflockio/core-cli.git → github.com, leaflockio, core-cli
//	git@github.com:leaflockio/core-cli.git     → github.com, leaflockio, core-cli
func parseRemoteURL(rawURL string) (string, string, string) {
	if rawURL == "" {
		return "", "", ""
	}
	url := strings.TrimSuffix(rawURL, ".git")
	if strings.HasPrefix(url, "git@") {
		return parseSSHRemoteURL(url)
	}
	return parseHTTPSRemoteURL(url)
}

// parseSSHRemoteURL extracts host, owner, and repo from an SSH remote URL
// of the form git@github.com:org/repo.
func parseSSHRemoteURL(url string) (string, string, string) {
	url = strings.TrimPrefix(url, "git@")
	parts := strings.SplitN(url, ":", 2)
	if len(parts) != 2 {
		return "", "", ""
	}
	host := parts[0]
	var owner, repo string
	pathParts := strings.SplitN(parts[1], "/", 2)
	if len(pathParts) == 2 {
		owner, repo = pathParts[0], pathParts[1]
	}
	return host, owner, repo
}

// parseHTTPSRemoteURL extracts host, owner, and repo from an HTTPS remote URL
// of the form https://github.com/org/repo.
func parseHTTPSRemoteURL(url string) (string, string, string) {
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(url, prefix) {
			url = strings.TrimPrefix(url, prefix)
			break
		}
	}
	parts := strings.SplitN(url, "/", 3)
	if len(parts) < 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}

// licenseFileNames are the candidate filenames for a LICENSE file.
var licenseFileNames = []string{
	"LICENSE", "LICENSE.md", "LICENSE.txt",
}

// detectLicense checks whether a LICENSE file exists at root and classifies
// its content to extract the SPDX identifier.
func detectLicense(root string) LicenseInfo {
	for _, name := range licenseFileNames {
		p := filepath.Join(root, name)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		info := LicenseInfo{Found: true, File: name}
		c, err := assets.DefaultClassifier()
		if err == nil {
			results := c.Match(data)
			if len(results.Matches) > 0 {
				info.SPDXID = results.Matches[0].Name
			}
		}
		return info
	}
	return LicenseInfo{}
}
