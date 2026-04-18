// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package strutil

import "testing"

func TestTitle(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"tools", "Tools"},
		{"general", "General"},
		{"a", "A"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := Title(tc.in); got != tc.want {
			t.Errorf("Title(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
