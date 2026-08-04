// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package repo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/git"
)

var (
	errWalkFailed  = errors.New("walk failed")
	errGetwdFailed = errors.New("getwd failed")
)

// --- Detect ---

func TestDetect_gitRepoWithRemote(t *testing.T) {
	dir := t.TempDir()
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) {
		return &git.Snapshot{
			IsRepo:    true,
			RootDir:   dir,
			RemoteURL: "https://github.com/leaflockio/core-cli.git",
		}, nil
	}
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }

	info := Detect()

	if !info.IsGit {
		t.Error("expected IsGit=true")
	}
	if info.RootDir != dir {
		t.Errorf("expected RootDir=%q, got %q", dir, info.RootDir)
	}
	if info.RemoteURL != "https://github.com/leaflockio/core-cli.git" {
		t.Errorf("unexpected RemoteURL: %q", info.RemoteURL)
	}
	if info.Host != "github.com" {
		t.Errorf("expected Host=github.com, got %q", info.Host)
	}
	if info.Owner != "leaflockio" {
		t.Errorf("expected Owner=leaflockio, got %q", info.Owner)
	}
	if info.RepoName != "core-cli" {
		t.Errorf("expected RepoName=core-cli, got %q", info.RepoName)
	}
	if info.License.Found {
		t.Error("expected License.Found=false for empty dir")
	}
}

func TestDetect_gitRepoRemoteFails(t *testing.T) {
	dir := t.TempDir()
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) {
		return &git.Snapshot{IsRepo: true, RootDir: dir}, nil
	}
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }

	info := Detect()

	if !info.IsGit {
		t.Error("expected IsGit=true")
	}
	if info.RemoteURL != "" {
		t.Errorf("expected empty RemoteURL, got %q", info.RemoteURL)
	}
	if info.Host != "" || info.Owner != "" || info.RepoName != "" {
		t.Error("expected empty Host/Owner/RepoName when RemoteURL is empty")
	}
}

func TestDetect_notGitRepo(t *testing.T) {
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) { return &git.Snapshot{}, nil }
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }

	info := Detect()

	if info.IsGit {
		t.Error("expected IsGit=false")
	}
	if info.RootDir == "" {
		t.Error("expected RootDir to be set from os.Getwd")
	}
}

func TestDetect_getwdFails(t *testing.T) {
	origDetect := detectGit
	origWalk := walkFiles
	origGetwd := osGetwd
	defer func() { detectGit = origDetect; walkFiles = origWalk; osGetwd = origGetwd }()

	detectGit = func(_ string) (*git.Snapshot, []error) { return &git.Snapshot{}, nil }
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }
	osGetwd = func() (string, error) { return "", errGetwdFailed }

	info := Detect()

	if info.IsGit {
		t.Error("expected IsGit=false")
	}
	if info.RootDir != "" {
		t.Errorf("expected empty RootDir when getwd fails, got %q", info.RootDir)
	}
}

func TestDetect_withLicenseFile(t *testing.T) {
	dir := t.TempDir()
	mitText := `MIT License

Copyright (c) 2024 LeafLock

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte(mitText), 0o600); err != nil {
		t.Fatal(err)
	}
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) {
		return &git.Snapshot{IsRepo: true, RootDir: dir}, nil
	}
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }

	info := Detect()

	if !info.License.Found {
		t.Error("expected License.Found=true")
	}
	if info.License.SPDXID == "" {
		t.Error("expected License.SPDXID to be detected")
	}
}

func TestDetect_withLanguages(t *testing.T) {
	dir := t.TempDir()
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) {
		return &git.Snapshot{IsRepo: true, RootDir: dir}, nil
	}
	walkFiles = func(_ string, _ bool) ([]string, error) {
		return []string{
			filepath.Join("repo", "main.go"),
			filepath.Join("repo", "util.go"),
			filepath.Join("repo", "app.ts"),
		}, nil
	}

	info := Detect()

	if info.Languages.Primary != "Go" {
		t.Errorf("expected Primary=Go, got %v", info.Languages.Primary)
	}
}

// --- parseRemoteURL ---

func TestParseRemoteURL_empty(t *testing.T) {
	h, o, r := parseRemoteURL("")
	if h != "" || o != "" || r != "" {
		t.Errorf("expected empty results for empty input, got %q %q %q", h, o, r)
	}
}

func TestParseRemoteURL_httpsWithGitSuffix(t *testing.T) {
	h, o, r := parseRemoteURL("https://github.com/leaflockio/core-cli.git")
	if h != "github.com" || o != "leaflockio" || r != "core-cli" {
		t.Errorf("unexpected: host=%q owner=%q repo=%q", h, o, r)
	}
}

func TestParseRemoteURL_httpsWithoutGitSuffix(t *testing.T) {
	h, o, r := parseRemoteURL("https://github.com/leaflockio/core-cli")
	if h != "github.com" || o != "leaflockio" || r != "core-cli" {
		t.Errorf("unexpected: host=%q owner=%q repo=%q", h, o, r)
	}
}

func TestParseRemoteURL_httpScheme(t *testing.T) {
	h, o, r := parseRemoteURL("http://github.com/leaflockio/core-cli")
	if h != "github.com" || o != "leaflockio" || r != "core-cli" {
		t.Errorf("unexpected: host=%q owner=%q repo=%q", h, o, r)
	}
}

func TestParseRemoteURL_httpsInsufficientParts(t *testing.T) {
	h, o, r := parseRemoteURL("https://github.com/onlyorg")
	if h != "" || o != "" || r != "" {
		t.Errorf("expected empty results for incomplete HTTPS URL, got %q %q %q", h, o, r)
	}
}

func TestParseRemoteURL_sshFull(t *testing.T) {
	h, o, r := parseRemoteURL("git@github.com:leaflockio/core-cli.git")
	if h != "github.com" || o != "leaflockio" || r != "core-cli" {
		t.Errorf("unexpected: host=%q owner=%q repo=%q", h, o, r)
	}
}

func TestParseRemoteURL_sshNoColon(t *testing.T) {
	h, o, r := parseRemoteURL("git@github.com")
	if h != "" || o != "" || r != "" {
		t.Errorf("expected empty results for SSH URL with no colon, got %q %q %q", h, o, r)
	}
}

func TestParseRemoteURL_sshNoSlashInPath(t *testing.T) {
	h, o, r := parseRemoteURL("git@github.com:leaflockio")
	if h != "github.com" || o != "" || r != "" {
		t.Errorf("expected host-only result for SSH URL without slash, got %q %q %q", h, o, r)
	}
}

// --- detectLicense ---

func TestDetectLicense_noFile(t *testing.T) {
	dir := t.TempDir()
	info := detectLicense(dir)
	if info.Found || info.File != "" || info.SPDXID != "" {
		t.Errorf("expected empty LicenseInfo, got %+v", info)
	}
}

func TestDetectLicense_foundLicenseFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("MIT LICENSE\nPermission is hereby granted")
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	info := detectLicense(dir)
	if !info.Found {
		t.Error("expected Found=true")
	}
	if info.File != "LICENSE" {
		t.Errorf("expected File=LICENSE, got %q", info.File)
	}
}

func TestDetectLicense_licenseMdFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("MIT LICENSE\nPermission is hereby granted")
	if err := os.WriteFile(filepath.Join(dir, "LICENSE.md"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	info := detectLicense(dir)
	if !info.Found {
		t.Error("expected Found=true")
	}
	if info.File != "LICENSE.md" {
		t.Errorf("expected File=LICENSE.md, got %q", info.File)
	}
}

func TestDetectLicense_preferLicenseOverMd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT LICENSE"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "LICENSE.md"), []byte("Apache License"), 0o600); err != nil {
		t.Fatal(err)
	}
	info := detectLicense(dir)
	if info.File != "LICENSE" {
		t.Errorf("expected LICENSE to take precedence, got %q", info.File)
	}
}

// --- walk error ---

func TestDetect_walkError(t *testing.T) {
	dir := t.TempDir()
	origDetect := detectGit
	origWalk := walkFiles
	defer func() { detectGit = origDetect; walkFiles = origWalk }()

	detectGit = func(_ string) (*git.Snapshot, []error) {
		return &git.Snapshot{IsRepo: true, RootDir: dir}, nil
	}
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, errWalkFailed }

	info := Detect()

	if info.Languages.Primary != "" {
		t.Errorf("expected empty Primary on walk error, got %q", info.Languages.Primary)
	}
}
