// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import "testing"

func TestOS_String(t *testing.T) {
	cases := []struct {
		os       OS
		expected string
	}{
		{MacOS, "macos"},
		{Linux, "linux"},
		{Windows, "windows"},
		{UnknownOS, "unknown"},
	}

	for _, tc := range cases {
		if tc.os.String() != tc.expected {
			t.Errorf("OS(%q).String() = %q, want %q", tc.os, tc.os.String(), tc.expected)
		}
	}
}

func TestArch_String(t *testing.T) {
	cases := []struct {
		arch     Arch
		expected string
	}{
		{AMD64, "amd64"},
		{ARM64, "arm64"},
		{UnknownArch, "unknown"},
	}

	for _, tc := range cases {
		if tc.arch.String() != tc.expected {
			t.Errorf("Arch(%q).String() = %q, want %q", tc.arch, tc.arch.String(), tc.expected)
		}
	}
}

func TestPackageManager_String(t *testing.T) {
	cases := []struct {
		pm       PackageManager
		expected string
	}{
		{Homebrew, "homebrew"},
		{Apt, "apt"},
		{UnknownPM, "unknown"},
	}

	for _, tc := range cases {
		if tc.pm.String() != tc.expected {
			t.Errorf("PackageManager(%q).String() = %q, want %q", tc.pm, tc.pm.String(), tc.expected)
		}
	}
}
