// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	isolateConfig(t, EnvTest)

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Log.Level != defaultLogLevelDev {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, defaultLogLevelDev)
	}
	if cfg.Log.Format != defaultLogFormat {
		t.Errorf("Log.Format = %q, want %q", cfg.Log.Format, defaultLogFormat)
	}
}

func TestLoad_defaults_prod(t *testing.T) {
	isolateConfig(t, EnvProd)

	cfg, err := Load(EnvProd)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Log.Level != defaultLogLevelProd {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, defaultLogLevelProd)
	}
}

func TestLoad_baseConfig(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yaml", `
log:
  level: warn
  format: json
`)

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "warn")
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
}

func TestLoad_envOverlay(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yaml", `
log:
  level: warn
  format: text
`)
	writeYAML(t, dir, "config.test.yaml", `
log:
  level: debug
`)

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// overlay overrides base level but inherits base format
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want %q (overlay should win)", cfg.Log.Level, "debug")
	}
	if cfg.Log.Format != "text" {
		t.Errorf("Log.Format = %q, want %q (base should carry through)", cfg.Log.Format, "text")
	}
}

func TestLoad_envVarOverride(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yaml", `
log:
  level: warn
  format: text
`)
	t.Setenv("LEAF_LOG_LEVEL", "error")

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Log.Level != "error" {
		t.Errorf("Log.Level = %q, want %q (env var should win)", cfg.Log.Level, "error")
	}
}

func TestLoad_malformedBase(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yaml", `log: [invalid: yaml: {`)

	_, err := Load(EnvTest)
	if err == nil {
		t.Fatal("expected error for malformed config.yaml, got nil")
	}
}

func TestLoad_malformedOverlay(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yaml", `log:\n  level: debug`)
	writeYAML(t, dir, "config.test.yaml", `log: [invalid: yaml: {`)

	_, err := Load(EnvTest)
	if err == nil {
		t.Fatal("expected error for malformed config.test.yaml, got nil")
	}
}

func TestLoad_unmarshalError(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	// log is expected to be a struct but is given a scalar — unmarshal will fail
	writeYAML(t, dir, "config.yaml", `
log: 42
`)

	_, err := Load(EnvTest)
	if err == nil {
		t.Fatal("expected unmarshal error, got nil")
	}
}

func TestLoad_ymlExtension(t *testing.T) {
	dir := isolateConfig(t, EnvTest)
	writeYAML(t, dir, "config.yml", `
log:
  level: warn
`)

	cfg, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q (config.yml should be loaded)", cfg.Log.Level, "warn")
	}
}

func TestLoad_debugEnabled(t *testing.T) {
	isolateConfig(t, EnvTest)
	t.Setenv(envVarConfigDebug, "1")

	// exercises the fmt.Fprintf path inside debugf — output goes to stderr
	_, err := Load(EnvTest)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}
