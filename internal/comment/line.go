// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import (
	"errors"

	"github.com/leaflockio/core-cli/internal/errs"
)

var errEmptyPrefix = errors.New("comment: line style requires a non-empty prefix")

// lineStyle is a prefix repeated on every line.
type lineStyle struct {
	prefix string
}

// NewLineStyle constructs a Style whose marker is repeated on every line.
func NewLineStyle(prefix string) (Style, error) {
	if prefix == "" {
		return nil, errs.Caller(errs.CMT001,
			"line comment style requires a non-empty prefix",
			errEmptyPrefix,
			errs.Context{
				Cause:      "no prefix was given for a line comment style",
				Resolution: "provide a non-empty prefix",
			},
		)
	}
	return lineStyle{prefix: prefix}, nil
}

func (s lineStyle) wrap(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if l == "" {
			out = append(out, s.prefix)
		} else {
			out = append(out, s.prefix+" "+l)
		}
	}
	return out
}
