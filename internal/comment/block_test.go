// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import (
	"errors"
	"reflect"
	"testing"

	"github.com/leaflockio/core-cli/internal/errs"
)

func TestNewBlockStyle_emptyOpenIsCallerError(t *testing.T) {
	_, err := NewBlockStyle("", " * ", "*/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errEmptyDelim) {
		t.Errorf("error = %v, want wrapping errEmptyDelim", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.CMT002 {
		t.Errorf("Code = %q, want %q", e.Code, errs.CMT002)
	}
}

func TestNewBlockStyle_emptyCloseIsCallerError(t *testing.T) {
	_, err := NewBlockStyle("/*", " * ", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errEmptyDelim) {
		t.Errorf("error = %v, want wrapping errEmptyDelim", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.CMT002 {
		t.Errorf("Code = %q, want %q", e.Code, errs.CMT002)
	}
}

func TestNewBlockStyle_valid(t *testing.T) {
	s, err := NewBlockStyle("/*", " * ", "*/")
	if err != nil {
		t.Fatalf("NewBlockStyle: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil Style")
	}
}

func TestBlockStyle_wrap(t *testing.T) {
	tests := []struct {
		name  string
		style blockStyle
		lines []string
		want  []string
	}{
		{
			name:  "starred block with content",
			style: blockStyle{start: "/*", middle: " * ", end: " */"},
			lines: []string{"hello", "world"},
			want:  []string{"/*", " * hello", " * world", " */"},
		},
		{
			name:  "empty line trims trailing space from middle",
			style: blockStyle{start: "/*", middle: " * ", end: " */"},
			lines: []string{""},
			want:  []string{"/*", " *", " */"},
		},
		{
			name:  "no lines still emits open and close",
			style: blockStyle{start: "/*", middle: " * ", end: " */"},
			lines: []string{},
			want:  []string{"/*", " */"},
		},
		{
			name:  "plain block with no middle marker",
			style: blockStyle{start: "<!--", middle: "", end: "-->"},
			lines: []string{"hello"},
			want:  []string{"<!--", "hello", "-->"},
		},
		{
			name:  "plain block, empty line with no middle marker",
			style: blockStyle{start: "<!--", middle: "", end: "-->"},
			lines: []string{""},
			want:  []string{"<!--", "", "-->"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.wrap(tt.lines)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrap(%v) = %v, want %v", tt.lines, got, tt.want)
			}
		})
	}
}
