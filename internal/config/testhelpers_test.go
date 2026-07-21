// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// isolateConfig points the loader at an empty temp dir so tests never
// accidentally read real config files on disk.
func isolateConfig(t *testing.T, env Env) string {
	t.Helper()
	dir := t.TempDir()
	switch env {
	case EnvDev, EnvTest:
		t.Setenv(envVarConfigDir, dir)
	case EnvProd:
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
	return dir
}

// captureStderr runs fn and returns everything written to os.Stderr during
// that call.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = old })

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
