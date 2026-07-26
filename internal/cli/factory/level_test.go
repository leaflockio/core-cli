// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name string
		l    level
		want string
	}{
		{"root", levelRoot, "root"},
		{"top", levelTop, "top"},
		{"nested", levelNested, "nested"},
		{"unknown", level(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLevel_next(t *testing.T) {
	tests := []struct {
		name string
		l    level
		want level
	}{
		{"root becomes top", levelRoot, levelTop},
		{"top becomes nested", levelTop, levelNested},
		{"nested stays nested", levelNested, levelNested},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.next(); got != tt.want {
				t.Errorf("next() = %v, want %v", got, tt.want)
			}
		})
	}
}
