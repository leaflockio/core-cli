// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package repo detects information about the current repository at startup.
// The result is injected into App and consumed by domain packages that need
// to know whether they are in a git repo, which languages are present, etc.
package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/leaflock/core-cli/internal/fstree"
	"github.com/leaflock/core-cli/internal/repo/lang"
)

const defaultRemote = "origin"

// License type identifiers returned by DetectedLicenseType.
const (
	licenseApache20    = "apache-2.0"
	licenseMIT         = "mit"
	licenseGPL3        = "gpl-3.0"
	licenseGPL2        = "gpl-2.0"
	licenseLGPL3       = "lgpl-3.0"
	licenseLGPL21      = "lgpl-2.1"
	licenseAGPL3       = "agpl-3.0"
	licenseMPL2        = "mpl-2.0"
	licenseBSD2Clause  = "bsd-2-clause"
	licenseBSD3Clause  = "bsd-3-clause"
	licenseISC         = "isc"
	licenseProprietary = "proprietary"
)

// Keywords used for license text classification.
const (
	kwApacheLicense           = "APACHE LICENSE"
	kwApacheVersion           = "2.0"
	kwMITLicense              = "MIT LICENSE"
	kwPermissionGranted       = "PERMISSION IS HEREBY GRANTED"
	kwWithoutWarranty         = "WITHOUT WARRANTY"
	kwGNUGeneralPublicLicense = "GNU GENERAL PUBLIC LICENSE"
	kwGNULesserPublicLicense  = "GNU LESSER GENERAL PUBLIC LICENSE"
	kwGNUAfferoPublicLicense  = "GNU AFFERO GENERAL PUBLIC LICENSE"
	kwMozillaPublicLicense    = "MOZILLA PUBLIC LICENSE"
	kwBSD2Clause              = "BSD 2-CLAUSE"
	kwBSDTwoClause            = "BSD TWO-CLAUSE"
	kwBSD3Clause              = "BSD 3-CLAUSE"
	kwBSDThreeClause          = "BSD THREE-CLAUSE"
	kwISCLicense              = "ISC LICENSE"
	kwVersion2                = "VERSION 2"
	kwVersion3                = "VERSION 3"
)

// Mockable runner for git root detection.
var execGitRoot = func() ([]byte, error) {
	return exec.Command("git", "rev-parse", "--show-toplevel").Output()
}

// Mockable runner for git remote URL detection.
var execGitRemoteURL = func() ([]byte, error) {
	return exec.Command("git", "remote", "get-url", defaultRemote).Output()
}

// Mockable file walker for language detection.
var walkFiles = fstree.Walk

// Mockable working directory resolver.
var osGetwd = os.Getwd

// Info holds repository information detected at startup.
// All fields are best-effort; undetectable fields are left at their zero value.
type Info struct {
	// IsGit reports whether the working directory is inside a git repository.
	IsGit bool
	// RootDir is the absolute path to the repository or working directory root.
	RootDir string
	// RemoteURL is the URL of the git remote named "origin". Empty if not a git
	// repo or no remote named "origin" exists.
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
	// Languages contains the languages detected in the repository, ordered by
	// file count descending.
	Languages lang.Detection
	// HasLicenseFile reports whether a LICENSE, LICENSE.md, or LICENSE.txt file
	// exists at the repository root.
	HasLicenseFile bool
	// DetectedLicenseType is a best-effort license type string derived from the
	// content of the LICENSE file (e.g. "apache-2.0", "mit", "proprietary").
	// Empty when HasLicenseFile is false.
	DetectedLicenseType string
}

// Detect builds an Info by inspecting the current working directory.
// It never returns an error; partial information is returned when detection fails.
func Detect() *Info {
	info := &Info{}

	if out, err := execGitRoot(); err == nil {
		info.IsGit = true
		info.RootDir = strings.TrimSpace(string(out))
		if out2, err := execGitRemoteURL(); err == nil {
			info.RemoteURL = strings.TrimSpace(string(out2))
		}
		info.Host, info.Owner, info.RepoName = parseRemoteURL(info.RemoteURL)
	} else {
		wd, err := osGetwd()
		if err != nil {
			wd = ""
		}
		info.RootDir = wd
	}

	info.HasLicenseFile, info.DetectedLicenseType = detectLicense(info.RootDir)
	info.Languages = detectLanguages(info.RootDir)

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

	// SSH format: git@github.com:org/repo
	if strings.HasPrefix(url, "git@") {
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

	// HTTPS format: https://github.com/org/repo
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

// detectLicense checks whether a LICENSE file exists at root and performs a
// best-effort classification of its type.
func detectLicense(root string) (bool, string) {
	for _, name := range licenseFileNames {
		p := filepath.Join(root, name)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		return true, classifyLicenseText(string(data))
	}
	return false, ""
}

// classifyLicenseText performs a simple keyword-based license identification.
func classifyLicenseText(text string) string {
	upper := strings.ToUpper(text)
	if r := classifyCommonLicenses(upper); r != "" {
		return r
	}
	return classifyRareLicenses(upper)
}

// classifyCommonLicenses matches the most widely used open-source licenses.
func classifyCommonLicenses(upper string) string {
	switch {
	case strings.Contains(upper, kwApacheLicense) && strings.Contains(upper, kwApacheVersion):
		return licenseApache20
	case strings.Contains(upper, kwMITLicense) ||
		(strings.Contains(upper, kwPermissionGranted) &&
			strings.Contains(upper, kwWithoutWarranty)):
		return licenseMIT
	case strings.Contains(upper, kwGNUGeneralPublicLicense) &&
		strings.Contains(upper, kwVersion3):
		return licenseGPL3
	case strings.Contains(upper, kwGNUGeneralPublicLicense) &&
		strings.Contains(upper, kwVersion2):
		return licenseGPL2
	case strings.Contains(upper, kwGNULesserPublicLicense) &&
		strings.Contains(upper, kwVersion3):
		return licenseLGPL3
	case strings.Contains(upper, kwGNULesserPublicLicense) &&
		strings.Contains(upper, kwVersion2):
		return licenseLGPL21
	}
	return ""
}

// classifyRareLicenses matches less common open-source licenses, defaulting
// to proprietary when no pattern matches.
func classifyRareLicenses(upper string) string {
	switch {
	case strings.Contains(upper, kwGNUAfferoPublicLicense):
		return licenseAGPL3
	case strings.Contains(upper, kwMozillaPublicLicense):
		return licenseMPL2
	case strings.Contains(upper, kwBSD2Clause) || strings.Contains(upper, kwBSDTwoClause):
		return licenseBSD2Clause
	case strings.Contains(upper, kwBSD3Clause) || strings.Contains(upper, kwBSDThreeClause):
		return licenseBSD3Clause
	case strings.Contains(upper, kwISCLicense):
		return licenseISC
	default:
		return licenseProprietary
	}
}

// countLanguages maps each recognized language to its file count in files.
func countLanguages(files []string) map[lang.Language]int {
	counts := make(map[lang.Language]int)
	for _, f := range files {
		ext := filepath.Ext(f)
		if l := lang.FromExtension(ext); l != lang.Unknown {
			counts[l]++
		}
	}
	return counts
}

// buildDetection constructs a Detection from a language→count map, ordered by
// file count descending.
func buildDetection(counts map[lang.Language]int) lang.Detection {
	type entry struct {
		l     lang.Language
		count int
	}
	entries := make([]entry, 0, len(counts))
	for l, c := range counts {
		entries = append(entries, entry{l, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].l < entries[j].l
	})

	det := lang.Detection{FileCounts: counts}
	for _, e := range entries {
		det.All = append(det.All, e.l)
	}
	if len(det.All) > 0 {
		det.Primary = det.All[0]
	}
	return det
}

// detectLanguages scans the repository for source files and returns a
// Detection ordered by file count descending.
func detectLanguages(root string) lang.Detection {
	files, err := walkFiles(root, true)
	if err != nil {
		files = nil
	}
	return buildDetection(countLanguages(files))
}
