// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"errors"
	"log/slog"
	"testing"
)

func TestParseLevel_allValid(t *testing.T) {
	cases := []struct {
		input Level
		want  slog.Level
	}{
		{LevelDebug, slog.LevelDebug},
		{LevelInfo, slog.LevelInfo},
		{LevelWarn, slog.LevelWarn},
		{LevelWarning, slog.LevelWarn},
		{LevelError, slog.LevelError},
	}
	for _, tc := range cases {
		t.Run(string(tc.input), func(t *testing.T) {
			got, err := parseLevel(tc.input)
			if err != nil {
				t.Fatalf("parseLevel(%q) error = %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseLevel_invalid(t *testing.T) {
	_, err := parseLevel("verbose") // intentional invalid value
	if !errors.Is(err, ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got %v", err)
	}
}
