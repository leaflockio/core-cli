// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// consoleOnly returns a Config with only console output enabled.
func consoleOnly(level Level, format Format) *Config {
	return &Config{
		Level:   level,
		Console: ConsoleConfig{Enabled: true, Format: format},
	}
}

// bothOutputs returns a Config with console (text) and file (json) both enabled.
func bothOutputs(dir string) *Config {
	return &Config{
		Level:   LevelInfo,
		Console: ConsoleConfig{Enabled: true, Format: FormatText},
		File:    FileConfig{Enabled: true, Format: FormatJSON, Path: dir, Filename: "test.log"},
	}
}

func TestBuild_consoleText(t *testing.T) {
	var buf bytes.Buffer
	log, err := build(consoleOnly(LevelInfo, FormatText), &buf)
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	log.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected log output to contain %q, got %q", "hello", buf.String())
	}
}

func TestBuild_consoleJSON(t *testing.T) {
	var buf bytes.Buffer
	log, err := build(consoleOnly(LevelInfo, FormatJSON), &buf)
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	log.Info("hello")
	if !strings.Contains(buf.String(), `"msg"`) {
		t.Errorf("expected JSON log output, got %q", buf.String())
	}
}

func TestBuild_noOutputs(t *testing.T) {
	_, err := build(&Config{Level: LevelInfo}, io.Discard)
	if !errors.Is(err, ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got %v", err)
	}
}

func TestBuild_invalidLevel(t *testing.T) {
	_, err := build(consoleOnly("verbose", FormatText), io.Discard) // intentional invalid value
	if !errors.Is(err, ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got %v", err)
	}
}

func TestBuild_invalidConsoleFormat(t *testing.T) {
	_, err := build(consoleOnly(LevelInfo, "xml"), io.Discard) // intentional invalid value
	if !errors.Is(err, ErrUnknownFormat) {
		t.Errorf("expected ErrUnknownFormat, got %v", err)
	}
}

func TestBuild_invalidFileFormat(t *testing.T) {
	cfg := &Config{
		Level: LevelInfo,
		// intentional invalid format value
		File: FileConfig{Enabled: true, Format: "xml", Path: t.TempDir(), Filename: "test.log"},
	}
	_, err := build(cfg, io.Discard)
	if !errors.Is(err, ErrUnknownFormat) {
		t.Errorf("expected ErrUnknownFormat, got %v", err)
	}
}

func TestBuild_filePathEmpty(t *testing.T) {
	cfg := &Config{
		Level: LevelInfo,
		File:  FileConfig{Enabled: true, Format: FormatJSON, Path: "", Filename: "test.log"},
	}
	_, err := build(cfg, io.Discard)
	if !errors.Is(err, ErrFilePathEmpty) {
		t.Errorf("expected ErrFilePathEmpty, got %v", err)
	}
}

func TestBuild_bothOutputsDifferentFormats(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	log, err := build(bothOutputs(dir), &buf)
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}

	log.Info("both-outputs-test")

	consoleOut := buf.String()
	if !strings.Contains(consoleOut, "both-outputs-test") {
		t.Errorf("expected console output to contain %q, got %q", "both-outputs-test", consoleOut)
	}
	if strings.Contains(consoleOut, `"msg"`) {
		t.Errorf("expected text format on console, got JSON: %q", consoleOut)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.log"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), `"msg"`) {
		t.Errorf("expected JSON format in file, got: %q", string(data))
	}
}

func TestBuild_public(t *testing.T) {
	log, err := Build(consoleOnly(LevelInfo, FormatText))
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}
