// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/store"
)

type fixture struct {
	Name  string `mapstructure:"name"`
	Count int    `mapstructure:"count"`
}

func TestLoad_notFound(t *testing.T) {
	dir := t.TempDir()
	var v fixture
	found, err := store.Load(dir, "config", &v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found should be false when no file exists")
	}
}

func TestSave_thenLoad_roundTrip(t *testing.T) {
	dir := t.TempDir()
	original := fixture{Name: "leaf", Count: 3}

	if err := store.Save(dir, "config", original, 0o755, 0o644); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var got fixture
	found, err := store.Load(dir, "config", &got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found {
		t.Fatal("found should be true after Save")
	}
	if got != original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestSave_writesDefaultExtension(t *testing.T) {
	dir := t.TempDir()
	if err := store.Save(dir, "config", fixture{Name: "leaf"}, 0o755, 0o644); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err != nil {
		t.Errorf("expected config.yaml to exist: %v", err)
	}
}

func TestSave_preservesExistingExtension(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"name": "old"}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if err := store.Save(dir, "config", fixture{Name: "new"}, 0o755, 0o644); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err == nil {
		t.Error("Save must not create config.yaml when config.json already exists")
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("expected config.json to still exist: %v", err)
	}
	if !strings.Contains(string(data), `"name":"new"`) {
		t.Errorf("expected config.json to be updated with new content, got: %s", data)
	}
}

func TestSave_createsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "path")
	if err := store.Save(dir, "config", fixture{Name: "leaf"}, 0o755, 0o644); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir to be created: %v", err)
	}
}

func TestLoad_json(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{"name": "leaf", "count": 9}`)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var got fixture
	found, err := store.Load(dir, "config", &got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found || got.Name != "leaf" || got.Count != 9 {
		t.Errorf("got %+v, found=%v, want {leaf 9}, found=true", got, found)
	}
}

func TestLoad_yml(t *testing.T) {
	dir := t.TempDir()
	data := []byte("name: leaf\ncount: 4\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var got fixture
	found, err := store.Load(dir, "config", &got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found || got.Name != "leaf" || got.Count != 4 {
		t.Errorf("got %+v, found=%v, want {leaf 4}, found=true", got, found)
	}
}

func TestLoad_jsonc(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{
  // a comment
  "name": "leaf",
  "count": 2,
}`)
	if err := os.WriteFile(filepath.Join(dir, "config.jsonc"), data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var got fixture
	found, err := store.Load(dir, "config", &got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found || got.Name != "leaf" || got.Count != 2 {
		t.Errorf("got %+v, found=%v, want {leaf 2}, found=true", got, found)
	}
}

func TestLoad_firstExtensionWins(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: yaml\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"name": "json"}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var got fixture
	found, err := store.Load(dir, "config", &got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !found || got.Name != "yaml" {
		t.Errorf("got %+v, found=%v, want name=yaml (yaml precedes json in the discoverable order)", got, found)
	}
}

func TestExistingExtensions_none(t *testing.T) {
	dir := t.TempDir()
	got := store.ExistingExtensions(dir, "config")
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestExistingExtensions_multiple(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: yaml\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"name": "json"}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got := store.ExistingExtensions(dir, "config")
	want := []string{"yaml", "json"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			break
		}
	}
}

func TestLoad_readError(t *testing.T) {
	dir := t.TempDir()
	// A directory named config.yaml causes os.ReadFile to fail with a
	// non-not-exist error, which Load must propagate rather than swallow.
	if err := os.Mkdir(filepath.Join(dir, "config.yaml"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var got fixture
	_, err := store.Load(dir, "config", &got)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
