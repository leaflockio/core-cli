// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cmdconfig

import (
	"errors"
	"testing"
)

type testConfig struct {
	Name  string `mapstructure:"name"`
	Count int    `mapstructure:"count"`
}

var errTestConfigInvalid = errors.New("testConfig: invalid")

func (c *testConfig) Validate() error {
	if c.Name == "invalid" {
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
	if dest.Name != "leaf" || dest.Count != 3 {
		t.Errorf("dest = %+v, want {leaf 3}", dest)
	}
}

func TestCommandConfig_Load_nilSection_leavesDestUnchanged(t *testing.T) {
	dest := &testConfig{Name: "default", Count: 7}
	cc := CommandConfig[testConfig, *testConfig]{Dest: dest}

	err := cc.Load(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.Name != "default" || dest.Count != 7 {
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

func TestCommandConfig_Validate_ok(t *testing.T) {
	cc := CommandConfig[testConfig, *testConfig]{Dest: &testConfig{Name: "leaf"}}

	if err := cc.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommandConfig_Validate_forwardsDestError(t *testing.T) {
	cc := CommandConfig[testConfig, *testConfig]{Dest: &testConfig{Name: "invalid"}}

	err := cc.Validate()
	if !errors.Is(err, errTestConfigInvalid) {
		t.Errorf("Validate() = %v, want errTestConfigInvalid", err)
	}
}
