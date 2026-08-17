// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import (
	"errors"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

var errEmptyDelim = errors.New("comment: block style requires non-empty open and close delimiters")

// blockStyle is a start/end delimiter pair wrapping the whole block,
// with an optional per-interior-line marker.
type blockStyle struct {
	start, middle, end string
}

// NewBlockStyle constructs a Style that wraps lines between start and
// end, with middle prefixing each interior line.
func NewBlockStyle(start, middle, end string) (Style, error) {
	if start == "" || end == "" {
		return nil, errs.Caller(errs.CMT002,
			"block comment style requires non-empty open and close delimiters",
			errEmptyDelim,
			errs.Context{
				Cause:      "open or close was empty for a block comment style",
				Resolution: "provide non-empty open and close delimiters",
			},
		)
	}
	return blockStyle{start: start, middle: middle, end: end}, nil
}

func (s blockStyle) wrap(lines []string) []string {
	out := make([]string, 0, len(lines)+2)
	out = append(out, s.start)
	for _, l := range lines {
		if l == "" {
			out = append(out, strings.TrimRight(s.middle, " "))
		} else {
			out = append(out, s.middle+l)
		}
	}
	out = append(out, s.end)
	return out
}
