// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package git wraps git subprocess invocations.
package git

import "os/exec"

// gitBinary is the executable name resolved on PATH for every invocation.
const gitBinary = "git"

// runOutput runs a git command in dir and returns its captured stdout.
var runOutput = func(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command(gitBinary, args...)
	cmd.Dir = dir
	return cmd.Output()
}

// lookPath resolves gitBinary on PATH.
var lookPath = exec.LookPath

// IsInstalled reports whether the git binary is available on PATH.
func IsInstalled() bool {
	_, err := lookPath(gitBinary)
	return err == nil
}
