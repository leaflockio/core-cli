// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/util/pathutil"
)

var (
	errExecFail    = errors.New("executable error")
	errHomeDirFail = errors.New("home dir error")
)

func TestConfigFileExtensions_yaml(t *testing.T) {
	got := configFileExtensions("yaml")
	if len(got) != 2 || got[0] != "yaml" || got[1] != "yml" {
		t.Errorf("got %v, want [yaml yml]", got)
	}
}

func TestConfigFileExtensions_yml(t *testing.T) {
	got := configFileExtensions("yml")
	if len(got) != 2 || got[0] != "yaml" || got[1] != "yml" {
		t.Errorf("got %v, want [yaml yml]", got)
	}
}

func TestConfigFileExtensions_other(t *testing.T) {
	for _, ft := range []string{"json", "toml", "ini"} {
		got := configFileExtensions(ft)
		if len(got) != 1 || got[0] != ft {
			t.Errorf("configFileExtensions(%q) = %v, want [%s]", ft, got, ft)
		}
	}
}

func TestFindConfigFile_found(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(""), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, ok := findConfigFile([]string{dir}, "config", []string{"yaml", "yml"})
	if !ok {
		t.Fatal("expected file to be found, got false")
	}
	want := filepath.Join(dir, "config.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindConfigFile_secondExtension(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(""), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, ok := findConfigFile([]string{dir}, "config", []string{"yaml", "yml"})
	if !ok {
		t.Fatal("expected file to be found via second extension, got false")
	}
	want := filepath.Join(dir, "config.yml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindConfigFile_notFound(t *testing.T) {
	dir := t.TempDir()

	_, ok := findConfigFile([]string{dir}, "config", []string{"yaml", "yml"})
	if ok {
		t.Error("expected file not to be found, got true")
	}
}

func TestFindConfigFile_secondDir(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	if err := os.WriteFile(filepath.Join(second, "config.yaml"), []byte(""), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, ok := findConfigFile([]string{first, second}, "config", []string{"yaml"})
	if !ok {
		t.Fatal("expected file to be found in second dir, got false")
	}
	want := filepath.Join(second, "config.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExecutableDir(t *testing.T) {
	dir := executableDir()
	if dir == "" {
		t.Error("executableDir() returned empty string")
	}
}

func TestExecutableDir_osError(t *testing.T) {
	old := osExecutable
	osExecutable = func() (string, error) { return "", errExecFail }
	t.Cleanup(func() { osExecutable = old })

	if dir := executableDir(); dir != "" {
		t.Errorf("expected empty string on error, got %q", dir)
	}
}

func TestXdgConfigDirs_homeDirError(t *testing.T) {
	t.Setenv(xdgConfigHomeEnvVar, "")
	old := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errHomeDirFail }
	t.Cleanup(func() { osUserHomeDir = old })

	if dirs := xdgConfigDirs(); len(dirs) != 0 {
		t.Errorf("expected empty dirs on home dir error, got %v", dirs)
	}
}

func TestDevConfigDirs_noOverride_found(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, pathutil.ModuleRootMarker), []byte("module test\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", pathutil.ModuleRootMarker, err)
	}
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	old := executableDir
	executableDir = func() string { return bin }
	t.Cleanup(func() { executableDir = old })
	t.Setenv(envVarConfigDir, "")

	dirs := devConfigDirs()
	if len(dirs) == 0 {
		t.Fatal("expected at least one config dir, got none")
	}
	want := filepath.Join(root, configDirName)
	if dirs[0] != want {
		t.Errorf("got %q, want %q", dirs[0], want)
	}
}

func TestDevConfigDirs_noOverride_notFound(t *testing.T) {
	old := executableDir
	executableDir = func() string { return "" }
	t.Cleanup(func() { executableDir = old })
	t.Setenv(envVarConfigDir, "")

	dirs := devConfigDirs()
	if len(dirs) != 1 || dirs[0] != configDirName {
		t.Errorf("got %v, want [%q]", dirs, configDirName)
	}
}

func TestPaths_leafConfigDirOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(envVarConfigDir, dir)
	writeYAML(t, dir, "config.yaml", `
log:
  level: warn
`)

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q (%s not used)", cfg.Log.Level, "warn", envVarConfigDir)
	}
}

func TestPaths_xdgConfigHome(t *testing.T) {
	xdgBase := t.TempDir()
	leafDir := filepath.Join(xdgBase, appFolder)
	if err := os.MkdirAll(leafDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Setenv(xdgConfigHomeEnvVar, xdgBase)
	writeYAML(t, leafDir, "config.yaml", `
log:
  level: error
`)

	cfg, err := Load(EnvProd)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Level != "error" {
		t.Errorf("Log.Level = %q, want %q (%s not used)", cfg.Log.Level, "error", xdgConfigHomeEnvVar)
	}
}

func TestPaths_xdgConfigHome_default(t *testing.T) {
	t.Setenv(xdgConfigHomeEnvVar, "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	leafDir := filepath.Join(home, xdgDefaultConfigDir, appFolder)
	marker := filepath.Join(leafDir, "config.yaml")
	if _, err := os.Stat(marker); err == nil {
		t.Skipf("real %s/config.yaml exists, skipping to avoid interference", leafDir)
	}

	if err := os.MkdirAll(leafDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { os.Remove(marker) })
	writeYAML(t, leafDir, "config.yaml", `
log:
  level: warn
`)

	cfg, err := Load(EnvProd)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q (%s not used)", cfg.Log.Level, "warn", leafDir)
	}
}
