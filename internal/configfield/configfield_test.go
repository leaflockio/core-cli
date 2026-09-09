// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package configfield

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testConfig struct {
	Name  Field[string]
	Count Field[int]
}

var errTestConfigInvalid = errors.New("testConfig: invalid")

func (c *testConfig) Validate() error {
	if c.Name.Value() == "invalid" {
		return errTestConfigInvalid
	}
	return nil
}

func TestCommandConfig_Load_decodesSection(t *testing.T) {
	dest := &testConfig{}
	cc := CommandConfig[testConfig, *testConfig]{Dest: dest}

	err := cc.Load(map[string]any{"name": "leaf", "count": 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Name.Value() != "leaf" || dest.Count.Value() != 3 {
		t.Errorf("dest = %+v, want {leaf 3}", dest)
	}
}

func TestCommandConfig_Load_nilSection_leavesDestUnchanged(t *testing.T) {
	dest := &testConfig{Name: Field[string]{Default: "default"}, Count: Field[int]{Default: 7}}
	cc := CommandConfig[testConfig, *testConfig]{Dest: dest}

	err := cc.Load(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Name.Value() != "default" || dest.Count.Value() != 7 {
		t.Errorf("dest = %+v, want unchanged {default 7}", dest)
	}
}

func TestCommandConfig_Load_decodeError(t *testing.T) {
	dest := &testConfig{}
	cc := CommandConfig[testConfig, *testConfig]{Dest: dest}

	// count expects an int; a string can't be decoded into it.
	err := cc.Load(map[string]any{"count": "not-a-number"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type unwrappedTestConfig struct {
	Name string
}

func (c *unwrappedTestConfig) Validate() error { return nil }

func TestCommandConfig_Load_rejectsUnwrappedDest(t *testing.T) {
	cc := CommandConfig[unwrappedTestConfig, *unwrappedTestConfig]{Dest: &unwrappedTestConfig{}}

	err := cc.Load(map[string]any{"name": "leaf"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCommandConfig_Validate_ok(t *testing.T) {
	cc := CommandConfig[testConfig, *testConfig]{Dest: &testConfig{Name: Field[string]{Default: "leaf"}}}

	if err := cc.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommandConfig_Validate_forwardsDestError(t *testing.T) {
	cc := CommandConfig[testConfig, *testConfig]{Dest: &testConfig{Name: Field[string]{Default: "invalid"}}}

	err := cc.Validate()
	if !errors.Is(err, errTestConfigInvalid) {
		t.Errorf("Validate() = %v, want errTestConfigInvalid", err)
	}
}

func TestCommandConfig_Save_writesFlattenedFile(t *testing.T) {
	dest := &testConfig{Name: Field[string]{Default: "leaf"}, Count: Field[int]{Default: 3}}
	cc := CommandConfig[testConfig, *testConfig]{Dest: dest}

	base := filepath.Join(t.TempDir(), "config")
	if err := cc.Save(base, 0o755, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(base + ".yaml")
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "name: leaf") || !strings.Contains(got, "count: 3") {
		t.Errorf("saved file = %q, want name/count present", got)
	}
}

func TestCommandConfig_Save_flattenErrorIsPropagated(t *testing.T) {
	cc := CommandConfig[unwrappedTestConfig, *unwrappedTestConfig]{Dest: &unwrappedTestConfig{}}

	err := cc.Save(filepath.Join(t.TempDir(), "config"), 0o755, 0o644)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
