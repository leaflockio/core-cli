// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package terminal

import (
	"io"
	"os"
)

// Terminal abstracts all user-facing I/O. Commands write to Out and Err rather
// than directly to os.Stdout and os.Stderr, so tests can capture output via
// a bytes.Buffer without spawning a subprocess.
type Terminal struct {
	Out   io.Writer // normal output (stdout)
	Err   io.Writer // error output (stderr)
	In    io.Reader // user input for prompts (stdin)
	IsTTY bool      // false in CI or when output is piped; gates prompts and color
}

// New constructs a Terminal and detects whether out is a TTY.
func New(out, err io.Writer, in io.Reader) *Terminal {
	return &Terminal{
		Out:   out,
		Err:   err,
		In:    in,
		IsTTY: isTTY(out),
	}
}

// isTTY reports whether w is a real terminal.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := f.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
