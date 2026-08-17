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

func TestNewLineStyle_emptyPrefixIsCallerError(t *testing.T) {
	_, err := NewLineStyle("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errEmptyPrefix) {
		t.Errorf("error = %v, want wrapping errEmptyPrefix", err)
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.CMT001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.CMT001)
	}
}

func TestNewLineStyle_valid(t *testing.T) {
	s, err := NewLineStyle("#")
	if err != nil {
		t.Fatalf("NewLineStyle: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil Style")
	}
}

func TestLineStyle_wrap(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "single non-empty line",
			lines: []string{"hello"},
			want:  []string{"# hello"},
		},
		{
			name:  "multiple lines",
			lines: []string{"line one", "line two"},
			want:  []string{"# line one", "# line two"},
		},
		{
			name:  "empty line has no trailing space",
			lines: []string{""},
			want:  []string{"#"},
		},
		{
			name:  "mixed empty and non-empty lines",
			lines: []string{"a", "", "b"},
			want:  []string{"# a", "#", "# b"},
		},
		{
			name:  "no lines",
			lines: []string{},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := lineStyle{prefix: "#"}
			got := s.wrap(tt.lines)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrap(%v) = %v, want %v", tt.lines, got, tt.want)
			}
		})
	}
}
