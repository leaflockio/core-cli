// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package workspace

import (
	"path/filepath"
	"testing"

	"github.com/leaflockio/core-cli/internal/config"
	"github.com/leaflockio/core-cli/internal/paths"
)

func TestKnown(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	ws, _ := New(home, repo)

	known := ws.Known()

	wantManifest := filepath.Join(repo, config.AppName, config.ManifestFile)
	if known.Manifest.Path != wantManifest {
		t.Errorf("Manifest.Path = %q, want %q", known.Manifest.Path, wantManifest)
	}
	if known.Manifest.Scope != paths.ScopeProject {
		t.Errorf("Manifest.Scope = %v, want ScopeProject", known.Manifest.Scope)
	}
	if known.Manifest.Generated {
		t.Error("Manifest.Generated should be false")
	}

	wantUserConfig := filepath.Join(home, config.EntityFolder, config.AppName, config.UserConfigFile)
	if known.UserConfig.Path != wantUserConfig {
		t.Errorf("UserConfig.Path = %q, want %q", known.UserConfig.Path, wantUserConfig)
	}
	if known.UserConfig.Scope != paths.ScopeUser {
		t.Errorf("UserConfig.Scope = %v, want ScopeUser", known.UserConfig.Scope)
	}

	if known.Credentials.Path != ws.CredentialsPath() {
		t.Errorf("Credentials.Path = %q, want %q", known.Credentials.Path, ws.CredentialsPath())
	}
	if known.Credentials.Scope != paths.ScopeUser {
		t.Errorf("Credentials.Scope = %v, want ScopeUser", known.Credentials.Scope)
	}
	if !known.Credentials.Generated {
		t.Error("Credentials.Generated should be true")
	}
}

func TestKnownPaths_All(t *testing.T) {
	ws, _ := New(t.TempDir(), t.TempDir())
	known := ws.Known()

	all, err := known.All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("All() length = %d, want 3", len(all))
	}
	if all[0] != known.Manifest || all[1] != known.UserConfig || all[2] != known.Credentials {
		t.Errorf("All() = %v, want [Manifest UserConfig Credentials] in order", all)
	}
}

func TestKnownPaths_All_duplicateName(t *testing.T) {
	kp := KnownPaths{
		Manifest:    paths.KnownPath{Name: "dup"},
		UserConfig:  paths.KnownPath{Name: "dup"},
		Credentials: paths.KnownPath{Name: "credentials"},
	}

	if _, err := kp.All(); err == nil {
		t.Fatal("expected error for duplicate Name, got nil")
	}
}
