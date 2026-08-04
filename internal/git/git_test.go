// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package git

import "testing"

func TestRunOutput_realImpl(t *testing.T) {
	out, err := runOutput("", "--version")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty output from git --version")
	}
}

func TestIsInstalled_trueWhenLookPathSucceeds(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(_ string) (string, error) { return "git", nil }

	if !IsInstalled() {
		t.Error("expected IsInstalled to be true when lookPath succeeds")
	}
}

func TestIsInstalled_falseWhenLookPathFails(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(_ string) (string, error) { return "", errCmdFailed }

	if IsInstalled() {
		t.Error("expected IsInstalled to be false when lookPath fails")
	}
}
