// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package comment

import (
	"reflect"
	"testing"
)

func TestWrap_delegatesToStyle(t *testing.T) {
	line, err := NewLineStyle("#")
	if err != nil {
		t.Fatalf("NewLineStyle: %v", err)
	}
	got := Wrap([]string{"hello"}, line)
	want := []string{"# hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap(...) = %v, want %v", got, want)
	}

	block, err := NewBlockStyle("/*", " * ", " */")
	if err != nil {
		t.Fatalf("NewBlockStyle: %v", err)
	}
	got = Wrap([]string{"hello"}, block)
	want = []string{"/*", " * hello", " */"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap(...) = %v, want %v", got, want)
	}
}
