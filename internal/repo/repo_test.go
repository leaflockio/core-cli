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

	"github.com/leaflock/core-cli/internal/repo/lang"
)

var (
	errGitFailed   = errors.New("git failed")
	errWalkFailed  = errors.New("walk failed")
	errGetwdFailed = errors.New("getwd failed")
)

// --- Detect ---

func TestDetect_gitRepoWithRemote(t *testing.T) {
	dir := t.TempDir()
	origRoot := execGitRoot
	origRemote := execGitRemoteURL
	origWalk := walkFiles
	defer func() { execGitRoot = origRoot; execGitRemoteURL = origRemote; walkFiles = origWalk }()

	execGitRoot = func() ([]byte, error) { return []byte(dir), nil }
	execGitRemoteURL = func() ([]byte, error) {
		return []byte("https://github.com/leaflockio/core-cli.git"), nil
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
	if info.HasLicenseFile {
		t.Error("expected HasLicenseFile=false for empty dir")
	}
}

func TestDetect_gitRepoRemoteFails(t *testing.T) {
	dir := t.TempDir()
	origRoot := execGitRoot
	origRemote := execGitRemoteURL
	origWalk := walkFiles
	defer func() { execGitRoot = origRoot; execGitRemoteURL = origRemote; walkFiles = origWalk }()

	execGitRoot = func() ([]byte, error) { return []byte(dir), nil }
	execGitRemoteURL = func() ([]byte, error) { return nil, errGitFailed }
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
	origRoot := execGitRoot
	origWalk := walkFiles
	origGetwd := osGetwd
	defer func() { execGitRoot = origRoot; walkFiles = origWalk; osGetwd = origGetwd }()

	execGitRoot = func() ([]byte, error) { return nil, errGitFailed }
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
	origRoot := execGitRoot
	origWalk := walkFiles
	origGetwd := osGetwd
	defer func() { execGitRoot = origRoot; walkFiles = origWalk; osGetwd = origGetwd }()

	execGitRoot = func() ([]byte, error) { return nil, errGitFailed }
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
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT LICENSE"), 0o600); err != nil {
		t.Fatal(err)
	}
	origRoot := execGitRoot
	origRemote := execGitRemoteURL
	origWalk := walkFiles
	defer func() { execGitRoot = origRoot; execGitRemoteURL = origRemote; walkFiles = origWalk }()

	execGitRoot = func() ([]byte, error) { return []byte(dir), nil }
	execGitRemoteURL = func() ([]byte, error) { return nil, errGitFailed }
	walkFiles = func(_ string, _ bool) ([]string, error) { return nil, nil }

	info := Detect()

	if !info.HasLicenseFile {
		t.Error("expected HasLicenseFile=true")
	}
	if info.DetectedLicenseType != licenseMIT {
		t.Errorf("expected DetectedLicenseType=%q, got %q", licenseMIT, info.DetectedLicenseType)
	}
}

func TestDetect_withLanguages(t *testing.T) {
	dir := t.TempDir()
	origRoot := execGitRoot
	origRemote := execGitRemoteURL
	origWalk := walkFiles
	defer func() { execGitRoot = origRoot; execGitRemoteURL = origRemote; walkFiles = origWalk }()

	execGitRoot = func() ([]byte, error) { return []byte(dir), nil }
	execGitRemoteURL = func() ([]byte, error) { return nil, errGitFailed }
	walkFiles = func(_ string, _ bool) ([]string, error) {
		return []string{
			filepath.Join("repo", "main.go"),
			filepath.Join("repo", "util.go"),
			filepath.Join("repo", "app.ts"),
		}, nil
	}

	info := Detect()

	if info.Languages.Primary != lang.Go {
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
	exists, licenseType := detectLicense(dir)
	if exists || licenseType != "" {
		t.Errorf("expected no license, got exists=%v type=%q", exists, licenseType)
	}
}

func TestDetectLicense_licenseFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT LICENSE"), 0o600); err != nil {
		t.Fatal(err)
	}
	exists, licenseType := detectLicense(dir)
	if !exists || licenseType != licenseMIT {
		t.Errorf("expected MIT license, got exists=%v type=%q", exists, licenseType)
	}
}

func TestDetectLicense_licenseMdFile(t *testing.T) {
	dir := t.TempDir()
	// Only LICENSE.md exists, not LICENSE.
	if err := os.WriteFile(filepath.Join(dir, "LICENSE.md"), []byte("Apache License\n2.0"), 0o600); err != nil {
		t.Fatal(err)
	}
	exists, licenseType := detectLicense(dir)
	if !exists || licenseType != licenseApache20 {
		t.Errorf("expected apache-2.0, got exists=%v type=%q", exists, licenseType)
	}
}

// --- classifyLicenseText ---

func TestClassifyLicenseText_apache20(t *testing.T) {
	if got := classifyLicenseText("Apache License 2.0"); got != licenseApache20 {
		t.Errorf("expected %q, got %q", licenseApache20, got)
	}
}

func TestClassifyLicenseText_mit(t *testing.T) {
	if got := classifyLicenseText("MIT LICENSE"); got != licenseMIT {
		t.Errorf("expected %q, got %q", licenseMIT, got)
	}
}

func TestClassifyLicenseText_mitByPermission(t *testing.T) {
	if got := classifyLicenseText("Permission is hereby granted without warranty of any kind"); got != licenseMIT {
		t.Errorf("expected %q, got %q", licenseMIT, got)
	}
}

func TestClassifyLicenseText_gpl3(t *testing.T) {
	if got := classifyLicenseText("GNU General Public License Version 3"); got != licenseGPL3 {
		t.Errorf("expected %q, got %q", licenseGPL3, got)
	}
}

func TestClassifyLicenseText_gpl2(t *testing.T) {
	if got := classifyLicenseText("GNU General Public License Version 2"); got != licenseGPL2 {
		t.Errorf("expected %q, got %q", licenseGPL2, got)
	}
}

func TestClassifyLicenseText_lgpl3(t *testing.T) {
	if got := classifyLicenseText("GNU Lesser General Public License Version 3"); got != licenseLGPL3 {
		t.Errorf("expected %q, got %q", licenseLGPL3, got)
	}
}

func TestClassifyLicenseText_lgpl21(t *testing.T) {
	if got := classifyLicenseText("GNU Lesser General Public License Version 2"); got != licenseLGPL21 {
		t.Errorf("expected %q, got %q", licenseLGPL21, got)
	}
}

func TestClassifyLicenseText_agpl3(t *testing.T) {
	if got := classifyLicenseText("GNU Affero General Public License"); got != licenseAGPL3 {
		t.Errorf("expected %q, got %q", licenseAGPL3, got)
	}
}

func TestClassifyLicenseText_mpl2(t *testing.T) {
	if got := classifyLicenseText("Mozilla Public License"); got != licenseMPL2 {
		t.Errorf("expected %q, got %q", licenseMPL2, got)
	}
}

func TestClassifyLicenseText_bsd2Clause(t *testing.T) {
	if got := classifyLicenseText("BSD 2-Clause"); got != licenseBSD2Clause {
		t.Errorf("expected %q, got %q", licenseBSD2Clause, got)
	}
}

func TestClassifyLicenseText_bsd2ClauseTwoClause(t *testing.T) {
	if got := classifyLicenseText("BSD Two-Clause"); got != licenseBSD2Clause {
		t.Errorf("expected %q, got %q", licenseBSD2Clause, got)
	}
}

func TestClassifyLicenseText_bsd3Clause(t *testing.T) {
	if got := classifyLicenseText("BSD 3-Clause"); got != licenseBSD3Clause {
		t.Errorf("expected %q, got %q", licenseBSD3Clause, got)
	}
}

func TestClassifyLicenseText_bsd3ClauseThreeClause(t *testing.T) {
	if got := classifyLicenseText("BSD Three-Clause"); got != licenseBSD3Clause {
		t.Errorf("expected %q, got %q", licenseBSD3Clause, got)
	}
}

func TestClassifyLicenseText_isc(t *testing.T) {
	if got := classifyLicenseText("ISC License"); got != licenseISC {
		t.Errorf("expected %q, got %q", licenseISC, got)
	}
}

func TestClassifyLicenseText_proprietary(t *testing.T) {
	if got := classifyLicenseText("All rights reserved. Proprietary software."); got != licenseProprietary {
		t.Errorf("expected %q, got %q", licenseProprietary, got)
	}
}

// --- countLanguages ---

func TestCountLanguages_empty(t *testing.T) {
	counts := countLanguages(nil)
	if len(counts) != 0 {
		t.Errorf("expected empty counts, got %v", counts)
	}
}

func TestCountLanguages_knownExtensions(t *testing.T) {
	files := []string{
		filepath.Join("repo", "main.go"),
		filepath.Join("repo", "util.go"),
		filepath.Join("repo", "app.ts"),
	}
	counts := countLanguages(files)
	if counts[lang.Go] != 2 {
		t.Errorf("expected Go count=2, got %d", counts[lang.Go])
	}
	if counts[lang.TypeScript] != 1 {
		t.Errorf("expected TypeScript count=1, got %d", counts[lang.TypeScript])
	}
}

func TestCountLanguages_unknownExtensionIgnored(t *testing.T) {
	counts := countLanguages([]string{filepath.Join("repo", "Makefile"), filepath.Join("repo", "data.bin")})
	if len(counts) != 0 {
		t.Errorf("expected no counts for unknown extensions, got %v", counts)
	}
}

// --- buildDetection ---

func TestBuildDetection_empty(t *testing.T) {
	det := buildDetection(map[lang.Language]int{})
	if det.Primary != lang.Unknown {
		t.Errorf("expected Unknown primary for empty counts, got %v", det.Primary)
	}
	if len(det.All) != 0 {
		t.Errorf("expected empty All, got %v", det.All)
	}
}

func TestBuildDetection_singleLanguage(t *testing.T) {
	det := buildDetection(map[lang.Language]int{lang.Go: 3})
	if det.Primary != lang.Go {
		t.Errorf("expected Primary=Go, got %v", det.Primary)
	}
	if len(det.All) != 1 || det.All[0] != lang.Go {
		t.Errorf("expected All=[Go], got %v", det.All)
	}
}

func TestBuildDetection_orderedByCount(t *testing.T) {
	// Go: 3, TypeScript: 1, Python: 1 — tie between TypeScript and Python broken by Language value.
	counts := map[lang.Language]int{
		lang.Go:         3,
		lang.TypeScript: 1,
		lang.Python:     1,
	}
	det := buildDetection(counts)
	if det.Primary != lang.Go {
		t.Errorf("expected Primary=Go, got %v", det.Primary)
	}
	if len(det.All) != 3 {
		t.Errorf("expected 3 languages, got %v", det.All)
	}
	if det.All[0] != lang.Go {
		t.Errorf("expected Go first, got %v", det.All[0])
	}
}

// --- detectLanguages ---

func TestDetectLanguages_walkError(t *testing.T) {
	orig := walkFiles
	defer func() { walkFiles = orig }()
	walkFiles = func(_ string, _ bool) ([]string, error) {
		return nil, errWalkFailed
	}

	det := detectLanguages("root")
	if det.Primary != lang.Unknown {
		t.Errorf("expected Unknown primary on walk error, got %v", det.Primary)
	}
}

// --- execGitRoot / execGitRemoteURL real implementations ---

func TestExecGitRoot_realImpl(t *testing.T) {
	_, err := execGitRoot()
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
}

func TestExecGitRemoteURL_realImpl(t *testing.T) {
	_, err := execGitRemoteURL()
	if err != nil {
		t.Skipf("git remote get-url origin not available: %v", err)
	}
}
