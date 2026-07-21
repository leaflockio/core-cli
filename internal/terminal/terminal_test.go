// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package terminal

import (
	"bytes"
	"os"
	"testing"
)

func TestNew_setsFields(t *testing.T) {
	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	in := &bytes.Buffer{}

	term := New(out, errW, in)

	if term.Out != out {
		t.Error("Out not set to provided writer")
	}
	if term.Err != errW {
		t.Error("Err not set to provided writer")
	}
	if term.In != in {
		t.Error("In not set to provided reader")
	}
}

func TestNew_notTTYForBuffer(t *testing.T) {
	term := New(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	if term.IsTTY {
		t.Error("expected IsTTY=false when Out is a bytes.Buffer")
	}
}

func TestIsTTY_falseForNonFile(t *testing.T) {
	if isTTY(&bytes.Buffer{}) {
		t.Error("expected false for bytes.Buffer")
	}
}

func TestIsTTY_falseForRegularFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "terminal-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })

	if isTTY(f) {
		t.Error("expected false for a regular file — only a char device is a TTY")
	}
}

func TestIsTTY_falseWhenStatFails(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "terminal-test-*")
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}

	// Stat on a closed file returns an error — covers the err != nil branch.
	if isTTY(f) {
		t.Error("expected false when Stat fails on a closed file")
	}
}
