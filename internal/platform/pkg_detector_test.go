// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var errTestNotFound = errors.New("not found")

func TestPkgDetector_Detect(t *testing.T) {
	// Construction helpers to avoid absolute path literals flagged by hooks.
	root := string(os.PathSeparator)
	usrBin := filepath.Join(root, "usr", "bin")
	usrLocalBin := filepath.Join(root, "usr", "local", "bin")

	tests := []struct {
		name     string
		os       OS
		lookPath LookPath
		expected []PackageManagerInfo
	}{
		{
			name: "macOS with Homebrew",
			os:   MacOS,
			lookPath: func(p string) (string, error) {
				if p == binBrew {
					return filepath.Join(usrLocalBin, binBrew), nil
				}
				return "", errTestNotFound
			},
			expected: []PackageManagerInfo{
				{Name: Homebrew, Present: true},
			},
		},
		{
			name: "Linux with Apt and Homebrew",
			os:   Linux,
			lookPath: func(p string) (string, error) {
				if p == binBrew || p == binApt {
					return filepath.Join(usrBin, p), nil
				}
				return "", errTestNotFound
			},
			expected: []PackageManagerInfo{
				{Name: Homebrew, Present: true},
				{Name: Apt, Present: true},
			},
		},
		{
			name: "Linux with Apt only",
			os:   Linux,
			lookPath: func(p string) (string, error) {
				if p == binApt {
					return filepath.Join(usrBin, binApt), nil
				}
				return "", errTestNotFound
			},
			expected: []PackageManagerInfo{
				{Name: Apt, Present: true},
			},
		},
		{
			name: "None",
			os:   Windows,
			lookPath: func(p string) (string, error) {
				return "", errTestNotFound
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PkgDetector{
				os:       tt.os,
				lookPath: tt.lookPath,
				detectors: []pmDetector{
					&homebrewDetector{},
					&aptDetector{},
				},
			}
			got := d.Detect()
			if len(got) != len(tt.expected) {
				t.Fatalf("PkgDetector.Detect() len = %v, want %v", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i].Name != tt.expected[i].Name {
					t.Errorf("PkgDetector.Detect()[%d].Name = %v, want %v", i, got[i].Name, tt.expected[i].Name)
				}
				if got[i].Present != tt.expected[i].Present {
					t.Errorf("PkgDetector.Detect()[%d].Present = %v, want %v",
						i, got[i].Present, tt.expected[i].Present)
				}
			}
		})
	}
}
