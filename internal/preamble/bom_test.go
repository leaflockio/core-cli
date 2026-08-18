// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package preamble

import (
	"bytes"
	"testing"
)

func TestHasBOM(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{name: "has BOM", content: append([]byte{0xEF, 0xBB, 0xBF}, "package main"...), want: true},
		{name: "no BOM", content: []byte("package main"), want: false},
		{name: "empty content", content: []byte{}, want: false},
		{name: "shorter than BOM", content: []byte{0xEF, 0xBB}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasBOM(tt.content); got != tt.want {
				t.Errorf("HasBOM(%v) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestStripBOM(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    []byte
	}{
		{
			name:    "strips a leading BOM",
			content: append([]byte{0xEF, 0xBB, 0xBF}, "package main"...),
			want:    []byte("package main"),
		},
		{
			name:    "no-op when absent",
			content: []byte("package main"),
			want:    []byte("package main"),
		},
		{
			name:    "empty content",
			content: []byte{},
			want:    []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripBOM(tt.content)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("StripBOM(%v) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}
