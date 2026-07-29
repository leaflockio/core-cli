// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package level

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name string
		l    Level
		want string
	}{
		{"root", LevelRoot, "root"},
		{"top", LevelTop, "top"},
		{"nested", LevelNested, "nested"},
		{"unknown", Level(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLevel_Next(t *testing.T) {
	tests := []struct {
		name string
		l    Level
		want Level
	}{
		{"root becomes top", LevelRoot, LevelTop},
		{"top becomes nested", LevelTop, LevelNested},
		{"nested stays nested", LevelNested, LevelNested},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.Next(); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
