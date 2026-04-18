// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileWriter_writes(t *testing.T) {
	dir := t.TempDir()
	w, err := newFileWriter(&FileConfig{Path: dir, Filename: "test.log"})
	if err != nil {
		t.Fatalf("newFileWriter() error = %v", err)
	}
	if _, err := fmt.Fprint(w, "test content"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "test.log"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "test content") {
		t.Errorf("expected file to contain %q, got %q", "test content", string(data))
	}
}

func TestNewFileWriter_emptyPath(t *testing.T) {
	_, err := newFileWriter(&FileConfig{Path: "", Filename: "test.log"})
	if !errors.Is(err, ErrFilePathEmpty) {
		t.Errorf("expected ErrFilePathEmpty, got %v", err)
	}
}
